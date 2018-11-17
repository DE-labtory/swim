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
	msg := pb.Message{Id: seq}
	cb := callback{
		fn: func(msg pb.Message) {
			assert.Equal(t, seq, msg.Id)
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
	msg := pb.Message{Id: seq}
	cb := callback{
		fn: func(msg pb.Message) {
			assert.Equal(t, seq, msg.Id)
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
	msg := pb.Message{Id: "arbitrarySeq"}
	cb := callback{
		fn:      func(msg pb.Message) {},
		created: time.Now(),
	}

	rh.addCallback(seq, cb)
	assert.Panics(t, func() { rh.handle(msg) }, "Panic, no matching callback function")
}

func TestResponseHandler_hasCallback(t *testing.T) {
	rh := newResponseHandler(time.Hour)
	cb1 := callback{
		fn:      func(msg pb.Message) {},
		created: time.Now(),
	}
	cb2 := callback{
		fn:      func(msg pb.Message) {},
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
		fn:      func(msg pb.Message) {},
		created: time.Now(),
	}
	cb2 := callback{
		fn:      func(msg pb.Message) {},
		created: time.Now().Add(time.Second * 4),
	}

	rh.addCallback("seq1", cb1)
	rh.addCallback("seq2", cb2)

	assert.Equal(t, len(rh.callbacks), 2)

	// sleep 3 sec, cb1 should be collected
	time.Sleep(time.Second * 3)

	assert.Equal(t, len(rh.callbacks), 1)

	// sleep 4 sec, cb2 should be collected
	time.Sleep(time.Second * 4)

	assert.Equal(t, len(rh.callbacks), 0)
}

func TestNewMessageEndpoint(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11110,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	// if CallbackCollectInterval not provided, then should return error
	mc := MessageEndpointConfig{
		EncryptionEnabled:  false,
		DefaultSendTimeout: time.Second,
	}

	h := MockMessageHandler{}

	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.Error(t, err, ErrCallbackCollectIntervalNotSpecified)
	assert.Nil(t, e)

	// if CallbackCollectInterval provided, then should return object successfully
	mc2 := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	e2, err := NewMessageEndpoint(mc2, p, h, a)
	assert.NoError(t, err)
	assert.NotNil(t, e2)
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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	// mocking MessageHandler to verify message
	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		// message assertion
		assert.Equal(t, msg2.Id, (*msg).Id)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))

		// after assertion done wait group to finish test
		wg.Done()
	}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	defer func() {
		e.Shutdown()
	}()

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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
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
			assert.Equal(t, msg2.Id, (*msg).Id)
			assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
			wg.Done()
		},
		created: time.Now(),
	})

	defer func() {
		e.Shutdown()
	}()

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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
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
	assert.Equal(t, processed.Id, (*msg).Id)
	assert.Equal(t, getAckPayload(processed), getAckPayload(*msg))
}

// validateMessage should have both Seq value and Payload
func Test_validateMessage(t *testing.T) {
	// 1. message w/o payload expect to false
	msg1 := pb.Message{Id: "seq1"}
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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to handle
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Id, (*msg).Id)
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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare invalid message to return error
	msg := getAckMessage("", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Id, (*msg).Id)
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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
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
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
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
func TestMessageEndpoint_SyncSend_MySelf(t *testing.T) {
	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11118,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	a := NewAwareness(8)

	// increase nsa score in advance to test decreasing nsa score
	// in the case when member probe successfully, self node actually
	// decrease nsa point
	a.ApplyDelta(2)

	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	defer func() {
		e.Shutdown()
	}()

	// start to listen
	go e.Listen()

	sentMsg, err := e.SyncSend("127.0.0.1:11118", *msg, time.Second*2)

	assert.NoError(t, err)
	assert.Equal(t, sentMsg.Id, (*msg).Id)
	assert.Equal(t, getAckPayload(sentMsg), getAckPayload(*msg))
	assert.Equal(t, e.awareness.GetHealthScore(), 1)
}

