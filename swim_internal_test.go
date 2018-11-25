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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/DE-labtory/swim/pb"
	"github.com/stretchr/testify/assert"
)

type MockMbrStatsMsgStore struct {
	LenFunc     func() int
	PushFunc    func(mbrStatsMsg pb.MbrStatsMsg)
	GetFunc     func() (pb.MbrStatsMsg, error)
	IsEmptyFunc func() bool
}

func (p MockMbrStatsMsgStore) Len() int {
	return p.LenFunc()
}
func (p MockMbrStatsMsgStore) Push(mbrStatsMsg pb.MbrStatsMsg) {
	p.PushFunc(mbrStatsMsg)
}
func (p MockMbrStatsMsgStore) Get() (pb.MbrStatsMsg, error) {
	return p.GetFunc()
}
func (p MockMbrStatsMsgStore) IsEmpty() bool {
	return p.IsEmptyFunc()
}

func TestSWIM_ShutDown(t *testing.T) {
	s := New(&Config{
		K:             2,
		T:             4000,
		AckTimeOut:    1000,
		MaxlocalCount: 1,
		MaxNsaCounter: 8,
		BindAddress:   "127.0.0.1",
		BindPort:      3000,
	},
		&SuspicionConfig{},
		MessageEndpointConfig{
			CallbackCollectInterval: 1000,
		},
		&Member{},
	)
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
		s.Start()
		wg.Done()
	}()

	time.Sleep(5 * time.Second)

	// End startFailureDetector
	s.ShutDown()

	wg.Wait()
}

// When you receive message that you are suspected, you refute
// inc always sets the standard in small inc
// TEST : When your inc is bigger than(or same) received MbrStatsMsg data.
func TestSWIM_handlembrStatsMsg_refute_bigger(t *testing.T) {
	heapSize := 1
	q := setUpHeap(heapSize)

	mbrStatsMsgStore := MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		item := Item{
			value:    mbrStatsMsg,
			priority: 1,
		}
		q.Push(&item)
	}

	mbrStatsMsgStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return q.Pop().(*Item).value.(pb.MbrStatsMsg), nil
	}

	swim := SWIM{}
	mI := createMessageEndpoint(&swim, time.Second, 11140)

	swim.awareness = NewAwareness(8)
	swim.messageEndpoint = mI
	swim.mbrStatsMsgStore = mbrStatsMsgStore
	swim.member = &Member{
		ID:          MemberID{"abcde"},
		Incarnation: 5,
	}

	msg := pb.Message{
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{
				Id:          "abcde",
				Incarnation: 5,
				Address:     "127.0.0.1:11141",
			},
		},
	}

	swim.handleMbrStatsMsg(msg.PiggyBack.MbrStatsMsg)

	assert.Equal(t, swim.member.Incarnation, uint32(6))

	mbrStatsMsg, err := swim.mbrStatsMsgStore.Get()
	assert.NoError(t, err)
	assert.Equal(t, &mbrStatsMsg, msg.PiggyBack.MbrStatsMsg)
}

// When you receive message that you are suspected, you refute
// inc always sets the standard in small inc
// TEST : When your inc is less than your MbrStatsMsg data.
func TestSWIM_handlembrStatsMsg_refute_less(t *testing.T) {
	heapSize := 1
	q := setUpHeap(heapSize)

	mbrStatsMsgStore := MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		item := Item{
			value:    mbrStatsMsg,
			priority: 1,
		}
		q.Push(&item)
	}

	mbrStatsMsgStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return q.Pop().(*Item).value.(pb.MbrStatsMsg), nil
	}

	swim := SWIM{}
	mI := createMessageEndpoint(&swim, time.Second, 11140)

	swim.awareness = NewAwareness(8)
	swim.messageEndpoint = mI
	swim.mbrStatsMsgStore = mbrStatsMsgStore
	swim.member = &Member{
		ID:          MemberID{"abcde"},
		Incarnation: 3,
	}

	msg := pb.Message{
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{
				Id:          "abcde",
				Incarnation: 5,
				Address:     "127.0.0.1:11141",
			},
		},
	}

	swim.handleMbrStatsMsg(msg.PiggyBack.MbrStatsMsg)

	assert.Equal(t, swim.member.Incarnation, uint32(4))

	mbrStatsMsg, err := swim.mbrStatsMsgStore.Get()
	assert.NoError(t, err)
	assert.Equal(t, &mbrStatsMsg, msg.PiggyBack.MbrStatsMsg)
}

