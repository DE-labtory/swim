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

func TestSuspicion_Confirm(t *testing.T) {

	// given
	timeoutHandler := func() {}

	s, err := swim.NewSuspicion(swim.MemberID{ID: "me"}, 3, 2*time.Second, 5*time.Second, timeoutHandler)
	assert.NoError(t, err)

	f := s.Confirm(swim.MemberID{ID: "test"})
	assert.True(t, f)

	// when
	// check duplicated member confirm
	f = s.Confirm(swim.MemberID{ID: "test"})

	// then
	assert.False(t, f)
}
