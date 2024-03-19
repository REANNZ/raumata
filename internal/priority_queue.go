package internal

import "container/heap"

// PriorityQueue is a heap-based priority queue
// using the standard library heap
type PriorityQueue[T any] struct {
	data minHeap[T]
}

type item[T any] struct {
	value    T
	priority int
}

type minHeap[T any] []*item[T]

func (h minHeap[T]) Len() int {
	return len(h)
}

func (h minHeap[T]) Less(i, j int) bool {
	return h[i].priority < h[j].priority
}

func (h *minHeap[T]) Swap(i, j int) {
	tmp := (*h)[i]
	(*h)[i] = (*h)[j]
	(*h)[j] = tmp
}

func (h *minHeap[T]) Push(x any) {
	item := x.(*item[T])
	*h = append(*h, item)
}

func (h *minHeap[T]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil

	*h = old[0 : n-1]

	return item
}

// Push a new element with the given priority
func (pq *PriorityQueue[T]) Push(data T, priority int) {
	heap.Push(&pq.data, &item[T]{
		value:    data,
		priority: priority,
	})
}

// Empty returns true when the queue is empty
func (pq *PriorityQueue[T]) Empty() bool {
	return len(pq.data) == 0
}

// Remove the item at the top of the queue and return it
// Returns (nil, false) if the queue is empty
func (pq *PriorityQueue[T]) Pop() (*T, bool) {
	if pq.Empty() {
		return nil, false
	} else {
		item := heap.Pop(&pq.data).(*item[T])
		return &item.value, true
	}
}