func TestSWIM_HandleMbrStatsMsg_AliveMsg(t *testing.T) {
	// setup MbrStatsMsgStore
	var pushedMbrStatsMsg pb.MbrStatsMsg

	mbrStatsMsgStore := &MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		pushedMbrStatsMsg = mbrStatsMsg
	}

	// setup member_map
	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:     MemberID{ID: "1"},
		Status: Dead,
	}

	// setup SWIM
	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
		memberMap:        mm,
		mbrStatsMsgStore: mbrStatsMsgStore,
	}

	// given
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	// when
	swim.handleMbrStatsMsg(stats)

	// then
	assert.Equal(t, pushedMbrStatsMsg.Type, stats.Type)
	assert.Equal(t, pushedMbrStatsMsg.Id, stats.Id)
	assert.Equal(t, pushedMbrStatsMsg.Incarnation, stats.Incarnation)
	assert.Equal(t, pushedMbrStatsMsg.Address, stats.Address)

	assert.Equal(t, mm.members[MemberID{ID: "1"}].Status, Alive)
	assert.Equal(t, mm.members[MemberID{ID: "1"}].Incarnation, stats.Incarnation)
}

func TestSWIM_HandleMbrStatsMsg_AliveMsg_Error(t *testing.T) {
	// setup MbrStatsMsgStore
	var pushedMbrStatsMsg pb.MbrStatsMsg

	mbrStatsMsgStore := &MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		pushedMbrStatsMsg = mbrStatsMsg
	}

	// setup member_map
	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:          MemberID{ID: "1"},
		Status:      Dead,
		Incarnation: uint32(1),
	}

	// setup SWIM
	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
		memberMap:        mm,
		mbrStatsMsgStore: mbrStatsMsgStore,
	}

	// given
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		// error occured becuase of invalid address format
		Address: "1.2.3.4.555:5555",
	}

	// when
	swim.handleMbrStatsMsg(stats)

	// then
	assert.Equal(t, pushedMbrStatsMsg, pb.MbrStatsMsg{})

	assert.Equal(t, mm.members[MemberID{ID: "1"}].Status, Dead)
	assert.Equal(t, mm.members[MemberID{ID: "1"}].Incarnation, uint32(1))
}

func TestSWIM_HandleMbrStatsMsg_SuspectMsg(t *testing.T) {
	// setup MbrStatsMsgStore
	var pushedMbrStatsMsg pb.MbrStatsMsg

	mbrStatsMsgStore := &MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		pushedMbrStatsMsg = mbrStatsMsg
	}

	// setup member_map
	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:     MemberID{ID: "1"},
		Status: Alive,
	}

	// setup SWIM
	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
		memberMap:        mm,
		mbrStatsMsgStore: mbrStatsMsgStore,
	}

	// given
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	// when
	swim.handleMbrStatsMsg(stats)

	// then
	assert.Equal(t, pushedMbrStatsMsg.Type, stats.Type)
	assert.Equal(t, pushedMbrStatsMsg.Id, stats.Id)
	assert.Equal(t, pushedMbrStatsMsg.Incarnation, stats.Incarnation)
	assert.Equal(t, pushedMbrStatsMsg.Address, stats.Address)

	assert.Equal(t, mm.members[MemberID{ID: "1"}].Status, Suspected)
	assert.Equal(t, mm.members[MemberID{ID: "1"}].Incarnation, stats.Incarnation)
}

func TestSWIM_HandleMbrStatsMsg_SuspectMsg_Error(t *testing.T) {
	// setup MbrStatsMsgStore
	var pushedMbrStatsMsg pb.MbrStatsMsg

	mbrStatsMsgStore := &MockMbrStatsMsgStore{}
	mbrStatsMsgStore.PushFunc = func(mbrStatsMsg pb.MbrStatsMsg) {
		pushedMbrStatsMsg = mbrStatsMsg
	}

	// setup member_map
	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:          MemberID{ID: "1"},
		Status:      Alive,
		Incarnation: uint32(1),
	}

	// setup SWIM
	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
		memberMap:        mm,
		mbrStatsMsgStore: mbrStatsMsgStore,
	}

	// given
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		// error occured becuase of invalid address format
		Address: "1.2.3.4.555:5555",
	}

	// when
	swim.handleMbrStatsMsg(stats)

	// then
	assert.Equal(t, pushedMbrStatsMsg, pb.MbrStatsMsg{})

	assert.Equal(t, mm.members[MemberID{ID: "1"}].Status, Alive)
	assert.Equal(t, mm.members[MemberID{ID: "1"}].Incarnation, uint32(1))
}

