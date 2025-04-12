package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const Eating = 10

type Philosopher struct {
	name  string
	left  *sync.Mutex
	right *sync.Mutex
}

var philosophers = []*Philosopher{
	{
		name:  "Plato",
		left:  &sync.Mutex{},
		right: &sync.Mutex{},
	},
	{
		name:  "Socrates",
		left:  &sync.Mutex{},
		right: &sync.Mutex{},
	},
	{
		name:  "Aristotle",
		left:  &sync.Mutex{},
		right: &sync.Mutex{},
	},
	{
		name:  "Democritus",
		left:  &sync.Mutex{},
		right: &sync.Mutex{},
	},
	{
		name:  "Hippocrates",
		left:  &sync.Mutex{},
		right: &sync.Mutex{},
	},
}

func (p Philosopher) Eat(wg *sync.WaitGroup, eating int) {
	defer wg.Done()

	for i := 0; i < eating; i++ {
		fmt.Printf("Philosopher %s thinking\n", p.name)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		p.left.Lock()
		p.right.Lock()
		fmt.Printf("Philosopher %s eating\n", p.name)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
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
