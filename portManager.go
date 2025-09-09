package main

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.bug.st/serial"
)

type PortManager struct {
	logger    *zerolog.Logger
	portName  string
	inChan    chan byte
	outChan   chan byte
	waitGroup *sync.WaitGroup
}

func NewPortManager(logger *zerolog.Logger, portName string, inChan chan byte, outChan chan byte, waitGroup *sync.WaitGroup) *PortManager {
	return &PortManager{
		logger:    ptr(logger.With().Str(LogKey.Module, "PortManager").Logger()),
		portName:  portName,
		waitGroup: waitGroup,
		inChan:    inChan,
		outChan:   outChan,
	}
}

func (pm *PortManager) Start(ctx context.Context) {
	pm.waitGroup.Add(1)
	go pm.loop(ctx)
}

func (pm *PortManager) loop(ctx context.Context) {
	var errorShown = false
	defer pm.waitGroup.Done()

	for {
		port, err := openPort(pm.portName)
		if port == nil {
			if !errorShown {
				pm.logger.Error().Msgf("Failed opening port '%v'", err)
				errorShown = true
			}

			select {
			case <-ctx.Done():
				{
					pm.logger.Info().Msg("Stopping")
					return
				}
			case <-time.After(500 * time.Millisecond):
				{
				}
			}
			continue
		}
		if errorShown {
			pm.logger.Info().Msg("Opened port")
			errorShown = false
		}

		time.Sleep(1 * time.Second)
		portContext, portCancel := context.WithCancel(context.Background())

		waitChan := make(chan bool)
		var portManagerWaitGroup sync.WaitGroup
		portManagerWaitGroup.Add(3)
		go pm.writer(port, &portManagerWaitGroup, portContext)
		go pm.reader(port, &portManagerWaitGroup, portCancel)
		go pm.heartbeat(portContext)
		go func() {
			portManagerWaitGroup.Wait()
			waitChan <- true
		}()
		select {
		case <-ctx.Done():
			{
				pm.logger.Info().Msg("portManager canceling jobs")
				portCancel()
				err := port.Close()
				if err != nil {
					pm.logger.Error().Msgf("Error closing port '%v'", err)
				}
				portManagerWaitGroup.Wait()
				pm.logger.Info().Msg("portManager done")
				return
			}
		case <-waitChan:
			{
				pm.logger.Info().Msg("portManager looping")
			}
		}
	}
}

func (pm *PortManager) writer(port serial.Port, wg *sync.WaitGroup, ctx context.Context) {
	buff := make([]byte, 1)
	defer wg.Done()

	for {
		select {
		case toWrite, more := <-pm.outChan:
			{
				if !more {
					return
				}
				buff[0] = toWrite
				_, err := port.Write(buff)
				if err != nil {
					pm.logger.Error().Msgf("Failed writing to port '%v'", err)
					return
				}
				pm.logger.Trace().Msgf("Wrote byte '%v'", toWrite)
			}
		case <-ctx.Done():
			{
				pm.logger.Info().Msg("Writer done")
				return
			}
		}
	}
}

func (pm *PortManager) reader(port serial.Port, wg *sync.WaitGroup, portCancel context.CancelFunc) {
	buff := make([]byte, 100)
	defer wg.Done()

	for {
		n, err := port.Read(buff)
		if n == 0 {
			pm.logger.Info().Msg("0 bytes to read, done reading\n")
			portCancel()
			return
		}
		if err != nil {
			//pm.logger.Info().Msg("Got error '%v', done reading\n", err)
			return
		}
		for i := 0; i < n; i++ {
			element := buff[i]
			//pm.logger.Info().Msg("Received '%v'", element)
			pm.inChan <- element
		}
	}
}

func (pm *PortManager) heartbeat(ctx context.Context) {
	defer func() { recover() }()

	pm.logger.Info().Msg("Heartbeat starting")

	for {
		select {
		case <-ctx.Done():
			{
				pm.logger.Info().Msg("Heartbeat done")
				pm.waitGroup.Done()
				return
			}
		case <-time.After(200 * time.Millisecond):
			{
				pm.outChan <- 255
			}
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
