package queue

import (
	"sync"
)

// Item represents a URL to crawl with its depth from the seed.
type Item struct {
	URL   string
	Depth int
}

type Queue struct {
	items      []Item
	mu         sync.Mutex
	cond       *sync.Cond
	terminated bool
}

func NewQueue() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Enqueue(item Item) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *Queue) Dequeue() (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 && !q.terminated {
		q.cond.Wait()
	}

	if len(q.items) == 0 {
		return Item{}, false
	}

	item := q.items[0]
	q.items[0] = Item{} // clear reference for GC
	q.items = q.items[1:]

	// Compact to reclaim memory when most items have been dequeued
	if cap(q.items) > 256 && len(q.items) < cap(q.items)/4 {
		compacted := make([]Item, len(q.items))
		copy(compacted, q.items)
		q.items = compacted
	}

	return item, true
}

func (q *Queue) Terminate() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.terminated = true
	q.cond.Broadcast()
}

func (q *Queue) IsTerminated() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.terminated
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *Queue) IsEmpty() bool {
	return q.Size() == 0
}
