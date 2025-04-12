package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Queue[T any] struct {
	index  int
	values []T
	c      *sync.Cond
}

func NewQueue[T any](capacity int) Queue[T] {
	return Queue[T]{
		values: make([]T, capacity),
		index:  -1,
		c:      sync.NewCond(&sync.Mutex{}),
	}
}

func (q *Queue[T]) Write(element T) {
	q.c.L.Lock()
	defer q.c.L.Unlock()
	for q.index == len(q.values) {
		q.c.Wait()
	}
	if q.index == -1 {
		q.index = 0
	}
	q.values[q.index] = element
	if q.index+1 <= len(q.values) {
		q.index++
	}
	q.c.Signal()
}

func (q *Queue[T]) Read() T {
	q.c.L.Lock()
	defer q.c.L.Unlock()
	for q.index < 0 {
		q.c.Wait()
	}
	q.index = q.index - 1
	var value T
	if q.index < 0 {
		value = q.values[0]
	} else {
		value = q.values[q.index]
	}
	q.c.Signal()
	return value
}

type Data struct {
	who   int
	value int64
}

func main() {
	var queue = NewQueue[Data](300)
	var n atomic.Int64
	var index atomic.Int64
	for i := 0; i < 20; i++ {
		go func() {
			for {
				value := queue.Read()
				fmt.Printf("[%010d] %02d:%02d:%d\n", index.Add(1), i, value.who, value.value)
			}
		}()
	}
	for i := 0; i < 1; i++ {
		go func() {
			for {
				queue.Write(Data{who: i, value: n.Add(1)})
			}
		}()
	}
	time.Sleep(time.Minute)
}
