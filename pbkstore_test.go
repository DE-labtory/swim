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

func NewMbrStatsMsg(id string) pb.MbrStatsMsg {
	return pb.MbrStatsMsg{
		Id:          id,
		Incarnation: 1,
		Address:     "123",
	}
}

func TestPiggyBackPriorityStore_Push(t *testing.T) {

	//given
	mbrStatsMsgStore := swim.NewPriorityMbrStatsMsgStore(3)

	//when
	mbrStatsMsgStore.Push(NewMbrStatsMsg("1"))
	mbrStatsMsgStore.Push(NewMbrStatsMsg("2"))

	//then
	assert.Equal(t, mbrStatsMsgStore.Len(), 2)
}

func TestPiggyBackPriorityStore_Get(t *testing.T) {

	// given
	mbrStatsMsgStore := swim.NewPriorityMbrStatsMsgStore(3)
	mbrStatsMsgStore.Push(NewMbrStatsMsg("1"))

	// when
	// When get function called, mbrStatsMsgstore internally increases the priority.
	// First time
	mbrStatsMsgData, err := mbrStatsMsgStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, mbrStatsMsgData, NewMbrStatsMsg("1"))
	assert.Equal(t, mbrStatsMsgStore.Len(), 1)

	// when
	// Second time
	mbrStatsMsgData_2, err := mbrStatsMsgStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, mbrStatsMsgData_2, NewMbrStatsMsg("1"))
	assert.Equal(t, mbrStatsMsgStore.Len(), 1)

	// when
	// Third time
	mbrStatsMsgData_3, err := mbrStatsMsgStore.Get()
	assert.Nil(t, err)

	// then
	assert.Equal(t, mbrStatsMsgData_3, NewMbrStatsMsg("1"))

	// This time data was queried three times, so it was deleted
	assert.Equal(t, mbrStatsMsgStore.Len(), 0)

	// when
	_, err = mbrStatsMsgStore.Get()
	assert.Equal(t, err, swim.ErrStoreEmpty)
}

func TestPiggyBackPriorityStore_Len(t *testing.T) {

	// given
	mbrStatsMsgStore := swim.NewPriorityMbrStatsMsgStore(3)
	for i := 0; i < 3; i++ {
		mbrStatsMsgStore.Push(NewMbrStatsMsg(strconv.Itoa(i)))
	}

	// when && then
	assert.Equal(t, mbrStatsMsgStore.Len(), 3)
}

func TestPiggyBackPriorityStore_IsEmpty(t *testing.T) {

	// given
	mbrStatsMsgStore := swim.NewPriorityMbrStatsMsgStore(3)

	// when && then
	assert.True(t, mbrStatsMsgStore.IsEmpty())
}
