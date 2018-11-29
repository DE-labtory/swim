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
	"time"

	"github.com/it-chain/iLogger"
	"github.com/rs/xid"

	"github.com/DE-labtory/swim/pb"
)

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

	// FailureDetector quit channel
	quitFD chan struct{}

	// Piggyback-store which store messages about recent state changes of member.
	pbkStore PBkStore
}

func New(config *Config, suspicionConfig *SuspicionConfig, messageEndpointConfig MessageEndpointConfig) *SWIM {
	if config.T < config.AckTimeOut {
		iLogger.Panic(nil, "T time must be longer than ack time-out")
	}

	swim := SWIM{
		config:    config,
		awareness: NewAwareness(config.MaxNsaCounter),
		memberMap: NewMemberMap(suspicionConfig),
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

}

// Dial to the all peerAddresses and exchange memberList.
func (s *SWIM) Join(peerAddresses []string) error {
	return nil
}

// Gossip message to p2p network.
func (s *SWIM) Gossip(msg []byte) {

}

// Shutdown the running swim.
func (s *SWIM) ShutDown() {
	s.quitFD <- struct{}{}
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
		for {
			// Get copy of current members from memberMap.
			members := s.memberMap.GetMembers()
			for _, member := range members {
				s.probe(member)
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

func (s *SWIM) probe(member Member) {

	if member.Status == Dead {
		return
	}

	end := make(chan struct{}, 1)
	defer close(end)

	go func() {

		// Ping to member
		time.Sleep(1 * time.Second)
		end <- struct{}{}
	}()

	T := time.NewTimer(time.Millisecond * time.Duration(s.config.T))

	select {
	case <-end:
		// Ended
		return
	case <-T.C:
		// Suspect the member.
		return
	}
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

	switch msg.Payload.(type) {
	case *pb.Message_Ping:
		s.handlePing(msg)
	case *pb.Message_Ack:
		// handle ack
	case *pb.Message_IndirectPing:
		s.handleIndirectPing(msg)
	default:

	}
}

// handle piggyback related to member status
func (s *SWIM) handlePbk(piggyBack *pb.PiggyBack) {
	// Check if piggyback message changes memberMap.
	hasChanged := false

	switch piggyBack.Type {
	case pb.PiggyBack_Alive:
		// Call Alive function in memberMap.
	case pb.PiggyBack_Confirm:
		// Call Confirm function in memberMap.
	case pb.PiggyBack_Suspect:
		// Call Suspect function in memberMap.
	default:
		// PiggyBack_type error
	}

	// Push piggyback when status of membermap has updated.
	// If the content of the piggyback is about a new state change,
	// it must propagate to inform the network of the new state change.
	if hasChanged {
		s.pbkStore.Push(*piggyBack)
	}
}

// handlePing send back Ack message by response
func (s *SWIM) handlePing(msg pb.Message) {
	id := msg.Id

	pbk, err := s.pbkStore.Get()
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	// address of messgae source member
	scrAddr := msg.Address

	ack := createAckMessage(id, &pbk)
	if err := s.messageEndpoint.Send(scrAddr, ack); err != nil {
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
	pbk, err := s.pbkStore.Get()
	if err != nil {
		iLogger.Error(nil, err.Error())
	}

	// address of message source member
	srcAddr := msg.Address

	// address of indirect-ping's target
	targetAddr := msg.Payload.(*pb.Message_IndirectPing).IndirectPing.Target

	ping := createPingMessage(&pbk)

	// first send the ping to target member, if target member could not send-back
	// ack message for whatever reason send nack message to source member,
	// if successfully received ack message from target, then send back ack message
	// to source member

	if _, err := s.messageEndpoint.SyncSend(targetAddr, ping); err != nil {
		nack := createNackMessage(id, &pbk)
		if err := s.messageEndpoint.Send(srcAddr, nack); err != nil {
			iLogger.Error(nil, err.Error())
		}
		return
	}

	ack := createAckMessage(id, &pbk)
	if err := s.messageEndpoint.Send(srcAddr, ack); err != nil {
		iLogger.Error(nil, err.Error())
	}
}

func createPingMessage(pbk *pb.PiggyBack) pb.Message {
	return pb.Message{
		Id: xid.New().String(),
		Payload: &pb.Message_Ping{
			Ping: &pb.Ping{},
		},
		PiggyBack: pbk,
	}
}

func createNackMessage(seq string, pbk *pb.PiggyBack) pb.Message {
	return pb.Message{
		Id: seq,
		Payload: &pb.Message_Nack{
			Nack: &pb.Nack{},
		},
		PiggyBack: pbk,
	}
}

func createAckMessage(seq string, pbk *pb.PiggyBack) pb.Message {
	return pb.Message{
		Id: seq,
		Payload: &pb.Message_Ack{
			Ack: &pb.Ack{},
		},
		PiggyBack: pbk,
	}
}
