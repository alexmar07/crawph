package queue

import (
	"sync"
	"testing"
	"time"
)

func TestEnqueueDequeue(t *testing.T) {
	q := NewQueue()
	q.Enqueue(Item{URL: "https://a.com", Depth: 0})
	q.Enqueue(Item{URL: "https://b.com", Depth: 1})

	item, ok := q.Dequeue()
	if !ok || item.URL != "https://a.com" || item.Depth != 0 {
		t.Errorf("expected (a.com, 0, true), got (%s, %d, %v)", item.URL, item.Depth, ok)
	}

	item, ok = q.Dequeue()
	if !ok || item.URL != "https://b.com" || item.Depth != 1 {
		t.Errorf("expected (b.com, 1, true), got (%s, %d, %v)", item.URL, item.Depth, ok)
	}
}

func TestDequeueBlocksUntilEnqueue(t *testing.T) {
	q := NewQueue()
	done := make(chan string)

	go func() {
		item, _ := q.Dequeue()
		done <- item.URL
	}()

	time.Sleep(50 * time.Millisecond)
	q.Enqueue(Item{URL: "delayed", Depth: 0})

	select {
	case url := <-done:
		if url != "delayed" {
			t.Errorf("expected delayed, got %s", url)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Dequeue did not unblock after Enqueue")
	}
}

func TestTerminationUnblocksDequeue(t *testing.T) {
	q := NewQueue()
	done := make(chan bool)

	go func() {
		_, ok := q.Dequeue()
		done <- ok
	}()

	time.Sleep(50 * time.Millisecond)
	q.Terminate()

	select {
	case ok := <-done:
		if ok {
			t.Error("expected ok=false after termination")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Dequeue did not unblock after Terminate")
	}
}

func TestConcurrentEnqueueDequeue(t *testing.T) {
	q := NewQueue()
	var wg sync.WaitGroup
	results := make(chan string, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				q.Enqueue(Item{URL: "item", Depth: 0})
			}
		}()
	}

	wg.Wait()

	if q.Size() != 100 {
		t.Errorf("expected 100 items, got %d", q.Size())
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, ok := q.Dequeue()
			if ok {
				results <- item.URL
			}
		}()
	}

	wg.Wait()
	close(results)

	count := 0
	for range results {
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 dequeued items, got %d", count)
	}
}

func TestIsEmpty(t *testing.T) {
	q := NewQueue()
	if !q.IsEmpty() {
		t.Error("new queue should be empty")
	}
	q.Enqueue(Item{URL: "a", Depth: 0})
	if q.IsEmpty() {
		t.Error("queue with item should not be empty")
	}
}
