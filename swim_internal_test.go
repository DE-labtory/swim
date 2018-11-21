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
	"sync"
	"testing"
	"time"

	"github.com/DE-labtory/swim/pb"
	"github.com/stretchr/testify/assert"
)

type MockPbkStore struct {
	LenFunc     func() int
	PushFunc    func(pbk pb.PiggyBack)
	GetFunc     func() (pb.PiggyBack, error)
	IsEmptyFunc func() bool
}

func (p MockPbkStore) Len() int {
	return p.LenFunc()
}
func (p MockPbkStore) Push(pbk pb.PiggyBack) {
	p.PushFunc(pbk)
}
func (p MockPbkStore) Get() (pb.PiggyBack, error) {
	return p.GetFunc()
}
func (p MockPbkStore) IsEmpty() bool {
	return p.IsEmptyFunc()
}

func TestSWIM_ShutDown(t *testing.T) {
	s := New(&Config{
		K:             2,
		T:             4000,
		AckTimeOut:    1000,
		MaxlocalCount: 1,
	}, &SuspicionConfig{})

	m := NewMemberMap(&SuspicionConfig{})
	m.Alive(AliveMessage{
		MemberMessage: MemberMessage{
			ID: "1",
		},
	})

	m.Alive(AliveMessage{
		MemberMessage: MemberMessage{
			ID: "2",
		},
	})

	s.memberMap = m

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Check whether the startFailureDetector ends or not
	go func() {
		s.startFailureDetector()
		wg.Done()
	}()

	time.Sleep(5 * time.Second)

	// End startFailureDetector
	s.ShutDown()

	wg.Wait()
}

func TestSWIM_handlePbk(t *testing.T) {

}

// This is successful scenario when target member response in time with
// ack message, then mediator sends back ack message to source member
func TestSWIM_handleIndirectPing(t *testing.T) {
	seq := "1"

	pbkStore := MockPbkStore{}
	pbkStore.GetFunc = func() (pb.PiggyBack, error) {
		return pb.PiggyBack{
			Type:        pb.PiggyBack_Alive,
			Id:          "pbk_id1",
			Incarnation: 123,
			Address:     "address123",
		}, nil
	}

	swim := SWIM{}
	mIMessageHandler := MockMessageHandler{}
	mJMessageHandler := MockMessageHandler{}

	mK := createMessageEndpoint(&swim, time.Second, 11140)
	mI := createMessageEndpoint(&mIMessageHandler, time.Second, 11141)
	mJ := createMessageEndpoint(&mJMessageHandler, time.Second, 11142)

	swim.messageEndpoint = mK
	swim.pbkStore = pbkStore

	// ** m_k's handleIndirectPing is implicitly called when m_k received
	// indirect-ping message from other member **
	go mK.Listen()
	go mI.Listen()
	go mJ.Listen()

	ind := pb.Message{
		Seq: seq,
		// address of source member
		Address: "127.0.0.1:11141",
		Payload: &pb.Message_IndirectPing{
			IndirectPing: &pb.IndirectPing{
				// target address
				Target: "127.0.0.1:11142",
			},
		},
		PiggyBack: &pb.PiggyBack{},
	}

	defer func() {
		mK.Shutdown()
		mI.Shutdown()
		mJ.Shutdown()
	}()

	mJMessageHandler.handleFunc = func(msg pb.Message) {
		// check whether msg.Payload type is *pb.Message_Ping
		assert.NotNil(t, msg.Payload.(*pb.Message_Ping))
		assert.Equal(t, msg.PiggyBack.Address, "address123")
		assert.Equal(t, msg.PiggyBack.Incarnation, uint32(123))
		assert.Equal(t, msg.PiggyBack.Id, "pbk_id1")

		ack := pb.Message{Seq: msg.Seq, Payload: &pb.Message_Ack{Ack: &pb.Ack{}}, PiggyBack: &pb.PiggyBack{}}
		mJ.Send("127.0.0.1:11140", ack)
	}

	resp, err := mI.SyncSend("127.0.0.1:11140", ind, DefaultSendTimeout)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Payload.(*pb.Message_Ack))
	assert.Equal(t, resp.Seq, seq)
	assert.Equal(t, resp.PiggyBack.Address, "address123")
	assert.Equal(t, resp.PiggyBack.Incarnation, uint32(123))
	assert.Equal(t, resp.PiggyBack.Id, "pbk_id1")
}

// This is NOT-successful scenario when target member DID NOT response in time
// ack message, then mediator sends back NACK message to source member
func TestSWIM_handleIndirectPing_Target_Timeout(t *testing.T) {
	seq := "1"

	pbkStore := MockPbkStore{}
	pbkStore.GetFunc = func() (pb.PiggyBack, error) {
		return pb.PiggyBack{
			Type:        pb.PiggyBack_Alive,
			Id:          "pbk_id1",
			Incarnation: 123,
			Address:     "address123",
		}, nil
	}

	swim := SWIM{}
	mIMessageHandler := MockMessageHandler{}
	mJMessageHandler := MockMessageHandler{}

	mK := createMessageEndpoint(&swim, time.Second, 11140)
	// source should have larger send timeout, because source should give mediator
	// enough time to ping to target
	mI := createMessageEndpoint(&mIMessageHandler, time.Second*3, 11141)
	mJ := createMessageEndpoint(&mJMessageHandler, time.Second, 11142)

	swim.messageEndpoint = mK
	swim.pbkStore = pbkStore

	// ** m_k's handleIndirectPing is implicitly called when m_k received
	// indirect-ping message from other member **
	go mK.Listen()

	go mI.Listen()
	go mJ.Listen()

	ind := pb.Message{
		Seq: seq,
		// address of source member
		Address: "127.0.0.1:11141",
		Payload: &pb.Message_IndirectPing{
			IndirectPing: &pb.IndirectPing{
				// target address
				Target: "127.0.0.1:11142",
			},
		},
		PiggyBack: &pb.PiggyBack{},
	}

	defer func() {
		mI.Shutdown()
		mK.Shutdown()
		mJ.Shutdown()
	}()

	mJMessageHandler.handleFunc = func(msg pb.Message) {
		// check whether msg.Payload type is *pb.Message_Ping
		assert.NotNil(t, msg.Payload.(*pb.Message_Ping))
		assert.Equal(t, msg.PiggyBack.Address, "address123")
		assert.Equal(t, msg.PiggyBack.Incarnation, uint32(123))
		assert.Equal(t, msg.PiggyBack.Id, "pbk_id1")

		// DO NOT ANYTHING: do not response back to m
	}

	resp, err := mI.SyncSend("127.0.0.1:11140", ind, DefaultSendTimeout)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Payload.(*pb.Message_Nack))
	assert.Equal(t, resp.Seq, seq)
	assert.Equal(t, resp.PiggyBack.Address, "address123")
	assert.Equal(t, resp.PiggyBack.Incarnation, uint32(123))
	assert.Equal(t, resp.PiggyBack.Id, "pbk_id1")
}

func createMessageEndpoint(messageHandler MessageHandler, sendTimeout time.Duration, port int) *MessageEndpoint {
	mConfig := MessageEndpointConfig{
		EncryptionEnabled:       false,
		DefaultSendTimeout:      sendTimeout,
		CallbackCollectInterval: time.Hour,
	}

	tConfig := &PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    port,
	}

	transport, _ := NewPacketTransport(tConfig)

	awareness := NewAwareness(8)

	m, _ := NewMessageEndpoint(mConfig, transport, messageHandler, awareness)

	return m
}
