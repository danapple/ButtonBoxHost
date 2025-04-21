package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func heartbeat(outChan chan<- byte, wg *sync.WaitGroup, ctx context.Context) {
	defer func() { recover() }()

	log.Printf("Heartbeat starting\n")

	for {
		select {
		case <-ctx.Done():
			{
				log.Printf("Heartbeat done\n")
				wg.Done()
				return
			}
		case <-time.After(200 * time.Millisecond):
			{
				outChan <- 255
			}
		}
	}
}
