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
	"net"
	"strconv"
)

type MemberID struct{
	ID string
}

type Member struct{
	Addr net.IP
	Port uint16
}

// convert member addr and port to string
func (m *Member) Address() string{
	return net.JoinHostPort(m.Addr.String(), strconv.Itoa(int(m.Port)))
}

type MemberMap struct{
	sync.RWMutex
	members map[MemberID]*Member
}

// select K random member from members
func (m *Member) selectKRandomMember(k int) []Member{
	return nil
}

