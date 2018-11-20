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

	"github.com/stretchr/testify/assert"
)

func TestSuspicion_calcRemainingSuspicionTime(t *testing.T) {

	// given
	tests := map[string]struct {
		n        int32
		k        int32
		elapsed  time.Duration
		min      time.Duration
		max      time.Duration
		expected time.Duration
	}{
		"case 1": {0, 3, 0, 2 * time.Second, 30 * time.Second, 30 * time.Second},
		"case 2": {1, 3, 2 * time.Second, 2 * time.Second, 30 * time.Second, 14 * time.Second},
		"case 3": {2, 3, 3 * time.Second, 2 * time.Second, 30 * time.Second, 4810 * time.Millisecond},
		"case 4": {3, 3, 4 * time.Second, 2 * time.Second, 30 * time.Second, -2 * time.Second},
		"case 5": {4, 3, 5 * time.Second, 2 * time.Second, 30 * time.Second, -3 * time.Second},
		"case 6": {5, 3, 10 * time.Second, 2 * time.Second, 30 * time.Second, -8 * time.Second},
	}

	for testName, test := range tests {
		t.Logf("Running test case %s", testName)

		// when
		remaining := calcRemainingSuspicionTime(test.n, test.k, test.elapsed, test.min, test.max)

		// then
		assert.Equal(t, remaining, test.expected)
	}
}

func TestNewSuspicion(t *testing.T) {

	// check nil timeout handler
	_, err := NewSuspicion(MemberID{ID: "me"}, 3, 2*time.Second, 30*time.Second, nil)
	assert.Error(t, err)

	timeoutHandler := func() {}

	_, err = NewSuspicion(MemberID{ID: "me"}, 3, 2*time.Second, 30*time.Second, timeoutHandler)
	assert.NoError(t, err)
}

func TestSuspicion_Received_One_Confirm(t *testing.T) {

	// given
	wg := sync.WaitGroup{}
	wg.Add(1)

	var start time.Time

	// given
	timeoutHandler := func() {
		// then
		// The duration was roughly calculated.
		assert.WithinDuration(t, start, time.Now(), 12*time.Second)
		wg.Done()
	}

	start = time.Now()

	_, err := NewSuspicion(MemberID{ID: "me"}, 3, 2*time.Second, 10*time.Second, timeoutHandler)
	assert.NoError(t, err)

	//
	wg.Wait()
}

func TestSuspicion_Received_Two_Confirm(t *testing.T) {

	// given
	wg := sync.WaitGroup{}
	wg.Add(1)

	var start time.Time

	// given
	timeoutHandler := func() {
		// then
		// The duration was roughly calculated.
		assert.WithinDuration(t, start, time.Now(), 8*time.Second)
		wg.Done()
	}

	start = time.Now()

	s, err := NewSuspicion(MemberID{ID: "me"}, 3, 2*time.Second, 10*time.Second, timeoutHandler)
	assert.NoError(t, err)

	// when received two confirm
	assert.True(t, s.Confirm(MemberID{ID: "hello"}))

	//
	wg.Wait()
}

func TestSuspicion_Received_Three_Confirm(t *testing.T) {

	// given
	wg := sync.WaitGroup{}
	wg.Add(1)

	var start time.Time

	// given
	timeoutHandler := func() {
		// then
		// The duration was roughly calculated.
		assert.WithinDuration(t, start, time.Now(), 6*time.Second)
		wg.Done()
	}

	start = time.Now()

	s, err := NewSuspicion(MemberID{ID: "me"}, 3, 2*time.Second, 10*time.Second, timeoutHandler)
	assert.NoError(t, err)

	// when received three confirm
	assert.True(t, s.Confirm(MemberID{ID: "hello"}))
	assert.True(t, s.Confirm(MemberID{ID: "hello1"}))

	//
	wg.Wait()
}
