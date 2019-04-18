package swim

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

/*
 * https://github.com/hashicorp/memberlist
 *
 * This Source Code Form is subject to the terms of the
 * Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
 * with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

import (
	"sync"
	"time"
)

// Awareness manages health of the local node. This related to Lifeguard L1
// "self-awareness" concept. Expect to receive replies to probe messages we sent
// Start timeouts low, increase in response to absence of replies
type Awareness struct {
	sync.RWMutex

	// Max is the upper threshold for the factor that increase the timeout value
	// (the score will be constrained from 0 <= score < Max)
	max int

	// score is the current awareness score. Lower values are healthier and
	// zero is the minimum value
	score int
}

func NewAwareness(max int) *Awareness {
	return &Awareness{
		max:   max,
		score: 0,
	}
}

func (a *Awareness) GetHealthScore() int {
	a.RLock()
	defer a.RUnlock()
	return a.score
}

// ApplyDelta with given delta applies it to the score in thread-safe manner
// score must be bound from 0 to Max value
func (a *Awareness) ApplyDelta(delta int) {

	a.RLock()
	defer a.RUnlock()

	a.score += delta
	if a.score < 0 {
		a.score = 0
	} else if a.score > (a.max - 1) {
		a.score = a.max - 1
	}
}

// ScaleTimeout takes the given duration and scales it based on the current score.
// Less healthyness will lead to longer timeouts.
func (a *Awareness) ScaleTimeout(timeout time.Duration) time.Duration {
	a.RLock()
	defer a.RUnlock()

	return timeout * (time.Duration(a.score) + 1)
}
