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

func TestNewAwareness(t *testing.T) {

	// given
	tests := map[string]struct {
		input int
		max   int
		score int
	}{
		"max-3": {
			input: 3,
			max:   3,
			score: 0,
		},
		"max-5": {
			input: 5,
			max:   5,
			score: 0,
		},
		"max-10": {
			input: 10,
			max:   10,
			score: 0,
		},
	}

	for testName, test := range tests {
		t.Logf("Running test case %s", testName)

		// when
		a := NewAwareness(test.input)

		// then
		assert.Equal(t, a.max, test.max)
		assert.Equal(t, a.score, test.score)
	}
}
