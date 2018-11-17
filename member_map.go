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
var ErrIncarnation = errors.New("New Member Incarnation Number is smaller than Member in MemberList")

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

// Suspect message struct
type SuspectMessage struct {
	MemberMessage
}

type AliveMessage struct {
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
	members map[MemberID]Member

	// This is for selecting k random member based on round-robin
	waitingMembers []Member
}

func NewMemberMap() *MemberMap {
	return &MemberMap{
		members:        make(map[MemberID]Member),
		waitingMembers: make([]Member, 0),
		lock:           sync.RWMutex{},
	}
}

// Select K random memberID from waitingMembers(length of returning member can be lower than k).
// ** WaitingMembers are shuffled every time when members are updated **, so just returning first K item in waitingMembers is same as
// selecting k random membersID.
func (m *MemberMap) SelectKRandomMemberID(k int) []Member {
	m.lock.Lock()
	defer m.lock.Unlock()

	// If length of members is lower then k,
	// return current waitingMembers and Reset waitingMembers.
	if len(m.waitingMembers) < k {
		kMembers := m.waitingMembers
		defer func() { m.waitingMembers = resetWaitingMembersID(m.members) }()
		return kMembers
	}

	// Remove first k membersID.
	kMembers := m.waitingMembers[:k]
	m.waitingMembers = m.waitingMembers[k:len(m.waitingMembers)]

	if len(m.waitingMembers) == 0 {
		m.waitingMembers = resetWaitingMembersID(m.members)
	}

	return kMembers
}

func (m *MemberMap) GetMembers() []Member {
	m.lock.Lock()
	defer m.lock.Unlock()

	members := make([]Member, 0)
	for _, v := range m.members {
		members = append(members, v)
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
// Return 2 parameter bool, error bool means Changed in MemberList
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
		m.members[MemberID{aliveMessage.ID}] = updateMember(aliveMessage.MemberMessage, Alive)
		return true, nil
	}
	// Check incarnation
	if aliveMessage.Incarnation <= existingMem.Incarnation {
		return false, ErrIncarnation
	}

	// update Member
	m.members[MemberID{aliveMessage.ID}] = updateMember(aliveMessage.MemberMessage, Alive)
	return true, nil
}

func updateMember(message MemberMessage, status Status) Member {
	return Member{
		ID:               MemberID{message.ID},
		Addr:             message.Addr,
		Port:             message.Port,
		Status:           status,
		LastStatusChange: time.Now(),
		Incarnation:      message.Incarnation,
	}
}

// Remove member and update waitingMembers
// todo
func (m *MemberMap) RemoveMember(member Member) error {
	return nil
}

// This function will be called when memberMap updated
// Create waiting memberID List and shuffle
func resetWaitingMembersID(memberMap map[MemberID]Member) []Member {

	// Convert Map to List
	waitingMembersID := make([]Member, 0)
	for _, member := range memberMap {
		waitingMembersID = append(waitingMembersID, member)
	}

	// Shuffle the list
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(waitingMembersID); n > 0; n-- {
		randIndex := r.Intn(n)
		waitingMembersID[n-1], waitingMembersID[randIndex] = waitingMembersID[randIndex], waitingMembersID[n-1]
	}

	return waitingMembersID
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

	m.waitingMembers = resetWaitingMembersID(m.members)
}
