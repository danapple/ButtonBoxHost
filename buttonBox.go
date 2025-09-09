package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
)

func main() {

	rootLogger := ptr(log.With().Logger())

	if len(os.Args) < 2 {
		rootLogger.Fatal().Msg("No port name provided")
	}

	var argsWithoutProg = os.Args[1:]
	var portName = argsWithoutProg[0]

	rootLogger = ptr(rootLogger.With().Str(LogKey.Port, portName).Logger())
	logger := ptr(rootLogger.With().Str(LogKey.Module, "Main").Logger())
	logger.Info().Msg("Starting host processor")

	ctx, cancel := context.WithCancel(context.Background())

	inChan := make(chan byte, 10)
	outChan := make(chan byte, 10)

	waitGroup := &sync.WaitGroup{}

	portManager := NewPortManager(rootLogger, portName, inChan, outChan, waitGroup)
	portManager.Start(ctx)

	restAPI := NewRestApi(rootLogger, inChan, outChan, waitGroup)
	restAPI.Start(ctx)

	buttonProcessor := NewButtonProcessor(rootLogger, inChan, outChan, waitGroup)
	buttonProcessor.Start(ctx)

	waitForSignal()
	cancel()
	waitGroup.Wait()

	close(inChan)
	close(outChan)
	logger.Info().Msg("Done")
}

func waitForSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-signals
		switch sig {
		case syscall.SIGINT:
			return
		case syscall.SIGTERM:
			return
		}
	}
}
