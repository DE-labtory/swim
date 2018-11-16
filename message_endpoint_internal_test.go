package swim

import (
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
	rh := newRespondHandler()
	assert.NotNil(t, rh)
}

func TestResponseHandler_addCallback(t *testing.T) {
	rh := newRespondHandler()
	seq := "seq1"
	msg := pb.Message{Seq: seq}
	cb := func(msg pb.Message) {
		assert.Equal(t, seq, msg.Seq)
	}

	rh.addCallback(seq, cb)
	rh.callbacks[seq](msg)
}

// responseHandle.handle when target callback exist
// this test expect successfully call callback function
// and remove target callback from map
func TestResponseHandler_handle_callback_exist(t *testing.T) {
	rh := newRespondHandler()
	seq := "seq1"
	msg := pb.Message{Seq: seq}
	cb := func(msg pb.Message) {
		assert.Equal(t, seq, msg.Seq)
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
	rh := newRespondHandler()
	seq := "seq1"
	msg := pb.Message{Seq: "arbitrarySeq"}
	cb := func(msg pb.Message) {}

	rh.addCallback(seq, cb)
	assert.Panics(t, func() { rh.handle(msg) }, "Panic, no matching callback function")
}

func TestResponseHandler_hasCallback(t *testing.T) {
	rh := newRespondHandler()
	cb1 := func(msg pb.Message) {}
	cb2 := func(msg pb.Message) {}

	rh.addCallback("seq1", cb1)
	rh.addCallback("seq2", cb2)

	// test when target callback exist
	result := rh.hasCallback("seq1")
	assert.Equal(t, result, true)

	// test when target callback not exist
	result = rh.hasCallback("seq3")
	assert.Equal(t, result, false)
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
	e := NewMessageEndpoint(mc, p, h, a)

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
		BindPort:    11111,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	a := NewAwareness(8)
	e := NewMessageEndpoint(mc, p, h, a)

	// ** add callback function with key "seq1" **
	e.resHandler.addCallback("seq1", func(msg2 pb.Message) {
		// message assertion
		assert.Equal(t, msg2.Seq, (*msg).Seq)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
		wg.Done()
	})

	// start to listen
	go e.Listen()

	data, _ := proto.Marshal(msg)

	// send marshaled data to MessageEndpoint
	_, err = p.WriteTo(data, "127.0.0.1:11111")
	assert.NoError(t, err)
	wg.Wait()
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
