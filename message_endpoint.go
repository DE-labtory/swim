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
var ErrCallbackCollectIntervalNotSpecified = errors.New("Error callback collect interval should be specified")

// sallback is called when target member sent back to local member a message
// created field is for clean up the old callback
type callback struct {
	fn func(msg pb.Message)
	created time.Time
}

// responseHandler manages callback functions
type responseHandler struct {
	callbacks map[string]callback
	collectInterval time.Duration
}

func newResponseHandler(collectInterval time.Duration) *responseHandler {
	h := &responseHandler{
		callbacks: make(map[string]callback),
		collectInterval: collectInterval,
	}

	go h.collectCallback()

	return h
}

func (r *responseHandler) addCallback(seq string, cb callback) {
	r.callbacks[seq] = cb
}

// Handle, each time other member sent back
// a message, callback matching message's seq is called
func (r *responseHandler) handle(msg pb.Message) {
	seq := msg.Seq
	cb, exist := r.callbacks[seq]

	if exist == false {
		iLogger.Panicf(nil, "Panic, no matching callback function")
	}

	cb.fn(msg)
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

// collectCallback every time callbackCollectInterval expired clean up
// the old (time.now - callback.created > time interval) callback deleted from map
// callbackCollectInterval specified in config
func (r *responseHandler) collectCallback() {
	timeout := r.collectInterval
	T := time.NewTimer(timeout)

	for {
		select {
		case <- T.C:
			for seq, cb := range r.callbacks {
				if time.Now().Sub(cb.created) > timeout {
					delete(r.callbacks, seq)
				}
			}
			T = time.NewTimer(timeout)
		}
	}
}

type MessageEndpointConfig struct {
	EncryptionEnabled  bool
	DefaultSendTimeout time.Duration

	// callbackCollect Interval indicate time interval to clean up old
	// callback function
	CallbackCollectInterval time.Duration
}

// MessageEndpoint basically do receiving packet and determine
// which logic executed based on the packet.
type MessageEndpoint struct {
	config         MessageEndpointConfig
	transport      UDPTransport
	messageHandler MessageHandler
	awareness      *Awareness
	resHandler     *responseHandler
	quit           chan struct{}
}

func NewMessageEndpoint(config MessageEndpointConfig, transport UDPTransport, messageHandler MessageHandler, awareness *Awareness) (*MessageEndpoint, error) {
	if config.CallbackCollectInterval == time.Duration(0) {
		return nil, ErrCallbackCollectIntervalNotSpecified
	}

	return &MessageEndpoint{
		config:         config,
		transport:      transport,
		messageHandler: messageHandler,
		awareness:      awareness,
		resHandler:     newResponseHandler(config.CallbackCollectInterval),
		quit:           make(chan struct{}),
	}, nil
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
				iLogger.Panic(nil, err.Error())
			}

			// before message that come from other handle by MessageHandler
			// check whether this message is sent-back message from other member
			// this is determined by message's Seq property which work as message id

			if m.resHandler.hasCallback(msg.Seq) {
				m.resHandler.handle(msg)
			} else {
				if err := m.handleMessage(msg); err != nil {
					iLogger.Panic(nil, err.Error())
				}
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
	if msg.Seq == "" {
		iLogger.Info(nil, "message seq value empty")
		return false
	}

	if msg.Payload == nil {
		iLogger.Info(nil, "message payload value empty")
		return false
	}

	return true
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
	if timeout == time.Duration(0) {
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

	// if @interval provided then, use parameters, if not timeout determine by
	// its awareness score
	timeout := m.determineSendTimeout(interval)

	d, err := proto.Marshal(&msg)
	if err != nil {
		return pb.Message{}, err
	}

	// register callback function, this callback function is called when
	// member with @addr sent back us packet
	m.resHandler.addCallback(msg.Seq, callback {
		fn: func(msg pb.Message) {
			onSucc <- msg
		},
		created: time.Now(),
	})

	// send the message
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
