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

	"github.com/it-chain/iLogger"
)

var ErrEmptyMemberID = errors.New("MemberID is empty")
var ErrMemberUnknownStates = errors.New("unknown status")

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

type SuspicionConfig struct {
	// k is the maximum number of independent confirmation's we'd like to see
	// this value is for making timer to drive @min value
	k int

	// min is the minimum timer value
	min time.Duration

	// max is the maximum timer value
	max time.Duration
}

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

// Convert member addr and port to string
func (m *Member) Address() string {
	return net.JoinHostPort(m.Addr.String(), strconv.Itoa(int(m.Port)))
}

// Get
func (m *Member) GetID() MemberID {
	return m.ID
}

type MemberMap struct {
	lock            sync.RWMutex
	members         map[MemberID]*Member
	suspicionConfig *SuspicionConfig
}

// Suspect message struct
type SuspectMessage struct {
	MemberMessage
	ConfirmerID string
}

func NewMemberMap(config *SuspicionConfig) *MemberMap {
	return &MemberMap{
		members:         make(map[MemberID]*Member),
		lock:            sync.RWMutex{},
		suspicionConfig: config,
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

// Suspect handle suspectMessage, if this function update member states
// return true otherwise false
func (m *MemberMap) Suspect(suspectMessage SuspectMessage) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// if member id is empty, return empty memberID err
	if suspectMessage.ID == "" {
		return false, ErrEmptyMemberID
	}

	config := m.suspicionConfig
	member, exist := m.members[MemberID{suspectMessage.ID}]

	// If member map does not have that member, ignore suspect message.
	if !exist {
		return false, nil
	}

	// if exist member incarnation is bigger than message, drop the message
	if member.Incarnation > suspectMessage.Incarnation {
		return false, nil
	}

	// if suspect message's incarnation is large or equal than that of member's incarnation
	// and if member is alive then turn this member's state into suspect with timeout handler
	if member.Status == Alive {
		suspicion, err := NewSuspicion(MemberID{suspectMessage.ConfirmerID}, config.k, config.min, config.max, getSuspicionCallback(m, member))
		if err != nil {
			iLogger.Panic(&iLogger.Fields{"err": err}, "error while create suspicion")
		}

		m.updateToSuspect(member, suspectMessage.Incarnation, suspicion, true)
		return true, nil
	}

	// if suspect message's incarnation is larger than that of member's incarnation
	// and if member is suspect reduce timeout according to Lifeguard Dynamic Suspicion Timeout
	// concept
	if member.Incarnation < suspectMessage.Incarnation && member.Status == Suspected {
		if member.Suspicion == nil {
			suspicion, err := NewSuspicion(MemberID{suspectMessage.ConfirmerID}, config.k, config.min, config.max, getSuspicionCallback(m, member))
			if err != nil {
				iLogger.Panic(&iLogger.Fields{"err": err}, "error while create suspicion")
			}

			m.updateToSuspect(member, suspectMessage.Incarnation, suspicion, false)
			return true, nil
		} else {
			member.Suspicion.Confirm(MemberID{suspectMessage.ConfirmerID})
			return true, nil
		}
	}

	iLogger.Error(nil, "unknown Status")
	return false, ErrMemberUnknownStates
}

func (m *MemberMap) updateToSuspect(member *Member, incarnation uint32, suspicion *Suspicion, updateTimestamp bool) {
	if updateTimestamp {
		member.LastStatusChange = time.Now()
	}
	member.Status = Suspected
	member.Incarnation = incarnation
	member.Suspicion = suspicion
}

func (m *MemberMap) updateToDead(member *Member) {
	member.Status = Dead
	member.LastStatusChange = time.Now()
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

func getSuspicionCallback(memberMap *MemberMap, member *Member) func() {
	return func() {
		memberMap.lock.Lock()
		defer memberMap.lock.Unlock()

		member, exist := memberMap.members[member.ID]
		if !exist {
			iLogger.Error(nil, "member is not found in callback suspicion")
			return
		}

		memberMap.updateToDead(member)
	}
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
