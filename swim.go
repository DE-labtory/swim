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

	"log"

	"github.com/DE-labtory/swim/pb"
	"github.com/rs/xid"
)

type Config struct {

	// The maximum number of times the same piggyback data can be queried
	MaxlocalCount int

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

	messageEndpoint *MessageEndpoint

	priorityPBStore *PriorityPBStore

	// FailureDetector quit channel
	quitFD chan struct{}
}

func New(config *Config, messageEndpointConfig MessageEndpointConfig, awareness *Awareness) *SWIM {
	if config.T < config.AckTimeOut {
		panic("T time must be longer than ack time-out")
	}

	swim := SWIM{
		config:          config,
		memberMap:       NewMemberMap(),
		messageEndpoint: nil,
		priorityPBStore: NewPriorityPBStore(config.MaxlocalCount),
		quitFD:          make(chan struct{}),
	}

	messageEndpoint := setMessageEndpoint(config, messageEndpointConfig, &swim, awareness)
	swim.messageEndpoint = messageEndpoint

	return &swim
}

// set MessageEndpoint
func setMessageEndpoint(config *Config, messageEndpointConfig MessageEndpointConfig, messageHandler MessageHandler, awareness *Awareness) *MessageEndpoint {
	packetTransportConfig := PacketTransportConfig{
		BindAddress: config.BindAddress,
		BindPort:    config.BindPort,
	}

	packetTransport, err := NewPacketTransport(&packetTransportConfig)
	if err != nil {
		log.Panic(err)
	}

	messageEndpoint, err := NewMessageEndpoint(messageEndpointConfig, packetTransport, messageHandler, awareness)
	if err != nil {
		log.Panic(err)
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
		Address := s.config.BindAddress + ":" + string(s.config.BindPort)
		s.messageEndpoint.Send(msg.Address, pb.Message{
			Seq:     xid.New().String(),
			Address: Address,
			Payload: &pb.Message_Ack{
				Ack: &pb.Ack{Payload: ""},
			},
			PiggyBack: &pb.PiggyBack{},
		})
	case *pb.Message_Ack:
		// handle ack
	case *pb.Message_IndirectPing:
		// handle indirect ping
	default:

	}
}

// handle piggyback related to member status
func (s *SWIM) handlePbk(piggyBack *pb.PiggyBack) {

}
