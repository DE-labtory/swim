package swim

import (
	"github.com/DE-labtory/swim/pb"
)

// MessageEndpoint basically do receiving packet and determine
// which logic executed based on the packet.
type MessageEndpoint struct {
	transport                  UDPTransport
	memberStatusChangeDelegate MemberStatusChangeDelegate
	awareness                  Awareness
}

// Listen is a log running goroutine that pulls packet from the
// transport and pass it for processing
func (m *MessageEndpoint) Listen() {

}

// ProcessPacket process given packet, this procedure may include
// decrypting data and converting it to message
func (m *MessageEndpoint) processPacket(packet Packet) pb.Message {
	return pb.Message{}
}

// with given message handleMessage determine which logic should be executed
// based on the message type. Additionally handleMessage can call MemberDelegater
// to update member status and encrypt messages
func (m *MessageEndpoint) handleMessage(msg pb.Message) {
	// validate message
	// call delegate func to update members states
}

// SyncSend synchronously send message to member of addr, waits until response come back,
// SyncSend can be used in the case of pinging to other members
func (m *MessageEndpoint) SyncSend(addr string, msg pb.Message) (<-chan pb.Message, error) {
	return nil, nil
}

// Send asynchronously send message to member of addr, don't wait until response come back,
// after response came back, callback function executed, Send can be used in the case of
// gossip message to other members
func (m *MessageEndpoint) Send(addr string, msg pb.Message) error {
	return nil
}