func TestSWIM_handlePing(t *testing.T) {
	id := "1"

	mbrStatsMsgStore := MockMbrStatsMsgStore{}
	mbrStatsMsgStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "mbrStatsMsg_id1",
			Incarnation: 123,
			Address:     "address123",
		}, nil
	}

	mIMessageHandler := MockMessageHandler{}

	//mJ
	swim := SWIM{}

	mI := createMessageEndpoint(&mIMessageHandler, time.Second, 11143)
	mJ := createMessageEndpoint(&swim, time.Second, 11144)

	swim.messageEndpoint = mJ
	swim.mbrStatsMsgStore = mbrStatsMsgStore
	swim.member = &Member{
		ID:          MemberID{"abcde2"},
		Incarnation: 3,
	}

	go mI.Listen()
	go mJ.Listen()

	ping := pb.Message{
		Id: id,
		// address of source member
		Address: "127.0.0.1:11143",
		Payload: &pb.Message_Ping{
			Ping: &pb.Ping{},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{},
		},
	}

	defer func() {
		mI.Shutdown()
		mJ.Shutdown()
	}()

	resp, err := mI.SyncSend("127.0.0.1:11144", ping)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Payload.(*pb.Message_Ack))
	assert.Equal(t, resp.Id, id)
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Address, "address123")
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Incarnation, uint32(123))
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Id, "mbrStatsMsg_id1")
}

// This is successful scenario when target member response in time with
// ack message, then mediator sends back ack message to source member
func TestSWIM_handleIndirectPing(t *testing.T) {
	id := "1"

	mbrStatsMsgStore := MockMbrStatsMsgStore{}
	mbrStatsMsgStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "mbrStatsMsg_id1",
			Incarnation: 123,
			Address:     "address123",
		}, nil
	}

	swim := SWIM{}
	mIMessageHandler := MockMessageHandler{}
	mJMessageHandler := MockMessageHandler{}

	mK := createMessageEndpoint(&swim, time.Second, 11145)
	mI := createMessageEndpoint(&mIMessageHandler, time.Second, 11146)
	mJ := createMessageEndpoint(&mJMessageHandler, time.Second, 11147)

	swim.messageEndpoint = mK
	swim.mbrStatsMsgStore = mbrStatsMsgStore
	swim.member = &Member{
		ID:          MemberID{"abcde"},
		Incarnation: 3,
	}

	// ** m_k's handleIndirectPing is implicitly called when m_k received
	// indirect-ping message from other member **
	go mK.Listen()
	go mI.Listen()
	go mJ.Listen()

	ind := pb.Message{
		Id: id,
		// address of source member
		Address: "127.0.0.1:11146",
		Payload: &pb.Message_IndirectPing{
			IndirectPing: &pb.IndirectPing{
				// target address
				Target: "127.0.0.1:11147",
			},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{},
		},
	}

	defer func() {
		mK.Shutdown()
		mI.Shutdown()
		mJ.Shutdown()
	}()

	mJMessageHandler.handleFunc = func(msg pb.Message) {
		// check whether msg.Payload type is *pb.Message_Ping
		assert.NotNil(t, msg.Payload.(*pb.Message_Ping))
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Address, "address123")
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Incarnation, uint32(123))
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Id, "mbrStatsMsg_id1")

		ack := pb.Message{Id: msg.Id, Payload: &pb.Message_Ack{Ack: &pb.Ack{}}, PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{},
		}}
		mJ.Send("127.0.0.1:11145", ack)
	}

	resp, err := mI.SyncSend("127.0.0.1:11145", ind)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Payload.(*pb.Message_Ack))
	assert.Equal(t, resp.Id, id)
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Address, "address123")
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Incarnation, uint32(123))
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Id, "mbrStatsMsg_id1")
}

