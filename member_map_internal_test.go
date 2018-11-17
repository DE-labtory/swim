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

	"github.com/stretchr/testify/assert"
)

func TestMemberMap_Alive_Internal(t *testing.T) {
	m := NewMemberMap()

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
	m := NewMemberMap()

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
	m := NewMemberMap()
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
	m := NewMemberMap()

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

func checkExist(members []Member, m Member) bool {
	for _, member := range members {
		if member.ID == m.ID {
			return true
		}
	}
	return false
}
