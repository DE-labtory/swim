package swim

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/DE-labtory/swim/pb"

	"github.com/stretchr/testify/assert"
)

type MockMessageHandler struct {
	handleFunc func(msg pb.Message)
}

func (h MockMessageHandler) handle(msg pb.Message) {
	h.handleFunc(msg)
}

func TestNewResponseHandler(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	assert.NotNil(t, rh)
}

func TestResponseHandler_addCallback(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	seq := "seq1"
	msg := pb.Message{Seq: seq}
	cb := callback{
		fn: func(msg pb.Message) {
			assert.Equal(t, seq, msg.Seq)
		},
		created: time.Now(),
	}

	rh.addCallback(seq, cb)
	rh.callbacks[seq].fn(msg)
}

// responseHandle.handle when target callback exist
// this test expect successfully call callback function
// and remove target callback from map
func TestResponseHandler_handle_callback_exist(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	seq := "seq1"
	msg := pb.Message{Seq: seq}
	cb := callback{
		fn: func(msg pb.Message) {
			assert.Equal(t, seq, msg.Seq)
		},
		created: time.Now(),
	}

	rh.addCallback(seq, cb)
	rh.handle(msg)

	// check the callback has deleted
	_, ok := rh.callbacks[seq]
	assert.Equal(t, ok, false)
}

// responseHandle.handle when target callback does not exist
// panic in that case
func TestResponseHandler_handle_callback_not_exist(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	seq := "seq1"
	msg := pb.Message{Seq: "arbitrarySeq"}
	cb := callback{
		fn: func(msg pb.Message) {},
		created: time.Now(),
	}

	rh.addCallback(seq, cb)
	assert.Panics(t, func() { rh.handle(msg) }, "Panic, no matching callback function")
}

func TestResponseHandler_hasCallback(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	cb1 := callback{
		fn: func(msg pb.Message) {},
		created: time.Now(),
	}
	cb2 := callback{
		fn: func(msg pb.Message) {},
		created: time.Now(),
	}

	rh.addCallback("seq1", cb1)
	rh.addCallback("seq2", cb2)

	// test when target callback exist
	result := rh.hasCallback("seq1")
	assert.Equal(t, result, true)

	// test when target callback not exist
	result = rh.hasCallback("seq3")
	assert.Equal(t, result, false)
}

func TestResponseHandler_collectCallback(t *testing.T) {
	rh := newResponseHandler(time.Second * 2)
	cb1 := callback{
		fn: func(msg pb.Message) {},
		created: time.Now(),
	}
	cb2 := callback{
		fn: func(msg pb.Message) {},
		created: time.Now().Add(time.Second * 4),
	}

	rh.addCallback("seq1", cb1)
	rh.addCallback("seq2", cb2)

	assert.Equal(t, len(rh.callbacks), 2)

	time.Sleep(time.Second * 3)

	assert.Equal(t, len(rh.callbacks), 1)

	time.Sleep(time.Second * 4)

	assert.Equal(t, len(rh.callbacks), 0)
}

// MessageEndpoint.Listen when there's no callback matching with message.Seq
// then MessageHandler handle function called
// here we test whether MessageHandler handle function successfully called
func TestMessageEndpoint_Listen_NoCallback(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11111,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	// mocking MessageHandler to verify message
	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		// message assertion
		assert.Equal(t, msg2.Seq, (*msg).Seq)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))

		// after assertion done wait group to finish test
		wg.Done()
	}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	// start to listen
	go e.Listen()

	data, _ := proto.Marshal(msg)

	// send marshaled data to MessageEndpoint
	_, err = p.WriteTo(data, "127.0.0.1:11111")
	assert.NoError(t, err)
	wg.Wait()
}