// This is NOT-successful scenario when target member DID NOT response in time
// ack message, then mediator sends back NACK message to source member
func TestSWIM_handleIndirectPing_Target_Timeout(t *testing.T) {
	id := "1"

	mbrStatsMsgStore := MockMbrStatsMsgStore{}
	mbrStatsMsgStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "mbrStatsMsg_id1",
			Incarnation: 123,
			Address:     "address123",
		}, nil
	}

	swim := SWIM{}
	mIMessageHandler := MockMessageHandler{}
	mJMessageHandler := MockMessageHandler{}

	mK := createMessageEndpoint(&swim, time.Second, 11148)
	// source should have larger send timeout, because source should give mediator
	// enough time to ping to target
	mI := createMessageEndpoint(&mIMessageHandler, time.Second*3, 11149)
	mJ := createMessageEndpoint(&mJMessageHandler, time.Second, 11150)

	swim.messageEndpoint = mK
	swim.mbrStatsMsgStore = mbrStatsMsgStore
	swim.member = &Member{
		ID:          MemberID{"abcde"},
		Incarnation: 3,
	}

	// ** m_k's handleIndirectPing is implicitly called when m_k received
	// indirect-ping message from other member **
	go mK.Listen()

	go mI.Listen()
	go mJ.Listen()

	ind := pb.Message{
		Id: id,
		// address of source member
		Address: "127.0.0.1:11149",
		Payload: &pb.Message_IndirectPing{
			IndirectPing: &pb.IndirectPing{
				// target address
				Target: "127.0.0.1:11150",
			},
		},
		PiggyBack: &pb.PiggyBack{
			MbrStatsMsg: &pb.MbrStatsMsg{},
		},
	}

	defer func() {
		mI.Shutdown()
		mK.Shutdown()
		mJ.Shutdown()
	}()

	mJMessageHandler.handleFunc = func(msg pb.Message) {
		// check whether msg.Payload type is *pb.Message_Ping
		assert.NotNil(t, msg.Payload.(*pb.Message_Ping))
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Address, "address123")
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Incarnation, uint32(123))
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg.Id, "mbrStatsMsg_id1")

		// DO NOT ANYTHING: do not response back to m
	}

	resp, err := mI.SyncSend("127.0.0.1:11148", ind)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Payload.(*pb.Message_Nack))
	assert.Equal(t, resp.Id, id)
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Address, "address123")
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Incarnation, uint32(123))
	assert.Equal(t, resp.PiggyBack.MbrStatsMsg.Id, "mbrStatsMsg_id1")
}

func TestConvMbrStatsToAliveMsg_InvalidMsgType(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{}
	// when
	msg, err := swim.convMbrStatsToAliveMsg(stats)

	// then
	assert.Equal(t, msg, AliveMessage{})
	assert.Equal(t, err, ErrInvalidMbrStatsMsgType)
}

func TestConvMbrStatsToAliveMsg_InvalidAddressFormat(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4.5:5555",
	}

	swim := &SWIM{}
	// when
	msg, err := swim.convMbrStatsToAliveMsg(stats)

	// then
	assert.Equal(t, msg, AliveMessage{})
	assert.Error(t, err)
}

func TestConvMbrStatsToAliveMsg_Success(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{}
	// when
	msg, err := swim.convMbrStatsToAliveMsg(stats)

	// then
	assert.Equal(t, msg.ID, stats.Id)
	assert.Equal(t, msg.Incarnation, stats.Incarnation)
	assert.Equal(t, msg.Addr, net.ParseIP("1.2.3.4"))
	assert.Equal(t, msg.Port, uint16(5555))
	assert.NoError(t, err)
}

func TestConvMbrStatsToSuspectMsg_InvalidMsgType(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{}
	// when
	msg, err := swim.convMbrStatsToSuspectMsg(stats)

	// then
	assert.Equal(t, msg, SuspectMessage{})
	assert.Equal(t, err, ErrInvalidMbrStatsMsgType)
}

func TestConvMbrStatsToSuspectMsg_InvalidAddressFormat(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4.555:5555",
	}

	swim := &SWIM{}
	// when
	msg, err := swim.convMbrStatsToSuspectMsg(stats)

	// then
	assert.Equal(t, msg, SuspectMessage{})
	assert.Error(t, err)
}

func TestConvMbrStatsToSuspectMsg_Success(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
	}
	// when
	msg, err := swim.convMbrStatsToSuspectMsg(stats)

	// then
	assert.Equal(t, msg.ID, stats.Id)
	assert.Equal(t, msg.Incarnation, stats.Incarnation)
	assert.Equal(t, msg.Addr, net.ParseIP("1.2.3.4"))
	assert.Equal(t, msg.Port, uint16(5555))
	assert.Equal(t, msg.ConfirmerID, "123")
	assert.NoError(t, err)
}

func TestHandleAliveMbrStats_InvalidMsgType(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{}
	// when
	result, err := swim.handleAliveMbrStatsMsg(stats)

	// then
	assert.Equal(t, result, false)
	assert.Equal(t, err, ErrInvalidMbrStatsMsgType)
}

