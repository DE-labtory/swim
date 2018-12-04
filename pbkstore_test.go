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
	"strconv"
	"testing"

	"github.com/DE-labtory/swim"
	"github.com/DE-labtory/swim/pb"
	"github.com/stretchr/testify/assert"
)

func NewPiggyBack(id string) pb.PiggyBack {
	return pb.PiggyBack{
		Member: &pb.Member{
			Id:          id,
			Incarnation: 1,
			Address:     "123",
		},
	}
}

func TestPiggyBackPriorityStore_Push(t *testing.T) {

	//given
	pbkStore := swim.NewPriorityPBStore(3)

	//when
	pbkStore.Push(NewPiggyBack("1"))
	pbkStore.Push(NewPiggyBack("2"))

	//then
	assert.Equal(t, pbkStore.Len(), 2)
}

func TestPiggyBackPriorityStore_Get(t *testing.T) {

	// given
	pbkStore := swim.NewPriorityPBStore(3)
	pbkStore.Push(NewPiggyBack("1"))

	// when
	// When get function called, pbkstore internally increases the priority.
	// First time
	pbkData, err := pbkStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, pbkData, NewPiggyBack("1"))
	assert.Equal(t, pbkStore.Len(), 1)

	// when
	// Second time
	pbkData_2, err := pbkStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, pbkData_2, NewPiggyBack("1"))
	assert.Equal(t, pbkStore.Len(), 1)

	// when
	// Third time
	pbkData_3, err := pbkStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, pbkData_3, NewPiggyBack("1"))

	// This time data was queried three times, so it was deleted
	assert.Equal(t, pbkStore.Len(), 0)

	// when
	_, err = pbkStore.Get()
	assert.Equal(t, err, swim.ErrStoreEmpty)
}

func TestPiggyBackPriorityStore_Len(t *testing.T) {

	// given
	pbkStore := swim.NewPriorityPBStore(3)
	for i := 0; i < 3; i++ {
		pbkStore.Push(NewPiggyBack(strconv.Itoa(i)))
	}

	// when && then
	assert.Equal(t, pbkStore.Len(), 3)
}

func TestPiggyBackPriorityStore_IsEmpty(t *testing.T) {

	// given
	pbkStore := swim.NewPriorityPBStore(3)

	// when && then
	assert.True(t, pbkStore.IsEmpty())
}
