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
	"testing"
	"time"

	"github.com/DE-labtory/swim/pb"

	"github.com/stretchr/testify/assert"
)

type MockMessageEndpoint struct {
	ListenFunc   func()
	SyncSendFunc func(addr string, msg pb.Message, interval time.Duration) (pb.Message, error)
	SendFunc     func(addr string, msg pb.Message) error
	ShutdownFunc func()
}

func (m MockMessageEndpoint) Listen() {
	m.ListenFunc()
}
func (m MockMessageEndpoint) SyncSend(addr string, msg pb.Message) (pb.Message, error) {
	return m.SyncSendFunc(addr, msg, interval)
}
func (m MockMessageEndpoint) Send(addr string, msg pb.Message) error {
	return m.SendFunc(addr, msg)
}
func (m MockMessageEndpoint) Shutdown() {
	m.ShutdownFunc()
}

func TestMemberMap_Alive_Internal(t *testing.T) {
	m := NewMemberMap(&SuspicionConfig{})

	//when member 1
	//when Incarnation Number of Parameter of Alive Func is bigger than member in memberList
	m.members[MemberID{ID: "1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Suspected,
	}
	isChanged, err := m.Alive(AliveMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: 2,
		},
	})

	//Then
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isChanged)
	assert.Equal(t, Alive, m.members[MemberID{ID: "1"}].Status)

	//when
	m.members[MemberID{ID: "1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 5,
		Status:      Suspected,
	}

	isChanged2, err := m.Alive(AliveMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: 4,
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, false, isChanged2)
	assert.Equal(t, Suspected, m.members[MemberID{ID: "1"}].Status)

}

func TestMemberMap_GetMembers_Internal(t *testing.T) {

	// given
	m := NewMemberMap(&SuspicionConfig{})

	m.members[MemberID{ID: "1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Alive,
	}
	m.members[MemberID{ID: "2"}] = &Member{
		ID: MemberID{
			ID: "2",
		},
		Incarnation: 1,
		Status:      Alive,
	}
	m.members[MemberID{ID: "3"}] = &Member{
		ID: MemberID{
			ID: "3",
		},
		Incarnation: 1,
		Status:      Alive,
	}

	// when
	members := m.GetMembers()

	// then
	assert.Equal(t, len(m.GetMembers()), 3)
	assert.Contains(t, members, Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Alive,
	})
	assert.Contains(t, members, Member{
		ID: MemberID{
			ID: "2",
		},
		Incarnation: 1,
		Status:      Alive,
	})
	assert.Contains(t, members, Member{
		ID: MemberID{
			ID: "3",
		},
		Incarnation: 1,
		Status:      Alive,
	})
}

func TestMemberMap_createMember_Internal(t *testing.T) {

	// given Member Message
	memberMessage := MemberMessage{
		ID:          "1",
		Port:        1111,
		Incarnation: 1,
	}

	// Update Member
	member := createMember(memberMessage, Alive)

	//then
	assert.Equal(t, "1", member.ID.ID)
	assert.Equal(t, uint32(1), member.Incarnation)
	assert.Equal(t, Alive, member.Status)
}

func TestMemberMap_Reset_Internal_Test(t *testing.T) {

	// given
	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{"1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Status: Alive,
	}
	m.members[MemberID{"2"}] = &Member{
		ID: MemberID{
			ID: "2",
		},
		Status: Alive,
	}
	m.members[MemberID{"3"}] = &Member{
		ID: MemberID{
			ID: "3",
		},
		Status: Dead,
	}

	// when
	m.Reset()

	// then
	members := m.GetMembers()
	assert.Equal(t, len(m.GetMembers()), 2)
	assert.Contains(t, members, Member{
		ID: MemberID{
			ID: "1",
		},
		Status: Alive,
	})
	assert.Contains(t, members, Member{
		ID: MemberID{
			ID: "2",
		},
		Status: Alive,
	})
}

func TestMemberMap_SelectKRandomMember(t *testing.T) {

	// given
	m := NewMemberMap(&SuspicionConfig{})

	m.members[MemberID{ID: "1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Alive,
	}
	m.members[MemberID{ID: "2"}] = &Member{
		ID: MemberID{
			ID: "2",
		},
		Incarnation: 1,
		Status:      Alive,
	}
	m.members[MemberID{ID: "3"}] = &Member{
		ID: MemberID{
			ID: "3",
		},
		Incarnation: 1,
		Status:      Alive,
	}

	// case 1: randomly select all members
	members := m.GetMembers()
	rMembers := m.SelectKRandomMemberID(3)
	for i := 0; i < len(members); i++ {

		assert.True(t, checkExist(rMembers, members[i]))
	}

	// case 2: when k is larger then length of members
	assert.Equal(t, len(m.SelectKRandomMemberID(5)), 3)
}

func TestMemberMap_Suspect_MessageId_Empty(t *testing.T) {
	// setup member map
	m := NewMemberMap(&SuspicionConfig{})

	msg := SuspectMessage{
		MemberMessage: MemberMessage{
			ID: "",
		},
	}

	res, err := m.Suspect(msg)

	assert.Equal(t, res, false)
	assert.Equal(t, err, ErrEmptyMemberID)
}

func TestMemberMap_Suspect_When_Member_Not_Exist(t *testing.T) {
	// setup member map
	m := NewMemberMap(&SuspicionConfig{})

	msg := SuspectMessage{
		MemberMessage: MemberMessage{
			ID: "12341234",
		},
	}

	res, err := m.Suspect(msg)

	assert.Equal(t, res, false)
	assert.NoError(t, err)
}

func TestMemberMap_Suspect_With_Smaller_Incarnation(t *testing.T) {
	// setup member map
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Alive,
	}

	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{ID: "1"}] = member1

	msg := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(1),
		},
	}

	res, err := m.Suspect(msg)

	assert.Equal(t, res, false)
	assert.Equal(t, err, nil)
}

