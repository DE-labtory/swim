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
	"testing"

	"github.com/stretchr/testify/assert"
)

// Set PriorityQueue
func setUpHeap(n int) PriorityQueue {

	pq := make(PriorityQueue, 0)
	for i := 0; i < n; i++ {
		pq.Push(&Item{
			value:    i,
			priority: i,
			index:    i,
		})
	}

	return pq
}

func TestPriorityQueue_Len(t *testing.T) {
	heapSize := 3

	q := setUpHeap(heapSize)
	assert.Equal(t, q.Len(), heapSize)
}

func TestPriorityQueue_Less(t *testing.T) {
	heapSize := 3

	// q item priority will be [0, 1, 2]
	q := setUpHeap(heapSize)

	// 0 will be lower then 1
	assert.True(t, q.Less(0, 1))

	// 1 will be lower then 2
	assert.True(t, q.Less(1, 2))

	// 0 will be lower then 2
	assert.True(t, q.Less(0, 1))
}

func TestPriorityQueue_Pop(t *testing.T) {
	heapSize := 3

	// q item value will be [0, 1, 2]
	// Pop function will pop item from tail
	q := setUpHeap(heapSize)

	assert.Equal(t, 2, q.Pop().(*Item).value)
	assert.Equal(t, 1, q.Pop().(*Item).value)
	assert.Equal(t, 0, q.Pop().(*Item).value)
}

func TestPriorityQueue_Push(t *testing.T) {
	heapSize := 3

	// q item value will be [0, 1, 2]
	q := setUpHeap(heapSize)

	// Push function will push item to last
	q.Push(&Item{
		value:    3,
		priority: 3,
		index:    3,
	})

	assert.Equal(t, 3, q.Pop().(*Item).value)
	assert.Equal(t, 2, q.Pop().(*Item).value)
}

func TestPriorityQueue_Swap(t *testing.T) {
	heapSize := 3
	q := setUpHeap(heapSize)

	// Swap function swap two item(index i, index j)
	q.Swap(0, 1)

	assert.Equal(t, q[0].value, 1)
	assert.Equal(t, q[1].value, 0)
}
