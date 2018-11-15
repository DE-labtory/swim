package swim

import (
	"errors"
	"github.com/DE-labtory/swim/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/it-chain/iLogger"

	"time"
)

var ErrSendFailed = errors.New("Error send failed")
var ErrSendTimeout = errors.New("Error send timeout")
var ErrUnreachable = errors.New("Error this shouldn't reach")

type MessageEndpointConfig struct {
	EncryptionEnabled bool
	DefaultSendTimeout time.Duration
}

// MessageEndpoint basically do receiving packet and determine
// which logic executed based on the packet.
type MessageEndpoint struct {
	config MessageEndpointConfig
	transport                  UDPTransport
	memberStatusChangeDelegate MemberStatusChangeDelegate
	awareness                  Awareness
	quit chan struct{}
}

func NewMessageEndpoint(config MessageEndpointConfig, transport UDPTransport, delegate MemberStatusChangeDelegate, awareness Awareness) *MessageEndpoint {
	return &MessageEndpoint{
		config: config,
		transport:transport,
		memberStatusChangeDelegate:delegate,
		awareness:awareness,
		quit: make(chan struct{}),
	}
}

// Listen is a log running goroutine that pulls packet from the
// transport and pass it for processing
func (m *MessageEndpoint) Listen() {
	for {
		select {
		case packet := <- m.transport.PacketCh():
			msg, err := m.processPacket(*packet)
			if err != nil {
				iLogger.Error(nil, err.Error())
			}
			m.handleMessage(msg)

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

// with given message handleMessage determine which logic should be executed
// based on the message type. Additionally handleMessage can call MemberDelegater
// to update member status and encrypt messages
func (m *MessageEndpoint) handleMessage(msg pb.Message) {
	// validate message

	// call delegate func to update members states
}

// determineSendTimeout if @timeout is given, then use this value as timeout value
// otherwise calculate timeout value based on the awareness
func (m *MessageEndpoint) determineSendTimeout(timeout time.Duration) time.Duration {
	if timeout != time.Duration(0) {
		return m.awareness.ScaleTimeout(m.config.DefaultSendTimeout)
	}

	return timeout
}

// SyncSend synchronously send message to member of addr, waits until response come back,
// whether it is timeout or send failed, SyncSend can be used in the case of pinging to other members.
// if @timeout is provided then set send timeout to given parameters, if not then calculate
// timeout based on the its awareness
func (m *MessageEndpoint) SyncSend(addr string, msg pb.Message, interval time.Duration) (pb.Message, error) {
	onSucc := make(chan struct{})
	onErr := make(chan struct{})

	timeout := m.determineSendTimeout(interval)

	d, err := proto.Marshal(&msg)
	if err != nil {
		return pb.Message{}, err
	}

	go func() {
		_, err := m.transport.WriteTo(d, addr)
		if err != nil {
			iLogger.Info(nil, err.Error())
			onErr <- struct{}{}
			return
		}
		onSucc <- struct{}{}
	}()

	// start timer
	T := time.NewTimer(timeout)

	select {
	case <- onSucc:

	case <- onErr:
		return pb.Message{}, ErrSendFailed

	case <- T.C:
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
