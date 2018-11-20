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

func TestMemberMap_Confirm_Internal(t *testing.T) {
	// given
	m := NewMemberMap()

	// when : alive
	m.members[MemberID{ID: "1"}] = &Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Alive,
	}

	c, e := m.Confirm(ConfirmMessage{
		MemberMessage: MemberMessage{
			ID:          "1",
			Incarnation: 1,
		},
	})

	// then
	assert.True(t, c)
	assert.NoError(t, e)
	assert.Equal(t, m.members[MemberID{"1"}].Status, Dead)

	// when : suspected
	m.members[MemberID{ID: "2"}] = &Member{
		ID: MemberID{
			ID: "2",
		},
		Incarnation: 1,
		Status:      Suspected,
	}

	c, e = m.Confirm(ConfirmMessage{
		MemberMessage: MemberMessage{
			ID:          "2",
			Incarnation: 2,
		},
	})

	// then
	assert.True(t, c)
	assert.NoError(t, e)
	assert.Equal(t, m.members[MemberID{"2"}].Status, Dead)

	// when : dead
	m.members[MemberID{ID: "3"}] = &Member{
		ID: MemberID{
			ID: "3",
		},
		Incarnation: 1,
		Status:      Dead,
	}

	c, e = m.Confirm(ConfirmMessage{
		MemberMessage: MemberMessage{
			ID:          "3",
			Incarnation: 3,
		},
	})

	// then
	assert.False(t, c)
	assert.NoError(t, e)
	assert.Equal(t, m.members[MemberID{"3"}].Status, Dead)
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

// Override will override member status based on incarnation number and status.
// Overriding rules are following...

// 1. {Alive Ml, inc=i} overrides {Alive Ml, inc=j}, i>j
func TestMember_override_alive_alive(t *testing.T) {

	m1 := Member{
		Status:      Alive,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, latter member(existing member) will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Alive,
		Incarnation: 1,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m3, m4), m4)
}

// 2. {Alive Ml, inc=i} overrides {Suspect Ml, inc=j}, i>j
func TestMember_override_alive_suspect(t *testing.T) {

	m1 := Member{
		Status:      Alive,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Suspected,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, latter member(existing member) will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Alive,
		Incarnation: 1,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Suspected,
		Incarnation: 1,
	}

	assert.Equal(t, override(m3, m4), m4)
}

// 3. {Suspect Ml, inc=i} overrides {Suspect Ml, inc=j}, i>j
func TestMember_override_suspect_alive(t *testing.T) {

	m1 := Member{
		Status:      Suspected,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, latter member(existing member) will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Alive,
		Incarnation: 1,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Suspected,
		Incarnation: 1,
	}

	assert.Equal(t, override(m3, m4), m4)
}

// 4. {Suspect Ml, inc=i} overrides {Alive Ml, inc=j}, i>=j
func TestMember_override_suspect_suspect(t *testing.T) {

	m1 := Member{
		Status:      Suspected,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, Suspect will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Suspected,
		Incarnation: 1,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m3, m4), m3)
}

// 5.  {Dead Ml, inc=i} overrides {Suspect Ml, inc=j}, any j
func TestMember_override_dead_suspect(t *testing.T) {

	m1 := Member{
		Status:      Dead,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Suspected,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, latter member(existing member) will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Dead,
		Incarnation: 1,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Suspected,
		Incarnation: 1,
	}

	assert.Equal(t, override(m3, m4), m4)
}

// 6. {Dead Ml, inc=i} overrides {Alive Ml, inc=j}, any j
func TestMember_override_dead_alive(t *testing.T) {
	m1 := Member{
		Status:      Dead,
		Incarnation: 2,
	}

	m2 := Member{
		Status:      Alive,
		Incarnation: 1,
	}

	assert.Equal(t, override(m1, m2), m1)
	assert.Equal(t, override(m2, m1), m1)

	// if both have same incarnation, latter member(existing member) will be returned
	m3 := Member{
		ID: MemberID{
			ID: "1",
		},
		Status:      Dead,
		Incarnation: 2,
	}

	m4 := Member{
		ID: MemberID{
			ID: "2",
		},
		Status:      Alive,
		Incarnation: 2,
	}

	assert.Equal(t, override(m3, m4), m4)
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
