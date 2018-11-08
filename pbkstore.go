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

import "sync"

// PiggyBackData consists of data and localCount.
// The localCount is incremented each time the data is queried.
// The lower the localCount, the faster the queried.
// The data has a member status, such as alive, suspected or dead.
type piggyBackData struct {
	data       []byte
	localCount int
}

// PiggyBackStore store a piggyback data. When we ping, ack or indirect-ack,
// get PiggyBackData from the PiggyBackStore and send it with ping, ack or indirect-ack.
type PiggyBackStore interface {
	Len() int
	Push(b []byte) error
	Get() ([]byte, error)
	IsEmpty() bool
}

// PiggyBackPriorityStore stores piggyback data and returns data with small local count.
type PiggyBackPriorityStore struct {
	list    []piggyBackData
	maxSize int
	lock    sync.RWMutex
}

func NewPiggyBackPriorityStore(maxSize int) *PiggyBackPriorityStore {
	return &PiggyBackPriorityStore{
		list:    make([]piggyBackData, 0),
		maxSize: maxSize,
		lock:    sync.RWMutex{},
	}
}

// Return current size of data
func (p *PiggyBackPriorityStore) Len() int {
	p.lock.Lock()
	defer p.lock.Unlock()

	return len(p.list)
}

// Enqueue []byte into list.
// Initially, set the local count to zero.
// If the queue size is max, delete the data with the highest localcount and insert it.
func (p *PiggyBackPriorityStore) Push(b []byte) error {
	return nil
}

// Return the piggyback data with the smallest local count in the list,
// increment the local count and sort it again, not delete the data.
func (p *PiggyBackPriorityStore) Get() ([]byte, error) {
	return nil, nil
}

func (p *PiggyBackPriorityStore) IsEmpty() bool {
	return true
}