func TestHandleAliveMbrStats_Success(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:     MemberID{ID: "1"},
		Status: Dead,
	}

	swim := &SWIM{
		memberMap: mm,
	}

	// when
	result, err := swim.handleAliveMbrStatsMsg(stats)

	// then
	assert.Equal(t, result, true)
	assert.NoError(t, err)
}

func TestHandleSuspectMbrStats_InvalidMsgType(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Alive,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	swim := &SWIM{}
	// when
	result, err := swim.handleSuspectMbrStatsMsg(stats)

	// then
	assert.Equal(t, result, false)
	assert.Equal(t, err, ErrInvalidMbrStatsMsgType)
}

func TestHandleSuspectMbrStats_Success(t *testing.T) {
	stats := &pb.MbrStatsMsg{
		Type:        pb.MbrStatsMsg_Suspect,
		Id:          "1",
		Incarnation: uint32(2),
		Address:     "1.2.3.4:5555",
	}

	mm := NewMemberMap(&SuspicionConfig{})
	mm.members[MemberID{ID: "1"}] = &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: uint32(1),
		Status:      Alive,
	}

	swim := &SWIM{
		member: &Member{
			ID: MemberID{ID: "123"},
		},
		memberMap: mm,
	}

	// when
	result, err := swim.handleSuspectMbrStatsMsg(stats)

	// then
	assert.Equal(t, true, result)
	assert.NoError(t, err)
}

func TestSWIM_indirectPing_When_Response_Success(t *testing.T) {
	stats := pb.PiggyBack{
		MbrStatsMsg: &pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "123",
			Incarnation: uint32(123),
			Address:     "pbk-addr1",
		},
	}
	pbkStore := MockMbrStatsMsgStore{}
	pbkStore.PushFunc = func(pbk pb.MbrStatsMsg) {}
	pbkStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return *stats.MbrStatsMsg, nil
	}

	messageEndpoint := MockMessageEndpoint{}
	messageEndpoint.SyncSendFunc = func(addr string, msg pb.Message, interval time.Duration) (pb.Message, error) {
		// this addr should be mediator address
		assert.Equal(t, addr, "1.2.3.4:11111")

		// msg.Address should be local-node address
		assert.Equal(t, msg.Address, "9.8.7.6:33333")
		assert.Equal(t, msg.PiggyBack.MbrStatsMsg, stats)
		assert.Equal(t, msg.Payload.(*pb.Message_IndirectPing).IndirectPing.Target, "3.4.5.6:22222")

		return pb.Message{
			Id:      "responseID",
			Address: "3.4.5.6:22222",
			Payload: &pb.Message_Ack{
				Ack: &pb.Ack{
					Payload: "ack-payload",
				},
			},
			PiggyBack: &pb.PiggyBack{
				MbrStatsMsg: &pb.MbrStatsMsg{
					Type:        pb.MbrStatsMsg_Alive,
					Id:          "555",
					Incarnation: uint32(345),
					Address:     "pbk-addr2",
				},
			},
		}, nil
	}

	member := &Member{
		ID:   MemberID{ID: "memberID"},
		Addr: net.ParseIP("1.2.3.4"),
		Port: uint16(11111),
	}
	target := &Member{
		ID:   MemberID{ID: "targetID"},
		Addr: net.ParseIP("3.4.5.6"),
		Port: uint16(22222),
	}

	indSucc := make(chan pb.Message)
	defer close(indSucc)

	config := &Config{
		BindAddress: "9.8.7.6",
		BindPort:    33333,
	}

	swim := &SWIM{}
	swim.config = config
	swim.mbrStatsMsgStore = pbkStore
	swim.messageEndpoint = messageEndpoint

	go swim.indirectPing(member, target, indSucc)

	select {
	case msg := <-indSucc:
		assert.Equal(t, msg.Id, "responseID")
		assert.Equal(t, msg.Address, "3.4.5.6:22222")
		assert.Equal(t, getAckPayload(msg), "ack-payload")
	}
}

