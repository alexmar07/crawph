package queue

import (
	"fmt"
	"sync"
)

type Queue struct {
	items []string
	mu    sync.Mutex
	cond  *sync.Cond
}

func NewQueue() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Enqueue(item string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *Queue) Dequeue(shouldTerminate *bool) (string, bool) {

	q.mu.Lock()
	defer q.mu.Unlock()

	fmt.Println("Sto rimuovendo un elemento dalla coda")

	if len(q.items) == 0 && !*shouldTerminate {
		q.cond.Wait()
	}

	if len(q.items) == 0 && *shouldTerminate {
		return "", false
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item, true
}

func (q *Queue) Size() int {

	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

func (q *Queue) BroadcastTermination() {
	q.cond.Broadcast()
}

func (q *Queue) IsEmpty() bool {
	return q.Size() == 0
}
