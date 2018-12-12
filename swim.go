/*
 * Copyright 2018 De-labtory
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package swim

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"sync/atomic"

	"github.com/DE-labtory/swim/pb"
	"github.com/it-chain/iLogger"
	"github.com/rs/xid"
)

var ErrInvalidMbrStatsMsgType = errors.New("error invalid mbrStatsMsg type")
var ErrPingFailed = errors.New("error ping failed")
var ErrIndProbeFailed = errors.New("error when indirect-probe failed")
var ErrInvalidPayloadType = errors.New("invalid indirect-ping response payload type")

type ProbeResponse struct {
	err error
}

func (r ProbeResponse) Ok() bool {
	return r.err == nil
}

type IndProbeResponse struct {
	err error
	msg pb.Message
}

func (r IndProbeResponse) Ok() bool {
	return r.err == nil && reflect.TypeOf(r.msg.Payload) == reflect.TypeOf(&pb.Message_Ack{})
}

type Config struct {

	// The maximum number of times the same piggyback data can be queried
	MaxlocalCount int

	// The maximum number of node-self-awareness counter
	MaxNsaCounter int

	// T is the the period of the probe
	T int

	// Timeout of ack after ping to a member
	AckTimeOut int

	// K is the number of members to send indirect ping
	K int

	// my address and port
	BindAddress string
	BindPort    int
}

type SWIM struct {
	// Swim Config
	config *Config

	// Currently connected memberList
	memberMap *MemberMap

	// Awareness manages health of the local node.
	awareness *Awareness

	// messageEndpoint work both as message transmitter and message receiver
	messageEndpoint MessageEndpoint

	// Information of this node
	member *Member

	// stopFlag tells whether SWIM is running or not
	stopFlag int32

	// FailureDetector quit channel
	quitFD chan struct{}

	// MbrStatsMsgStore which store messages about recent state changes of member.
	mbrStatsMsgStore MbrStatsMsgStore
}

func New(config *Config, suspicionConfig *SuspicionConfig, messageEndpointConfig MessageEndpointConfig, member *Member) *SWIM {
	if config.T < config.AckTimeOut {
		iLogger.Panic(nil, "T time must be longer than ack time-out")
	}

	swim := SWIM{
		config:    config,
		awareness: NewAwareness(config.MaxNsaCounter),
		memberMap: NewMemberMap(suspicionConfig),
		member:    member,
		quitFD:    make(chan struct{}),
	}

	messageEndpoint := messageEndpointFactory(config, messageEndpointConfig, &swim)
	swim.messageEndpoint = messageEndpoint

	return &swim
}

func messageEndpointFactory(config *Config, messageEndpointConfig MessageEndpointConfig, messageHandler MessageHandler) MessageEndpoint {
	packetTransportConfig := PacketTransportConfig{
		BindAddress: config.BindAddress,
		BindPort:    config.BindPort,
	}

	packetTransport, err := NewPacketTransport(&packetTransportConfig)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	messageEndpoint, err := NewMessageEndpoint(messageEndpointConfig, packetTransport, messageHandler)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	return messageEndpoint
}

// Start SWIM protocol.
func (s *SWIM) Start() {
	atomic.CompareAndSwapInt32(&s.stopFlag, DIE, AVAILABLE)
	go s.messageEndpoint.Listen()
	s.startFailureDetector()
}

// Dial to the all peerAddresses and exchange memberList.
func (s *SWIM) Join(peerAddresses []string) error {

	for _, address := range peerAddresses {
		s.exchangeMembership(address)
	}

	return nil
}

func (s *SWIM) exchangeMembership(address string) error {

	// Create membership message
	membership := s.createMembership()

	// Exchange membership
	msg, err := s.messageEndpoint.SyncSend(address, pb.Message{
		Address: s.member.Address(),
		Id:      xid.New().String(),
		Payload: &pb.Message_Membership{
			Membership: membership,
		},
	})

	if err != nil {
		return err
	}

	// Handle received membership
	switch msg := msg.Payload.(type) {
	case *pb.Message_Membership:
		for _, m := range msg.Membership.MbrStatsMsgs {
			s.handleMbrStatsMsg(m)
		}
	default:
		return errors.New("invaild response message")
	}

	return nil
}

func (s *SWIM) createMembership() *pb.Membership {

	membership := &pb.Membership{
		MbrStatsMsgs: make([]*pb.MbrStatsMsg, 0),
	}
	for _, m := range s.memberMap.GetMembers() {
		membership.MbrStatsMsgs = append(membership.MbrStatsMsgs, &pb.MbrStatsMsg{
			Incarnation: m.Incarnation,
			Id:          m.GetIDString(),
			Type:        pb.MbrStatsMsg_Type(m.Status.toInt()),
			Address:     m.Address(),
		})
	}

	return membership
}

func (s *SWIM) handleMbrStatsMsg(mbrStatsMsg *pb.MbrStatsMsg) {
	msgHandleErr := error(nil)
	hasChanged := false

	if s.member.ID.ID == mbrStatsMsg.Id {
		s.refute(mbrStatsMsg)
		s.mbrStatsMsgStore.Push(*mbrStatsMsg)
		return
	}

	switch mbrStatsMsg.Type {
	case pb.MbrStatsMsg_Alive:
		hasChanged, msgHandleErr = s.handleAliveMbrStatsMsg(mbrStatsMsg)
	case pb.MbrStatsMsg_Suspect:
		hasChanged, msgHandleErr = s.handleSuspectMbrStatsMsg(mbrStatsMsg)
	default:
		msgHandleErr = ErrInvalidMbrStatsMsgType
	}

	if msgHandleErr != nil {
		iLogger.Errorf(nil, "error occured when handling mbrStatsMsg [%s], error [%s]", mbrStatsMsg, msgHandleErr)
		return
	}

	// Push piggyback when status of membermap has updated.
	// If the content of the piggyback is about a new state change,
	// it must propagate to inform the network of the new state change.
	if hasChanged {
		s.mbrStatsMsgStore.Push(*mbrStatsMsg)
	}
}

func (s *SWIM) handleAliveMbrStatsMsg(stats *pb.MbrStatsMsg) (bool, error) {
	if stats.Type != pb.MbrStatsMsg_Alive {
		return false, ErrInvalidMbrStatsMsgType
	}

	msg, err := s.convMbrStatsToAliveMsg(stats)
	if err != nil {
		return false, err
	}

	return s.memberMap.Alive(msg)
}

func (s *SWIM) handleSuspectMbrStatsMsg(stats *pb.MbrStatsMsg) (bool, error) {
	if stats.Type != pb.MbrStatsMsg_Suspect {
		return false, ErrInvalidMbrStatsMsgType
	}

	msg, err := s.convMbrStatsToSuspectMsg(stats)
	if err != nil {
		return false, err
	}

	return s.memberMap.Suspect(msg)
}

func (s *SWIM) convMbrStatsToAliveMsg(stats *pb.MbrStatsMsg) (AliveMessage, error) {
	if stats.Type != pb.MbrStatsMsg_Alive {
		return AliveMessage{}, ErrInvalidMbrStatsMsgType
	}

	host, port, err := ParseHostPort(stats.Address)
	if err != nil {
		return AliveMessage{}, err
	}
	return AliveMessage{
		MemberMessage: MemberMessage{
			ID:          stats.Id,
			Addr:        host,
			Port:        port,
			Incarnation: stats.Incarnation,
		},
	}, nil
}

func (s *SWIM) convMbrStatsToSuspectMsg(stats *pb.MbrStatsMsg) (SuspectMessage, error) {
	if stats.Type != pb.MbrStatsMsg_Suspect {
		return SuspectMessage{}, ErrInvalidMbrStatsMsgType
	}

	host, port, err := ParseHostPort(stats.Address)
	if err != nil {
		return SuspectMessage{}, err
	}
	return SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          stats.Id,
			Addr:        host,
			Port:        port,
			Incarnation: stats.Incarnation,
		},
		ConfirmerID: s.member.ID.ID,
	}, nil
}

func (s *SWIM) refute(mbrStatsMsg *pb.MbrStatsMsg) {

	accusedInc := mbrStatsMsg.Incarnation
	inc := atomic.AddUint32(&s.member.Incarnation, 1)
	if s.member.Incarnation >= accusedInc {
		inc = atomic.AddUint32(&s.member.Incarnation, accusedInc-inc+1)
	}
	s.member.Incarnation = inc

	// Update piggyBack's incarnation to store to pbkStore
	mbrStatsMsg.Incarnation = inc

	// Increase awareness count(Decrease our health) because we are being asked to refute a problem.
	s.awareness.ApplyDelta(1)
}

// Gossip message to p2p network.
func (s *SWIM) Gossip(msg []byte) {

}

// Shutdown the running swim.
func (s *SWIM) ShutDown() {
	atomic.CompareAndSwapInt32(&s.stopFlag, AVAILABLE, DIE)
	s.messageEndpoint.Shutdown()
	s.quitFD <- struct{}{}
}

// toDie tell whether SWIM is stopped or not
func (s *SWIM) toDie() bool {
	return atomic.LoadInt32(&(s.stopFlag)) == DIE
}

// Total Failure Detection is performed for each` T`. (ref: https://github.com/DE-labtory/swim/edit/develop/docs/Docs.md)
//
// 1. SWIM randomly selects a member(j) in the memberMap and ping to the member(j).
//
// 2. SWIM waits for ack of the member(j) during the ack-timeout (time less than T).
//    End failure Detector if ack message arrives on ack-timeout.
//
// 3. SWIM selects k number of members from the memberMap and sends indirect-ping(request k members to ping the member(j)).
//    The nodes (that receive the indirect-ping) ping to the member(j) and ack when they receive ack from the member(j).
//
// 4. At the end of T, SWIM checks to see if ack was received from k members, and if there is no message,
//    The member(j) is judged to be failed, so check the member(j) as suspected or delete the member(j) from memberMap.
//
// ** When performing ping, ack, and indirect-ping in the above procedure, piggybackdata is sent together. **
//
//
// startFailureDetector function
//
// 1. Pick a member from memberMap.
// 2. Probe the member.
// 3. After finishing probing all members, Reset memberMap
//
func (s *SWIM) startFailureDetector() {
	go func() {
		interval := time.Millisecond * time.Duration(s.config.T)
		T := time.NewTimer(interval)

		for !s.toDie() {
			members := s.memberMap.GetMembers()
			if len(members) == 0 {
				<-T.C
				T.Reset(interval)
				continue
			}

			for _, m := range members {
				iLogger.Info(nil, "[swim] start probing ...")
				s.probe(m, T)

				iLogger.Info(nil, "[swim] done probing !")
				T.Reset(interval)
			}

			// Reset memberMap.
			s.memberMap.Reset()
		}
	}()

	<-s.quitFD
}

// probe function
//
// 1. Send ping to the member(j) during the ack-timeout (time less than T).
//    Return if ack message arrives on ack-timeout.
//
// 2. selects k number of members from the memberMap and sends indirect-ping(request k members to ping the member(j)).
//    The nodes (that receive the indirect-ping) ping to the member(j) and ack when they receive ack from the member(j).
//
// 3. At the end of T, SWIM checks to see if ack was received from k members, and if there is no message,
//    The member(j) is judged to be failed, so check the member(j) as suspected or delete the member(j) from memberMap.
//
func (s *SWIM) probe(member Member, timer *time.Timer) {

	if member.Status == Dead {
		return
	}

	end := make(chan ProbeResponse, 1)
	defer func() {
		close(end)
	}()

	// this context and wait group are used to cancel probe procedure
	// when probing time is out
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		task := func() (interface{}, error) {
			defer wg.Done()

			err := s.ping(&member)
			if err == nil {
				return nil, nil
			}

			if err != ErrSendTimeout {
				return nil, ErrPingFailed
			}

			err = s.indirectProbe(&member)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}
		resp := NewTaskRunner(task, ctx).Start()
		end <- ProbeResponse{err: resp.err}
	}()

	select {
	// if timed-out, then suspect member
	case <-timer.C:
		iLogger.Infof(nil, "[SWIM] probe member [%s] timed out, start suspect", member.ID)
		// when probing time is out, then cancel the probing procedure
		cancel()
		wg.Wait()

		s.awareness.ApplyDelta(1)
		s.suspect(&member)

	// if probe ended with error then suspect member and increase Awareness
	// otherwise just decrease Awareness score
	case resp := <-end:
		if !resp.Ok() {
			fmt.Println("not ok")
			s.awareness.ApplyDelta(1)
			s.suspect(&member)
			return
		}

		s.awareness.ApplyDelta(-1)
	}
}

// indirectProbe select k-random member from MemberMap, sends
// indirect-ping to them. if one of them sends back Ack message
// then indirectProbe success, otherwise failed.
// if one of k-member successfully received ACK message, then cancel
// k-1 member's probe
func (s *SWIM) indirectProbe(target *Member) error {
	wg := &sync.WaitGroup{}
	wg.Add(s.config.K)

	// with cancel we can send the signal to goroutines which share
	// this @ctx context
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan IndProbeResponse, s.config.K)

	defer func() {
		cancel()
		close(done)
	}()

	kMembers := s.memberMap.SelectKRandomMemberID(s.config.K)
	for _, m := range kMembers {
		go func(mediator Member) {
			defer wg.Done()

			task := func() (interface{}, error) {
				return s.indirectPing(mediator, *target)
			}

			resp := NewTaskRunner(task, ctx).Start()
			if resp.err != nil {
				done <- IndProbeResponse{
					err: resp.err,
					msg: pb.Message{},
				}
				return
			}

			msg, ok := resp.payload.(pb.Message)
			if !ok {
				done <- IndProbeResponse{
					err: ErrInvalidPayloadType,
					msg: pb.Message{},
				}
				return
			}
			done <- IndProbeResponse{
				err: nil,
				msg: msg,
			}
		}(m)
	}

	// wait until k-random member sends back response, if response message
	// is Ack message, then indirectProbe success because one of k-member
	// success UDP communication, if Nack message or Invalid message, increase
	// @unexpectedRespCounter then wait other member's response

	unexpectedResp := make([]IndProbeResponse, 0)

	for {
		resp := <-done

		if !resp.Ok() {
			unexpectedResp = append(unexpectedResp, resp)
			if len(unexpectedResp) >= s.config.K {
				iLogger.Infof(nil, "unexpected responses [%v]", unexpectedResp)
				return ErrIndProbeFailed
			}
			continue
		}

		cancel()
		wg.Wait()
		return nil
	}
}

// ping ping to member with piggyback message after sending ping message
// the result can be:
// 1. timeout
//    in this case, push signal to start indirect-ping request to k random nodes
// 2. successfully probed
//    in the case of successfully probe target node, update member state with
//    piggyback message sent from target member.
func (s *SWIM) ping(target *Member) error {
	stats, err := s.mbrStatsMsgStore.Get()
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	// send ping message
	addr := target.Address()
	pingId := xid.New().String()
	ping := createPingMessage(pingId, s.member.Address(), &stats)

	res, err := s.messageEndpoint.SyncSend(addr, ping)
	if err != nil {
		return err
	}

	// update piggyback data to store
	s.handlePbk(res.PiggyBack)

	return nil
}

// indirectPing sends indirect-ping to @member targeting @target member
// ** only when @member sends back to local node, push Message to channel
// otherwise just return **
// @ctx is for sending cancel signal from outside, when one of k-member successfully
// received ACK message or when in the exceptional situation
func (s *SWIM) indirectPing(mediator, target Member) (pb.Message, error) {
	stats, err := s.mbrStatsMsgStore.Get()
	if err != nil {
		return pb.Message{}, err
	}

	// send indirect-ping message
	addr := mediator.Address()
	indId := xid.New().String()
	ind := createIndMessage(indId, s.member.Address(), target.Address(), &stats)

	res, err := s.messageEndpoint.SyncSend(addr, ind)

	// when communicating member and target with indirect-ping failed,
	// just return.
	if err != nil {
		return pb.Message{}, err
	}
	// update piggyback data to store
	s.handlePbk(res.PiggyBack)

	return res, nil
}

func (s *SWIM) suspect(member *Member) {
	msg := createSuspectMessage(member, s.member.ID.ID)

	result, err := s.memberMap.Suspect(msg)
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	iLogger.Infof(nil, "Result of suspect [%t]", result)
}

// handler interface to handle received message
type MessageHandler interface {
	handle(msg pb.Message)
}

// The handle function does two things.
//
// 1. Update the member map using the piggyback-data contained in the message.
//  1-1. Check if member is me or not.
// 	1-2. Change status of member.
// 	1-3. If the state of the member map changes, store new status in the piggyback store.
//
// 2. Process Ping, Ack, Indirect-ping messages.
//
func (s *SWIM) handle(msg pb.Message) {

	s.handlePbk(msg.PiggyBack)

	switch p := msg.Payload.(type) {
	case *pb.Message_Ping:
		s.handlePing(msg)
	case *pb.Message_Ack:
		// handle ack
	case *pb.Message_IndirectPing:
		s.handleIndirectPing(msg)
	case *pb.Message_Membership:
		s.handleMembership(p.Membership, msg.Address)
	default:

	}
}

// handle piggyback related to member status
func (s *SWIM) handlePbk(piggyBack *pb.PiggyBack) {
	mbrStatsMsg := piggyBack.MbrStatsMsg

	// Check if piggyback message changes memberMap.
	s.handleMbrStatsMsg(mbrStatsMsg)
}

// handlePing send back Ack message by response
func (s *SWIM) handlePing(msg pb.Message) {
	id := msg.Id

	mbrStatsMsg, err := s.mbrStatsMsgStore.Get()
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	// address of messgae source member
	srcAddr := msg.Address

	ack := createAckMessage(id, srcAddr, &mbrStatsMsg)
	if err := s.messageEndpoint.Send(srcAddr, ack); err != nil {
		iLogger.Error(nil, err.Error())
	}
}

// handleIndirectPing receives IndirectPing message, so send the ping message
// to target member, if successfully receives ACK message then send back again
// ACK message to member who sent IndirectPing message to me.
// If ping was not successful then send back NACK message
func (s *SWIM) handleIndirectPing(msg pb.Message) {
	id := msg.Id

	// retrieve piggyback data from pbkStore
	mbrStatsMsg, err := s.mbrStatsMsgStore.Get()
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	// address of message source member
	srcAddr := msg.Address

	// address of indirect-ping's target
	targetAddr := msg.Payload.(*pb.Message_IndirectPing).IndirectPing.Target

	pingId := xid.New().String()
	ping := createPingMessage(pingId, s.member.Address(), &mbrStatsMsg)

	// first send the ping to target member, if target member could not send-back
	// ack message for whatever reason send nack message to source member,
	// if successfully received ack message from target, then send back ack message
	// to source member
	if _, err := s.messageEndpoint.SyncSend(targetAddr, ping); err != nil {
		nack := createNackMessage(id, srcAddr, &mbrStatsMsg)
		if err := s.messageEndpoint.Send(srcAddr, nack); err != nil {
			iLogger.Error(nil, err.Error())
		}
		return
	}

	ack := createAckMessage(id, srcAddr, &mbrStatsMsg)
	if err := s.messageEndpoint.Send(srcAddr, ack); err != nil {
		iLogger.Error(nil, err.Error())
	}
}

// handleMembership receives Membership message.
// create membership message with membermap and reply to the message with it.
func (s *SWIM) handleMembership(membership *pb.Membership, address string) error {
	// Create membership message
	m := s.createMembership()

	// Reply
	err := s.messageEndpoint.Send(address, pb.Message{
		Address: s.member.Address(),
		Id:      xid.New().String(),
		Payload: &pb.Message_Membership{
			Membership: m,
		},
	})

	if err != nil {
		return err
	}

	for _, m := range membership.MbrStatsMsgs {
		s.handleMbrStatsMsg(m)
	}

	return nil
}

func createPingMessage(id, src string, mbrStatsMsg *pb.MbrStatsMsg) pb.Message {
	return pb.Message{
		Id:      id,
		Address: src,
		Payload: &pb.Message_Ping{
			Ping: &pb.Ping{},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: mbrStatsMsg,
		},
	}
}

func createIndMessage(id, src, target string, mbrStatsMsg *pb.MbrStatsMsg) pb.Message {
	return pb.Message{
		Id:      id,
		Address: src,
		Payload: &pb.Message_IndirectPing{
			IndirectPing: &pb.IndirectPing{
				Target: target,
			},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: mbrStatsMsg,
		},
	}
}

func createNackMessage(id, src string, mbrStatsMsg *pb.MbrStatsMsg) pb.Message {
	return pb.Message{
		Id:      id,
		Address: src,
		Payload: &pb.Message_Nack{
			Nack: &pb.Nack{},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: mbrStatsMsg,
		},
	}
}

func createAckMessage(id, src string, mbrStatsMsg *pb.MbrStatsMsg) pb.Message {
	return pb.Message{
		Id:      id,
		Address: src,
		Payload: &pb.Message_Ack{
			Ack: &pb.Ack{},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: mbrStatsMsg,
		},
	}
}

func createSuspectMessage(suspect *Member, confirmer string) SuspectMessage {
	return SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          suspect.ID.ID,
			Addr:        suspect.Addr,
			Port:        suspect.Port,
			Incarnation: suspect.Incarnation,
		},
		ConfirmerID: confirmer,
	}
}