func TestSWIM_indirectPing_When_Response_Failed(t *testing.T) {
	pbk := pb.PiggyBack{
		MbrStatsMsg: &pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "123",
			Incarnation: uint32(123),
			Address:     "pbk-addr1",
		},
	}
	pbkStore := MockMbrStatsMsgStore{}
	pbkStore.PushFunc = func(pbk pb.MbrStatsMsg) {}
	pbkStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return *pbk.MbrStatsMsg, nil
	}

	messageEndpoint := MockMessageEndpoint{}
	messageEndpoint.SyncSendFunc = func(addr string, msg pb.Message, interval time.Duration) (pb.Message, error) {
		// this addr should be mediator address
		assert.Equal(t, addr, "1.2.3.4:11111")

		// msg.Address should be local-node address
		assert.Equal(t, msg.Address, "9.8.7.6:33333")
		assert.Equal(t, msg.PiggyBack, &pbk)
		assert.Equal(t, msg.Payload.(*pb.Message_IndirectPing).IndirectPing.Target, "3.4.5.6:22222")

		// in this case, send out Timeout error
		return pb.Message{}, ErrSendTimeout
	}

	member := &Member{
		ID:   MemberID{ID: "memberID"},
		Addr: net.ParseIP("1.2.3.4"),
		Port: uint16(11111),
	}
	target := &Member{
		ID:   MemberID{ID: "targetID"},
		Addr: net.ParseIP("3.4.5.6"),
		Port: uint16(22222),
	}
	indSucc := make(chan pb.Message)

	config := &Config{
		BindAddress: "9.8.7.6",
		BindPort:    33333,
	}

	swim := &SWIM{}
	swim.config = config
	swim.mbrStatsMsgStore = pbkStore
	swim.messageEndpoint = messageEndpoint

	go swim.indirectPing(member, target, indSucc)

	T := time.NewTimer(time.Second)

	select {
	case <-indSucc:
		panic("This shouldn't be called")
	case <-T.C:
		return
	}
}

func TestSWIM_ping_When_Response_Success(t *testing.T) {
	pbk := pb.PiggyBack{
		MbrStatsMsg: &pb.MbrStatsMsg{
			Type:        pb.MbrStatsMsg_Alive,
			Id:          "123",
			Incarnation: uint32(123),
			Address:     "pbk-addr1",
		},
	}
	pbkStore := MockMbrStatsMsgStore{}
	pbkStore.PushFunc = func(pbk pb.MbrStatsMsg) {}
	pbkStore.GetFunc = func() (pb.MbrStatsMsg, error) {
		return *pbk.MbrStatsMsg, nil
	}

	messageEndpoint := MockMessageEndpoint{}
	messageEndpoint.SyncSendFunc = func(addr string, msg pb.Message, interval time.Duration) (pb.Message, error) {
		assert.Equal(t, addr, "1.2.3.4:11111")
		assert.Equal(t, msg.Address, "1.2.3.4:11111")
		assert.Equal(t, msg.PiggyBack, &pbk)
		assert.Equal(t, msg.Payload.(*pb.Message_Ping).Ping, pb.Ping{})

		return pb.Message{
			Id:      "responseID",
			Address: "3.4.5.6:22222",
			Payload: &pb.Message_Ack{
				Ack: &pb.Ack{
					Payload: "ack-payload",
				},
			},
			PiggyBack: &pb.PiggyBack{
				MbrStatsMsg: &pb.MbrStatsMsg{
					Type:        pb.MbrStatsMsg_Alive,
					Id:          "555",
					Incarnation: uint32(345),
					Address:     "pbk-addr2",
				},
			},
		}, nil
	}

	member := &Member{
		ID:   MemberID{ID: "memberID"},
		Addr: net.ParseIP("1.2.3.4"),
		Port: uint16(11111),
	}

	end := make(chan struct{})
	pingFailed := make(chan struct{})
	defer func() {
		close(end)
		close(pingFailed)
	}()

	swim := &SWIM{}
	swim.mbrStatsMsgStore = pbkStore
	swim.messageEndpoint = messageEndpoint

	go swim.ping(member, end, pingFailed)

	select {
	case <-end:
		return
	case <-pingFailed:
		panic("This shouldn't be called")
	}
}

func TestSWIM_ping_When_Response_Failed(t *testing.T) {

}

func createMessageEndpoint(messageHandler MessageHandler, sendTimeout time.Duration, port int) MessageEndpoint {
	mConfig := MessageEndpointConfig{
		EncryptionEnabled:       false,
		SendTimeout:             sendTimeout,
		CallbackCollectInterval: time.Hour,
	}

	tConfig := &PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    port,
	}

	transport, _ := NewPacketTransport(tConfig)

	m, _ := NewMessageEndpoint(mConfig, transport, messageHandler)
	return m
}
