package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Cache struct {
	m      sync.RWMutex
	values map[string]uint
}

func (c *Cache) Get(key string) uint {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.values[key]
}

func (c *Cache) Set(key string, value uint) {
	c.m.Lock()
	defer c.m.Unlock()
	c.values[key] = value
}

func main() {
	var c = Cache{values: make(map[string]uint)}
	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				v := c.Get("key1")
				fmt.Printf("[%2d] %v\n", i, v)
				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
			}
		}()
	}
	go func() {
		for {
			n := c.Get("key1")
			c.Set("key1", n+1)
		}
	}()
	time.Sleep(time.Minute)
}
