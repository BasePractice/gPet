package main

import "time"

type Data any
type DataChannelSend chan<- Data
type DataChannelRecv <-chan Data

func main() {
	var c = time.After(5 * time.Second)
	select {
	case <-c:
	}
}
