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

import "time"

// Suspicion manages the suspect timer and helps to accelerate the timeout
// as member self got more independent confirmations that a target member is suspect.
type Suspicion struct {

	// n is the number of independent confirmations we've seen.
	n uint32

	// k is the maximum number of independent confirmation's we'd like to see
	// this value is for making timer to drive @min value
	k uint32

	// min is the minimum timer value
	min time.Duration

	// max is the maximum timer value
	max time.Duration

	// start captures the timestamp when the suspect began the timer. This value is used
	// for calculating durations
	start time.Time

	timer *time.Timer

	// confirmations is a map for saving unique member id whose member also confirmed that
	// given suspected node is suspect. This prevents double counting, the same confirmer will
	// excluded since we might get our own suspicion message gossiped back to us
	confirmations map[MemberID]struct{}
}

// TODO: NewSuspicion returns a timer started with the max value, and according to
// TODO: Lifeguard L2 (Dynamic Suspicion timeout) each unique confirmation will drive the timer
// TODO: to min value
func NewSuspicion(confirmer MemberID, k int, min time.Duration, max time.Duration) *Suspicion {
	return &Suspicion{}
}

// TODO: Confirm register new member who also determined the given suspected member as suspect.
// TODO: This returns true if this confirmer is new, and false if it was a duplicate information
// TODO: or if we've got enough confirmations to hit the value of timer to minimum
func (s *Suspicion) Confirm(confirmer MemberID) bool {
	return false
}

// TODO: remainingSuspicionTime helps to calculate the remaining time to wait before suspected node
// TODO: considered as a dead. The return value could be negative, in the case of return value
// TODO: is negative, immediately fire the timer
func calcRemainingSuspicionTime(n, k uint32, elapsed, min, max time.Duration) time.Duration {
	return 0
}
