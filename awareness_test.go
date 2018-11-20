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
	"time"

	"github.com/DE-labtory/swim"
	"github.com/stretchr/testify/assert"
)

func TestAwareness_ApplyDelta(t *testing.T) {
	// given
	tests := map[string]struct {
		max   int
		delta int
		score int
	}{
		"negative score": {
			max:   5,
			delta: -3,
			score: 0,
		},
		"max score": {
			max:   5,
			delta: 10,
			score: 4,
		},
		"normal": {
			max:   5,
			delta: 1,
			score: 1,
		},
	}

	for testName, test := range tests {
		t.Logf("Running test case %s", testName)

		// when
		a := swim.NewAwareness(test.max)
		a.ApplyDelta(test.delta)

		// then
		assert.Equal(t, a.GetHealthScore(), test.score)
	}
}

func TestAwareness_GetHealthScore(t *testing.T) {
	// given
	a := swim.NewAwareness(3)

	// when
	// then
	assert.Equal(t, a.GetHealthScore(), 0)

	// when
	a.ApplyDelta(1)

	// then
	assert.Equal(t, a.GetHealthScore(), 1)
}

func TestAwareness_ScaleTimeout(t *testing.T) {

	// given
	a := swim.NewAwareness(3)

	// when
	// then
	// score = 0
	assert.Equal(t, a.ScaleTimeout(100), time.Duration(100))

	// when
	// score = 1
	a.ApplyDelta(1)

	// then
	assert.Equal(t, a.ScaleTimeout(100), time.Duration(200))
}
