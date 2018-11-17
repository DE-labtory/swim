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
	"container/heap"
	"errors"
	"sync"

	"github.com/DE-labtory/swim/pb"
)

var ErrStoreEmpty = errors.New("empty store")
var ErrPopInvalidType = errors.New("pop invalid typed item")

const (
	InitialPriority = 0
)

// PiggyBackStore store a piggyback data. When swim ping, ack or indirect-ack,
// swim need to get PiggyBackData from the PiggyBackStore and send it with ping, ack or indirect-ack.

// The priority is incremented each time when the data is queried.
// The lower the priority, the faster the queried.
// The data has a member status, such as alive, suspected or dead.

type PBkStore interface {
	Len() int
	Push(pbk pb.PiggyBack)
	Get() (pb.PiggyBack, error)
	IsEmpty() bool
}

// PiggyBackStore stores piggyback data in the priority queue and returns data with smallest local count.
type PriorityPBStore struct {
	q             PriorityQueue
	maxLocalCount int
	lock          sync.RWMutex
}

// macLocalCount is the max priority value
func NewPriorityPBStore(maxLocalCount int) *PriorityPBStore {
	return &PriorityPBStore{
		q:             make(PriorityQueue, 0),
		maxLocalCount: maxLocalCount,
		lock:          sync.RWMutex{},
	}
}

// Return current size of data
func (p *PriorityPBStore) Len() int {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.q.Len()
}

// Initially, set the local count to zero.
// If the queue size is max, delete the data with the highest localcount and insert it.
func (p *PriorityPBStore) Push(pbk pb.PiggyBack) {
	p.lock.Lock()
	defer p.lock.Unlock()

	item := &Item{
		value:    pbk,
		priority: InitialPriority,
	}

	heap.Push(&p.q, item)
}

// Return the piggyback data with the smallest local count in the list,
// increment the local count and sort it again, not delete the data.
func (p *PriorityPBStore) Get() (pb.PiggyBack, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// Check empty
	if len(p.q) == 0 {
		return pb.PiggyBack{}, ErrStoreEmpty
	}

	// Pop from queue
	item := heap.Pop(&p.q).(*Item)
	pbk, ok := item.value.(pb.PiggyBack)
	if !ok {
		return pb.PiggyBack{}, ErrPopInvalidType
	}

	// If an item has been retrieved by maxPriority, remove it.
	// If not, push it again after increment priority
	item.priority = item.priority + 1
	if item.priority < p.maxLocalCount {
		heap.Push(&p.q, item)
	}

	return pbk, nil
}

func (p *PriorityPBStore) IsEmpty() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.q) == 0 {
		return true
	}
	return false
}
