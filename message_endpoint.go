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
	"errors"

	"github.com/DE-labtory/swim/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/it-chain/iLogger"

	"time"
)

var ErrSendTimeout = errors.New("Error send timeout")
var ErrUnreachable = errors.New("Error this shouldn't reach")
var ErrInvalidMessage = errors.New("Error invalid message")

// Callback is called when target member sent back to local member a message
type callback func(msg pb.Message)

// RespondHandler manages callback functions
type responseHandler struct {
	callbacks map[string]callback
}

// TODO: add old callback collectors
func NewRespondHandler() *responseHandler {
	return &responseHandler{
		callbacks: make(map[string]callback),
	}
}

func (r *responseHandler) addCallback(seq string, cb callback) {
	r.callbacks[seq] = cb
}

// Handle, each time other member sent back
// a message, callback matching message's seq is called
func (r *responseHandler) handle(msg pb.Message) {
	seq := msg.Seq
	cbFn := r.callbacks[seq]

	if cbFn == nil {
		iLogger.Panicf(nil, "")
	}

	cbFn(msg)
	delete(r.callbacks, seq)
}

func (r *responseHandler) hasCallback(seq string) bool {
	for s := range r.callbacks {
		if seq == s {
			return true
		}
	}
	return false
}

type MessageEndpointConfig struct {
	EncryptionEnabled  bool
	DefaultSendTimeout time.Duration
}

// MessageEndpoint basically do receiving packet and determine
// which logic executed based on the packet.
type MessageEndpoint struct {
	config         MessageEndpointConfig
	transport      UDPTransport
	messageHandler MessageHandler
	awareness      Awareness
	resHandler     *responseHandler
	quit           chan struct{}
}

func NewMessageEndpoint(config MessageEndpointConfig, transport UDPTransport, messageHandler MessageHandler, awareness Awareness) *MessageEndpoint {
	return &MessageEndpoint{
		config:         config,
		transport:      transport,
		messageHandler: messageHandler,
		awareness:      awareness,
		resHandler:     NewRespondHandler(),
		quit:           make(chan struct{}),
	}
}

// Listen is a log running goroutine that pulls packet from the
// transport and pass it for processing
func (m *MessageEndpoint) Listen() {
	for {
		select {
		case packet := <-m.transport.PacketCh():
			// validate packet then convert it to message
			msg, err := m.processPacket(*packet)
			if err != nil {
				iLogger.Error(nil, err.Error())
			}

			if m.resHandler.hasCallback(msg.Seq) {
				m.resHandler.handle(msg)
			} else {
				m.handleMessage(msg)
			}

		case <-m.quit:
			return
		}
	}
}

// ProcessPacket process given packet, this procedure may include
// decrypting data and converting it to message
func (m *MessageEndpoint) processPacket(packet Packet) (pb.Message, error) {
	msg := &pb.Message{}
	if m.config.EncryptionEnabled {
		// TODO: decrypt packet
	}

	if err := proto.Unmarshal(packet.Buf, msg); err != nil {
		return pb.Message{}, err
	}

	return *msg, nil
}

func (m *MessageEndpoint) Shutdown() {
	// close transport first
	m.transport.Shutdown()

	// then close message endpoint
	m.quit <- struct{}{}
}

func validateMessage(msg pb.Message) bool {
	return msg.Seq != "" &&
		msg.Payload != nil
}

// with given message handleMessage determine which logic should be executed
// based on the message type. Additionally handleMessage can call MemberDelegater
// to update member status and encrypt messages
func (m *MessageEndpoint) handleMessage(msg pb.Message) error {
	// validate message
	if validateMessage(msg) {
		// call delegate func to update members states
		m.messageHandler.handle(msg)
		return nil
	}

	return ErrInvalidMessage
}

// determineSendTimeout if @timeout is given, then use this value as timeout value
// otherwise calculate timeout value based on the awareness
func (m *MessageEndpoint) determineSendTimeout(timeout time.Duration) time.Duration {
	if timeout == (time.Duration)(nil) || timeout != time.Duration(0) {
		return m.awareness.ScaleTimeout(m.config.DefaultSendTimeout)
	}

	return timeout
}

// SyncSend synchronously send message to member of addr, waits until response come back,
// whether it is timeout or send failed, SyncSend can be used in the case of pinging to other members.
// if @timeout is provided then set send timeout to given parameters, if not then calculate
// timeout based on the its awareness
func (m *MessageEndpoint) SyncSend(addr string, msg pb.Message, interval time.Duration) (pb.Message, error) {
	onSucc := make(chan pb.Message)

	timeout := m.determineSendTimeout(interval)

	d, err := proto.Marshal(&msg)
	if err != nil {
		return pb.Message{}, err
	}

	m.resHandler.addCallback(msg.Seq, func(msg pb.Message) {
		onSucc <- msg
	})

	_, err = m.transport.WriteTo(d, addr)
	if err != nil {
		iLogger.Info(nil, err.Error())
		return pb.Message{}, err
	}

	// start timer
	T := time.NewTimer(timeout)

	select {
	case msg := <-onSucc:
		return msg, nil
	case <-T.C:
		return pb.Message{}, ErrSendTimeout
	}

	return pb.Message{}, ErrUnreachable
}

// Send asynchronously send message to member of addr, don't wait until response come back,
// after response came back, callback function executed, Send can be used in the case of
// gossip message to other members
func (m *MessageEndpoint) Send(addr string, msg pb.Message) error {
	return nil
}
