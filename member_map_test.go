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

func TestMemberMap_Alive(t *testing.T) {

	// given
	m := swim.NewMemberMap()

	// when
	// Add new member
	isChanged, err := m.Alive(swim.AliveMessage{
		MemberMessage: swim.MemberMessage{
			ID:          "1",
			Incarnation: 1,
		},
	})

	// then
	assert.Nil(t, err)
	assert.Equal(t, true, isChanged)

	// when
	// Add existing member
	isChanged, err = m.Alive(swim.AliveMessage{
		MemberMessage: swim.MemberMessage{
			ID:          "1",
			Incarnation: 2,
		},
	})

	// then
	assert.Nil(t, err)
	assert.Equal(t, true, isChanged)
}