// MessageEndpoint.Listen when there's callback matching with message.Seq
// then responseHandler callback function called
func TestMessageEndpoint_Listen_HaveCallback(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11112,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	// ** add callback function with key "seq1" **
	e.resHandler.addCallback("seq1", callback{
		fn: func(msg2 pb.Message) {
			// message assertion
			assert.Equal(t, msg2.Seq, (*msg).Seq)
			assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
			wg.Done()
		},
		created: time.Now(),
	})

	// start to listen
	go e.Listen()

	data, _ := proto.Marshal(msg)

	// send marshaled data to MessageEndpoint
	_, err = p.WriteTo(data, "127.0.0.1:11112")
	assert.NoError(t, err)
	wg.Wait()
}

func TestMessageEndpoint_processPacket(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11113,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")
	data, _ := proto.Marshal(msg)

	// create packet to process
	packet := createPacket("test.addr", data)
	processed, err := e.processPacket(packet)

	assert.NoError(t, err)
	assert.Equal(t, processed.Seq, (*msg).Seq)
	assert.Equal(t, getAckPayload(processed), getAckPayload(*msg))
}

// validateMessage should have both Seq value and Payload
func Test_validateMessage(t *testing.T) {
	// 1. message w/o payload expect to false
	msg1 := pb.Message{Seq: "seq1"}
	assert.Equal(t, validateMessage(msg1), false)

	// 2. message w/o seq expect to false
	msg2 := getAckMessage("", "payload")
	assert.Equal(t, validateMessage(*msg2), false)

	// 3. message w/ both seq and payload expect to true
	msg3 := getAckMessage("seq2", "payload")
	assert.Equal(t, validateMessage(*msg3), true)
}

func TestMessageEndpoint_handleMessage(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11114,
	}

	p, _ := NewPacketTransport(&tc)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to handle
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Seq, (*msg).Seq)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
	}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	err = e.handleMessage(*msg)
	assert.NoError(t, err)
}

func TestMessageEndpoint_handleMessage_InvalidMessage(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11115,
	}

	p, _ := NewPacketTransport(&tc)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare invalid message to return error
	msg := getAckMessage("", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Seq, (*msg).Seq)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
	}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	err = e.handleMessage(*msg)
	assert.Equal(t, err, ErrInvalidMessage)
}

func TestMessageEndpoint_determineSendTimeout_timeout_provided(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11116,
	}

	p, _ := NewPacketTransport(&tc)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	duration := e.determineSendTimeout(time.Second * 2)
	assert.Equal(t, duration, time.Second*2)
}

func TestMessageEndpoint_determineSendTimeout_timeout_not_provided(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11117,
	}

	p, _ := NewPacketTransport(&tc)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	duration := e.determineSendTimeout(time.Duration(0))

	// default awareness score is 0 so sec * (0 + 1) = 1sec
	assert.Equal(t, duration, time.Second)
}

// SyncSend internal test send message to myself
// 1. SyncSend register callback function
// 2. send message using Transport
// 3. Listen receive sent message
// 4. then call registered callback function
func TestMessageEndpoint_SyncSend(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11118,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	// start to listen
	go e.Listen()

	sentMsg, err := e.SyncSend("127.0.0.1:11118", *msg, time.Second*2)

	assert.NoError(t, err)
	assert.Equal(t, sentMsg.Seq, (*msg).Seq)
	assert.Equal(t, getAckPayload(sentMsg), getAckPayload(*msg))
}

func getAckMessage(seq, payload string) *pb.Message {
	return &pb.Message{
		Seq: seq,
		Payload: &pb.Message_Ack{
			Ack: &pb.Ack{Payload: payload},
		},
	}
}

func getAckPayload(ack pb.Message) string {
	return (*(ack.Payload).(*pb.Message_Ack)).Ack.Payload
}

func createPacket(addr string, data []byte) Packet {
	netAddr, _ := net.ResolveUDPAddr("udp", addr)
	return Packet{
		Buf:       data,
		Addr:      netAddr,
		Timestamp: time.Now(),
	}
}
