package main

import (
	"context"
	"log"
	"sync"
	"time"

	"go.bug.st/serial"
)

func portManager(portName string, inChan chan byte, outChan chan byte, ctx context.Context, wg *sync.WaitGroup) {
	var errorShown = false

	for {

		port, err := openPort(portName)
		if port == nil {
			if !errorShown {
				log.Printf("Failed opening port '%v': '%v'", portName, err)
				errorShown = true
			}
			//log.Printf("portManager selecting\n")

			select {
			case <-ctx.Done():
				{
					log.Printf("Stopping portManager\n")
					wg.Done()
					return
				}
			case <-time.After(500 * time.Millisecond):
				{
				}
			}
			continue
		}
		if errorShown {
			log.Printf("Opened port '%v'", portName)
			errorShown = false
		}

		time.Sleep(1 * time.Second)
		portContext, portCancel := context.WithCancel(context.Background())

		waitChan := make(chan bool)
		var portManagerWaitGroup sync.WaitGroup
		portManagerWaitGroup.Add(4)
		go writer(outChan, port, &portManagerWaitGroup, portContext)
		go reader(inChan, port, &portManagerWaitGroup, portCancel)
		go heartbeat(outChan, &portManagerWaitGroup, portContext)
		go buttonProcessor(inChan, outChan, &portManagerWaitGroup, portContext)
		go func() {
			portManagerWaitGroup.Wait()
			waitChan <- true
		}()
		select {
		case <-ctx.Done():
			{
				log.Printf("portManager canceling jobs on port\n")
				portCancel()
				port.Close()
				portManagerWaitGroup.Wait()
				log.Printf("portManager done\n")
				wg.Done()
				return
			}
		case <-waitChan:
			{
				log.Printf("portManager looping\n")
			}
		}
	}
}

func writer(outChan <-chan byte, port serial.Port, wg *sync.WaitGroup, ctx context.Context) {
	buff := make([]byte, 1)
	defer wg.Done()

	for {
		select {
		case toWrite, more := <-outChan:
			{
				if !more {
					return
				}
				buff[0] = toWrite
				port.Write(buff)
				//log.Printf("Writing byte '%v'\n", toWrite)
			}
		case <-ctx.Done():
			{
				log.Printf("Writer done")
				return
			}
		}
	}
}

func reader(inChan chan<- byte, port serial.Port, wg *sync.WaitGroup, portCancel context.CancelFunc) {
	buff := make([]byte, 100)
	defer wg.Done()

	for {
		n, err := port.Read(buff)
		if n == 0 {
			log.Printf("0 bytes to read, done reading\n")
			portCancel()
			return
		}
		if err != nil {
			//log.Printf("Got error '%v', done reading\n", err)
			return
		}
		for i := 0; i < n; i++ {
			element := buff[i]
			//log.Printf("Received '%v'", element)
			inChan <- element
		}
	}
}

func openPort(portName string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return port, nil
}
