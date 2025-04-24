package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var argsWithoutProg = os.Args[1:]
	var portName = argsWithoutProg[0]

	fmt.Printf("Starting ButtonBox host processor on serial port '%v'\n", portName)

	outChan := make(chan byte, 10)
	inChan := make(chan byte, 10)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)
	go portManager(portName, inChan, outChan, ctx, &wg)
	go restApi(outChan, &wg)

	waitForSignal()

	cancel()
	shutdownRestApi()
	close(outChan)
	close(inChan)

	wg.Wait()
	log.Printf("Done")

}

func waitForSignal() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		signal := <-signals

		switch signal {
		case syscall.SIGINT:
			return
		case syscall.SIGTERM:
			return
		}
	}
}