func TestMessageEndpoint_SyncSend_TwoMembers(t *testing.T) {
	receiverTransportConfig := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11130,
	}
	p, err := NewPacketTransport(&receiverTransportConfig)
	assert.NoError(t, err)

	senderTransportConfig := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11131,
	}
	p2, err := NewPacketTransport(&senderTransportConfig)
	assert.NoError(t, err)

	receiverConfig := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}
	senderConfig := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg pb.Message) {}
	a := NewAwareness(8)
	// receiver
	receiver, err := NewMessageEndpoint(receiverConfig, p, h, a)
	assert.NoError(t, err)

	h2 := MockMessageHandler{}
	h2.handleFunc = func(msg pb.Message) {}
	a2 := NewAwareness(8)

	// increase nsa score in advance to test decreasing nsa score
	// in the case when member probe successfully, self node actually
	// decrease nsa point
	a2.ApplyDelta(2)

	// sender
	sender, err := NewMessageEndpoint(senderConfig, p2, h2, a2)

	// start to listen
	go receiver.Listen()
	go sender.Listen()

	// mocking receiver sent-back to sender a response
	// by continuously send ack message to sender every 1 sec

	quit := make(chan bool)
	go func() {
		T := time.NewTicker(time.Second)

		for {
			select {
			case <-T.C:
				receiver.Send("127.0.0.1:11131", *msg)
			case <-quit:
				return
			}
		}
	}()

	defer func() {
		receiver.Shutdown()
		sender.Shutdown()
		quit <- true
	}()

	// sender sends message to receiver, and receiver send back a ack message
	// so sender's SyncSend return received message. as a result decrease NSA score w/o error
	receivedMsg, err := sender.SyncSend("127.0.0.1:11130", *msg, time.Second*2)

	assert.NoError(t, err)
	assert.Equal(t, receivedMsg.Id, (*msg).Id)
	assert.Equal(t, getAckPayload(receivedMsg), getAckPayload(*msg))

	// sender's NSA score will decrease
	assert.Equal(t, 1, sender.awareness.GetHealthScore())
}

// SyncSend internal timeout test send message to receiver
// 1. SyncSend register callback function
// 2. send message after expiring SyncSend timer
// 3. increase NSA score
func TestMessageEndpoint_SyncSend_When_Timeout(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11119,
	}
	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	tc2 := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11120,
	}
	p2, err := NewPacketTransport(&tc2)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}
	mc2 := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Id, (*msg).Id)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
		wg.Done()
	}
	a := NewAwareness(8)
	// receiver
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	// start to listen
	go e.Listen()

	h2 := MockMessageHandler{}
	a2 := NewAwareness(8)
	// sender
	e2, err := NewMessageEndpoint(mc2, p2, h2, a2)

	defer func() {
		e.Shutdown()
	}()

	// sender sends message to receiver, but receiver would never send back any message
	// so sender's SyncSend timer will expire. as a result increase NSA score with error
	sentMsg, err := e2.SyncSend("127.0.0.1:11119", *msg, time.Second*2)

	assert.Error(t, err)
	assert.Equal(t, pb.Message{}, sentMsg)
	// sender's NSA score will increase
	assert.Equal(t, 1, e2.awareness.GetHealthScore())

	wg.Wait()
}

// Send internal test send message to myself
// 1. send message using Transport
// 2. Listen receive sent message
// 3. then call message handler handle function
func TestMessageEndpoint_Send(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	tc := PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    11121,
	}

	p, err := NewPacketTransport(&tc)
	assert.NoError(t, err)

	mc := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      time.Second,
		CallbackCollectInterval: time.Hour,
	}

	// prepare message to send
	msg := getAckMessage("seq1", "payload1")

	// sent data should be handled by MessageHandler
	h := MockMessageHandler{}
	h.handleFunc = func(msg2 pb.Message) {
		assert.Equal(t, msg2.Id, (*msg).Id)
		assert.Equal(t, getAckPayload(msg2), getAckPayload(*msg))
		wg.Done()
	}
	a := NewAwareness(8)
	e, err := NewMessageEndpoint(mc, p, h, a)
	assert.NoError(t, err)

	defer func() {
		e.Shutdown()
	}()

	// start to listen
	go e.Listen()

	err = e.Send("127.0.0.1:11121", *msg)
	assert.NoError(t, err)
	wg.Wait()
}

func getAckMessage(seq, payload string) *pb.Message {
	return &pb.Message{
		Id: seq,
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
