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
	"net"
	"strconv"
	"sync"
)

// status of members
type Status int

const (

	// unknown status of a member
	Unknown Status = iota

	// alive status
	Alive

	// suspicious status of whether a member is dead or not
	Suspected

	// dead status
	Dead
)

type MemberID struct {
	ID string
}

// struct of Member
type Member struct {

	// ip address
	Addr net.IP

	// port
	Port uint16

	// status: Unknown, alive, suspected, dead
	Status Status
}

// convert member addr and port to string
func (m *Member) Address() string {
	return net.JoinHostPort(m.Addr.String(), strconv.Itoa(int(m.Port)))
}

type MemberMap struct {
	sync.RWMutex
	members map[MemberID]*Member
}

// select K random member from members
func (m *MemberMap) selectKRandomMember(k int) []Member {
	return nil
}
