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

package swim_test

import (
	"testing"

	"github.com/DE-labtory/swim"
	"github.com/stretchr/testify/assert"
)

func TestMemberMap_AddMember(t *testing.T) {

	// given
	m := swim.NewMemberMap()

	// when
	// Add new member
	err := m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "1",
		},
		Incarnation: 1,
		Status:      swim.Alive,
	})

	// then
	assert.Nil(t, err)

	// when
	// Add existing member
	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "1",
		},
		Incarnation: 2,
		Status:      swim.Alive,
	})

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(m.GetMembers()), 1)

	// when
	// Add new member
	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "3",
		},
	})

	// then
	assert.Equal(t, len(m.GetMembers()), 2)
}

func TestMemberMap_GetMembers(t *testing.T) {

	// given
	m := swim.NewMemberMap()
	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "1",
		},
	})
	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "2",
		},
	})
	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "3",
		},
	})

	// when
	members := m.GetMembers()

	// then
	assert.Equal(t, len(m.GetMembers()), 3)
	assert.Contains(t, members, swim.Member{
		ID: swim.MemberID{
			ID: "1",
		},
	})
	assert.Contains(t, members, swim.Member{
		ID: swim.MemberID{
			ID: "2",
		},
	})
	assert.Contains(t, members, swim.Member{
		ID: swim.MemberID{
			ID: "3",
		},
	})
}

func TestMemberMap_SelectKRandomMember(t *testing.T) {

	// given
	m := swim.NewMemberMap()
	err := m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "1",
		},
	})
	assert.NoError(t, err)

	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "2",
		},
	})
	assert.NoError(t, err)

	m.AddMember(swim.Member{
		ID: swim.MemberID{
			ID: "3",
		},
	})
	assert.NoError(t, err)

	members := m.GetMembers()
	//
	// case 1: get member one by one
	for i := 0; i < len(m.GetMembers()); i++ {

		// Get one random member which exists in members
		assert.True(t, checkExist(members, m.SelectKRandomMemberID(1)[0].ID))
	}

	// case 2: when k is larger then length of members
	assert.Equal(t, len(m.SelectKRandomMemberID(10)), 3)

	// case 3
	kMember := append(m.SelectKRandomMemberID(2), m.SelectKRandomMemberID(1)...)
	assert.Equal(t, len(kMember), 3)
	for _, member := range kMember {
		assert.True(t, checkExist(members, member.ID))
	}
}

func checkExist(members []swim.Member, memberID swim.MemberID) bool {
	for _, members := range members {
		if members.ID == memberID {
			return true
		}
	}

	return false
}
