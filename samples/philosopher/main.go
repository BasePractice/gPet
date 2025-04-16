package main

import (
	"fmt"
	"sync"
	"time"
)

const Eating = 10

type Philosopher struct {
	name  string
	left  *sync.Mutex
	right *sync.Mutex
}

var forks = make([]sync.Mutex, 5)
var philosophers = []*Philosopher{
	{
		name:  "Plato",
		left:  &forks[0],
		right: &forks[4],
	},
	{
		name:  "Socrates",
		left:  &forks[1],
		right: &forks[0],
	},
	{
		name:  "Aristotle",
		left:  &forks[2],
		right: &forks[1],
	},
	{
		name:  "Democritus",
		left:  &forks[3],
		right: &forks[2],
	},
	{
		name:  "Hippocrates",
		left:  &forks[4],
		right: &forks[3],
	},
}

func (p Philosopher) Eat(wg *sync.WaitGroup, eating int) {
	defer wg.Done()

	for i := 0; i < eating; i++ {
		fmt.Printf("[%02d] Philosopher %s thinking\n", i, p.name)
	lbl:
		for {
			p.left.Lock()
			time.Sleep(time.Second)
			for !p.right.TryLock() {
				p.left.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue lbl
			}
			break
		}
		fmt.Printf("[%02d] Philosopher %s eating\n", i, p.name)
		p.right.Unlock()
		p.left.Unlock()
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(len(philosophers))
	for _, philosopher := range philosophers {
		go philosopher.Eat(&wg, Eating)
	}
	wg.Wait()
	fmt.Println("All philosopher eat")
}
