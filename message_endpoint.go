package swim

import (
	"github.com/DE-labtory/swim/pb"
	"github.com/gogo/protobuf/proto"

	"time"
)

type MessageEndpointConfig struct {
	EncryptionEnabled bool
}

// MessageEndpoint basically do receiving packet and determine
// which logic executed based on the packet.
type MessageEndpoint struct {
	config MessageEndpointConfig
	transport                  UDPTransport
	memberStatusChangeDelegate MemberStatusChangeDelegate
	awareness                  Awareness
}

func NewMessageEndpoint(config MessageEndpointConfig, transport UDPTransport, delegate MemberStatusChangeDelegate, awareness Awareness) *MessageEndpoint {
	return &MessageEndpoint{
		config: config,
		transport:transport,
		memberStatusChangeDelegate:delegate,
		awareness:awareness,
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

			}
			m.handleMessage(msg)
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

// with given message handleMessage determine which logic should be executed
// based on the message type. Additionally handleMessage can call MemberDelegater
// to update member status and encrypt messages
func (m *MessageEndpoint) handleMessage(msg pb.Message) {
	// validate message
	
	// call delegate func to update members states
}

// SyncSend synchronously send message to member of addr, waits until response come back,
// whether it is timeout or send failed, SyncSend can be used in the case of pinging to other members.
// if @timeout is provided then set send timeout to given parameters, if not then calculate
// timeout based on the its awareness
func (m *MessageEndpoint) SyncSend(addr string, msg pb.Message, timeout time.Duration) (pb.Message, error) {
	return pb.Message{}, nil
}

// Send asynchronously send message to member of addr, don't wait until response come back,
// after response came back, callback function executed, Send can be used in the case of
// gossip message to other members
func (m *MessageEndpoint) Send(addr string, msg pb.Message) error {
	return nil
}