func TestMemberMap_Suspect_When_Member_Alive(t *testing.T) {
	// setup member map
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Alive,
	}
	member2 := &Member{
		ID:          MemberID{ID: "2"},
		Incarnation: 3,
		Status:      Alive,
	}

	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{ID: "1"}] = member1
	m.members[MemberID{ID: "2"}] = member2

	// Suspect message with equal incarnation
	msg1 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(3),
		},
		ConfirmerID: "IAMCONFIRMER",
	}

	res, err := m.Suspect(msg1)

	assert.Equal(t, res, true)
	assert.Equal(t, err, nil)
	assert.Equal(t, member1.Incarnation, msg1.Incarnation)
	assert.Equal(t, member1.Status, Suspected)
	assert.Equal(t, member1.Suspicion.confirmations, map[MemberID]struct{}{
		MemberID{ID: "IAMCONFIRMER"}: {},
	})
	assert.NotNil(t, member1.LastStatusChange)

	// Suspect message with large incarnation
	msg2 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "2",
			Incarnation: uint32(5),
		},
		ConfirmerID: "IAMCONFIRMER22",
	}

	res, err = m.Suspect(msg2)

	assert.Equal(t, res, true)
	assert.Equal(t, err, nil)
	assert.Equal(t, member2.Incarnation, msg2.Incarnation)
	assert.Equal(t, member2.Status, Suspected)
	assert.Equal(t, member2.Suspicion.confirmations, map[MemberID]struct{}{
		MemberID{ID: "IAMCONFIRMER22"}: {},
	})
	assert.NotNil(t, member2.LastStatusChange)
}

// In the case of member already have suspicion, then do not create new suspicion
// only Confirm
func TestMemberMap_Suspect_When_Member_Suspect(t *testing.T) {
	// setup member map
	suspicion, _ := NewSuspicion(MemberID{ID: "ALREADY-HAVE-CONFIRMER"}, 10, time.Second, time.Hour, func() {})
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Suspected,
		Suspicion:   suspicion,
	}

	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{ID: "1"}] = member1

	// Suspect message with equal incarnation
	msg1 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(5),
		},
		ConfirmerID: "IAMCONFIRMER",
	}

	res, err := m.Suspect(msg1)

	assert.Equal(t, res, true)
	assert.Equal(t, err, nil)
	assert.Equal(t, member1.Incarnation, msg1.Incarnation)
	assert.Equal(t, member1.Status, Suspected)
	assert.Equal(t, member1.Suspicion.confirmations, map[MemberID]struct{}{
		MemberID{ID: "ALREADY-HAVE-CONFIRMER"}: {},
		MemberID{ID: "IAMCONFIRMER"}:           {},
	})

	// When member already have suspicion not update timestamp
	// only update suspicion timeout, in this case member1 have no initial timestamp
	// so assert with isZero
	assert.True(t, member1.LastStatusChange.IsZero())
}

func TestMemberMap_Suspect_When_Member_Suspect_Without_Suspicion(t *testing.T) {
	// setup member map
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Suspected,
	}

	m := NewMemberMap(&SuspicionConfig{
		k:   1000,
		min: time.Hour,
		max: time.Hour * 8,
	})
	m.members[MemberID{ID: "1"}] = member1

	// Suspect message with equal incarnation
	msg1 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(5),
		},
		ConfirmerID: "IAMCONFIRMER",
	}

	res, err := m.Suspect(msg1)

	assert.Equal(t, res, true)
	assert.Equal(t, err, nil)
	assert.Equal(t, member1.Incarnation, msg1.Incarnation)
	assert.Equal(t, member1.Status, Suspected)
	assert.Equal(t, member1.Suspicion.confirmations, map[MemberID]struct{}{
		MemberID{ID: "IAMCONFIRMER"}: {},
	})

	// When member already have suspicion not update timestamp
	// only update suspicion timeout, in this case member1 have no initial timestamp
	// so assert with isZero
	assert.True(t, member1.LastStatusChange.IsZero())
}

func TestMemberMap_Suspect_When_Dead(t *testing.T) {
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Dead,
	}

	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{ID: "1"}] = member1

	msg1 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(5),
		},
		ConfirmerID: "IAMCONFIRMER",
	}

	res, err := m.Suspect(msg1)

	assert.Equal(t, res, false)
	assert.Equal(t, err, nil)
}

func TestMemberMap_Suspect_When_Unknown(t *testing.T) {
	member1 := &Member{
		ID:          MemberID{ID: "1"},
		Incarnation: 3,
		Status:      Unknown,
	}

	m := NewMemberMap(&SuspicionConfig{})
	m.members[MemberID{ID: "1"}] = member1

	msg1 := SuspectMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: uint32(5),
		},
		ConfirmerID: "IAMCONFIRMER",
	}

	res, err := m.Suspect(msg1)

	assert.Equal(t, res, false)
	assert.Equal(t, err, ErrMemberUnknownState)
}

func checkExist(members []Member, m Member) bool {
	for _, member := range members {
		if member.ID == m.ID {
			return true
		}
	}
	return false
}
