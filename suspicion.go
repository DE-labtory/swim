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
	"errors"
	"math"
	"sync/atomic"
	"time"
)

// Suspicion manages the suspect timer and helps to accelerate the timeout
// as member self got more independent confirmations that a target member is suspect.
type Suspicion struct {

	// n is the number of independent confirmations we've seen.
	n int32

	// K is the maximum number of independent confirmation's we'd like to see
	// this value is for making timer to drive @Min value
	k int32

	// Min is the minimum timer value
	min time.Duration

	// Max is the maximum timer value
	max time.Duration

	// start captures the timestamp when the suspect began the timer. This value is used
	// for calculating durations
	startTime time.Time

	// timer is the underlying timer that implements the timeout.
	timer *time.Timer

	// timeoutHandler is the function to call when the timer expires. We hold on to this
	// because there are cases where we call it directly.
	timeoutHandler func()

	// confirmations is a map for saving unique member id whose member also confirmed that
	// given suspected node is suspect. This prevents double counting, the same confirmer will
	// excluded since we might get our own suspicion message gossiped back to us
	confirmations map[MemberID]struct{}
}

// NewSuspicion returns a timer started with the Max value, and according to
// Lifeguard L2 (Dynamic Suspicion timeout) each unique confirmation will drive the timer
// to Min value
func NewSuspicion(confirmer MemberID, k int, min time.Duration, max time.Duration, timeoutHandler func()) (*Suspicion, error) {

	if timeoutHandler == nil {
		return nil, errors.New("timeout handler can not be nil")
	}

	s := &Suspicion{
		k:             int32(k),
		min:           min,
		max:           max,
		confirmations: make(map[MemberID]struct{}),
	}

	// Exclude the from node from any confirmations.
	s.confirmations[confirmer] = struct{}{}

	// Pass the number of confirmations into the timeout function for
	// easy telemetry.
	s.timeoutHandler = timeoutHandler

	// If there aren't any confirmations to be made then take the Min
	// time from the start.
	timeout := max
	if k < 1 {
		timeout = min
	}
	s.timer = time.AfterFunc(timeout, s.timeoutHandler)

	// Capture the start time right after starting the timer above so
	// we should always err on the side of a little longer timeout if
	// there's any preemption that separates this and the step above.

	s.startTime = time.Now()
	return s, nil
}

// Confirm register new member who also determined the given suspected member as suspect.
// This returns true if this confirmer is new, and false if it was a duplicate information
// or if we've got enough confirmations to hit the value of timer to minimum
func (s *Suspicion) Confirm(confirmer MemberID) bool {
	// If we've got enough confirmations then stop accepting them.
	if atomic.LoadInt32(&s.n) >= s.k {
		return false
	}

	// Only allow one confirmation from each possible peer.
	if _, ok := s.confirmations[confirmer]; ok {
		return false
	}
	s.confirmations[confirmer] = struct{}{}

	// Compute the new timeout given the current number of confirmations and
	// adjust the timer. If the timeout becomes negative *and* we can cleanly
	// stop the timer then we will call the timeout function directly from
	// here.
	n := atomic.AddInt32(&s.n, 1)
	elapsed := time.Since(s.startTime)
	remaining := calcRemainingSuspicionTime(n, s.k, elapsed, s.min, s.max)
	if s.timer.Stop() {
		if remaining > 0 {
			s.timer.Reset(remaining)
		} else {
			go s.timeoutHandler()
		}
	}
	return true
}

// CalcRemainingSuspicionTime takes the state variables of the suspicion timer and
// calculates the remaining time to wait before considering a node dead. The
// return value can be negative, so be prepared to fire the timer immediately in
// that case.
func calcRemainingSuspicionTime(n, k int32, elapsed, min, max time.Duration) time.Duration {
	frac := math.Log(float64(n)+1.0) / math.Log(float64(k)+1.0)
	raw := max.Seconds() - frac*(max.Seconds()-min.Seconds())
	timeout := time.Duration(math.Floor(1000.0*raw)) * time.Millisecond
	if timeout < min {
		timeout = min
	}

	// We have to take into account the amount of time that has passed so
	// far, so we get the right overall timeout.
	return timeout - elapsed
}
