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
	"errors"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var ErrEmptyMemberID = errors.New("MemberID is empty")

// Status of members
type Status int

const (
	// Unknown status of a member
	Unknown Status = iota

	// Alive status
	Alive

	// Suspicious status of whether a member is dead or not
	Suspected

	// Dead status
	Dead
)

type MemberID struct {
	ID string
}

// Struct of Member
type Member struct {
	// Id of member
	ID MemberID

	// Ip address
	Addr net.IP

	// Port
	Port uint16

	// Current member status from my point of view
	Status Status

	// Time last status change happened
	LastStatusChange time.Time

	// Incarnation helps to keep the most fresh information about member status in the system
	// which tells that suspect member confirming that it is alive, and only when suspect
	// got suspicion message, that member can increments incarnation
	Incarnation uint32

	// Suspicion manages the suspect timer and helps to accelerate the timeout
	// as member self got more independent confirmations that a target member is suspect.
	Suspicion *Suspicion
}

type MemberMessage struct {
	ID          string
	Addr        net.IP
	Port        uint16
	Incarnation uint32
}

type AliveMessage struct {
	MemberMessage
}

// Suspect message struct
type SuspectMessage struct {
	MemberMessage
}

type ConfirmMessage struct {
	MemberMessage
}

// Convert member addr and port to string
func (m *Member) Address() string {
	return net.JoinHostPort(m.Addr.String(), strconv.Itoa(int(m.Port)))
}

// Get
func (m *Member) GetID() MemberID {
	return m.ID
}

type MemberMap struct {
	lock    sync.RWMutex
	members map[MemberID]*Member
}

func NewMemberMap() *MemberMap {
	return &MemberMap{
		members: make(map[MemberID]*Member),
		lock:    sync.RWMutex{},
	}
}

// Select K random member (length of returning member can be lower than k).
func (m *MemberMap) SelectKRandomMemberID(k int) []Member {

	m.lock.Lock()
	defer m.lock.Unlock()

	// Filter non-alive member
	members := make([]Member, 0)
	for _, member := range m.members {
		if member.Status == Alive {
			members = append(members, *member)
		}
	}

	selectedMembers := make([]Member, 0)

	// Select K random members
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	for i := 0; i < k; i++ {

		// When K is larger than members
		if len(members) == 0 {
			break
		}

		index := r.Intn(len(members))
		selectedMembers = append(selectedMembers, members[index])
		members = append(members[:index], members[index+1:]...)
	}

	return selectedMembers
}

func (m *MemberMap) GetMembers() []Member {
	m.lock.Lock()
	defer m.lock.Unlock()

	members := make([]Member, 0)
	for _, member := range m.members {
		members = append(members, *member)
	}

	return members
}

// Override will override member status based on incarnation number and status.
//
// Overriding rules are following...
//
// 1. {Alive Ml, inc=i} overrides
//      - {Suspect Ml, inc=j}, i>j
//      - {Alive Ml, inc=j}, i>j
//
// 2. {Suspect Ml, inc=i} overrides
//      - {Suspect Ml, inc=j}, i>j
//      - {Alive Ml, inc=j}, i>=j
//
// 3. {Dead Ml, inc=i} overrides
//      - {Suspect Ml, inc=j}, i>j
//      - {Alive Ml, inc=j}, i>j
func override(newMem Member, existingMem Member) Member {

	// Check i, j
	// All cases if incarnation number is higher then another member,
	// override it.
	if newMem.Incarnation > existingMem.Incarnation {
		return newMem
	}

	// member2 == member1 case suspect status can override alive
	if newMem.Incarnation == existingMem.Incarnation {
		if newMem.Status == Suspected && existingMem.Status == Alive {
			return newMem
		}
	}

	return existingMem
}

// Process Alive message
//
// Return 2 parameter bool, error bool means Changed in MemberList
//
// 1. If aliveMessage Id is empty return false and ErrEmptyMemberID
// 2. if aliveMessage Id is not in MemberList then Create Member and Add MemberList
// 3. if aliveMessage Id is in MemberList and existingMember's Incarnation is bigger than AliveMessage's Incarnation Than return false, ErrIncarnation
// 4. if aliveMessage Id is in MemberList and AliveMessage's Incarnation is bigger than existingMember's Incarnation Than Update Member and Return true, nil
func (m *MemberMap) Alive(aliveMessage AliveMessage) (bool, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	// Check whether Member Id empty
	if aliveMessage.ID == "" {
		return false, ErrEmptyMemberID
	}
	// Check whether it is already exist
	existingMem, ok := m.members[MemberID{aliveMessage.ID}]

	// if Member is not exist in MemberList
	if !ok {
		m.members[MemberID{aliveMessage.ID}] = createMember(aliveMessage.MemberMessage, Alive)
		return true, nil
	}
	// Check incarnation
	if aliveMessage.Incarnation <= existingMem.Incarnation {
		return false, nil
	}

	// update Member
	existingMem.Status = Alive
	existingMem.Incarnation = aliveMessage.Incarnation
	existingMem.LastStatusChange = time.Now()
	return true, nil
}

// Handle Confirm Message

func (m *MemberMap) Confirm(confirmMessage ConfirmMessage) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Check whether Member ID in Confirm message is empty or not
	if confirmMessage.ID == "" {
		return false, ErrEmptyMemberID
	}

	// Get the existing member from the member list
	mID := MemberID{confirmMessage.ID}
	member, exist := m.members[mID]

	// Check that the member exists
	if !exist {
		return false, nil
	}

	// {Dead Ml, inc=i} overrides
	// - {Suspect Ml, inc=j}, any j
	// - {Alive Ml, inc=j}, any j
	status := member.Status
	if status == Alive || status == Suspected {
		member.Status = Dead
		member.LastStatusChange = time.Now()
		member.Incarnation = confirmMessage.Incarnation
		member.Suspicion = nil

		return true, nil
	}

	return false, nil
}

func createMember(message MemberMessage, status Status) *Member {
	return &Member{
		ID:               MemberID{message.ID},
		Addr:             message.Addr,
		Port:             message.Port,
		Status:           status,
		LastStatusChange: time.Now(),
		Incarnation:      message.Incarnation,
	}
}

// Delete all dead node,
// Reset waiting list,
func (m *MemberMap) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	// delete dead status node
	for k, member := range m.members {
		if member.Status == Dead {
			delete(m.members, k)
		}
	}
}
