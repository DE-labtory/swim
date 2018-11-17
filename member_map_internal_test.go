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

	//when member 1 suspectedMember
	//when Incarnation Number of Parameter of Alive Func is bigger than member in memberList
	m.members[MemberID{ID:"1"}] = Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      Suspected,
	}
	isChanged, err := m.Alive(Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 2,
		Status:      Alive,
	})

	//Then
	assert.Nil(t, err)
	assert.Equal(t, true, isChanged)
	assert.Equal(t, Alive, m.members[MemberID{ID:"1"}].Status)

	//when
	m.members[MemberID{ID:"1"}] = Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 5,
		Status:      Suspected,
	}

	isChanged2, err := m.Alive(Member{
		ID: MemberID{
			ID: "1",
		},
		Incarnation: 4,
		Status:      Alive,
	})

	assert.Equal(t, ErrIncarnation, err)
	assert.Equal(t, false, isChanged2)
	assert.Equal(t, Suspected, m.members[MemberID{ID:"1"}].Status)

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

func TestMemberMap_reset(t *testing.T) {
	memberMaps := map[MemberID]Member{
		MemberID{ID: "1"}: {ID: MemberID{ID: "1"}},
		MemberID{ID: "2"}: {ID: MemberID{ID: "2"}},
		MemberID{ID: "3"}: {ID: MemberID{ID: "3"}},
	}

	waitingMembersID := resetWaitingMembersID(memberMaps)
	for _, member := range memberMaps {
		assert.Contains(t, waitingMembersID, member)
	}
}
