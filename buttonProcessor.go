package main

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

type ButtonProcessor struct {
	logger       *zerolog.Logger
	inChan       chan byte
	outChan      chan byte
	waitGroup    *sync.WaitGroup
	buttonStates []bool
}

func NewButtonProcessor(logger *zerolog.Logger, inChan chan byte, outChan chan byte, waitGroup *sync.WaitGroup) *ButtonProcessor {
	return &ButtonProcessor{
		logger:       ptr(logger.With().Str(LogKey.Module, "ButtonProcessor").Logger()),
		inChan:       inChan,
		outChan:      outChan,
		waitGroup:    waitGroup,
		buttonStates: make([]bool, 100),
	}
}
func (bp *ButtonProcessor) Start(ctx context.Context) {
	bp.waitGroup.Add(1)
	go bp.loop(ctx)
}

func (bp *ButtonProcessor) loop(ctx context.Context) {
	defer bp.waitGroup.Done()
	bp.logger.Info().Msg("Starting")

	bp.initLEDs()

	for {
		select {
		case <-ctx.Done():
			{
				bp.logger.Info().Msg("Done")
				return
			}
		case readByte, more := <-bp.inChan:
			{
				if !more {
					bp.logger.Info().Msg("No more\n")
					return
				}
				var led byte = 0

				var button = readByte & 0x7f
				var goingOff = readByte & 0x80
				if goingOff != 0 {
					continue
				}

				led = convertButtonToLed(button)
				if led == 0 {
					continue
				}
				newState := !bp.buttonStates[button]
				if !newState {
					led |= 0x80
				}
				bp.buttonStates[button] = newState
				bp.outChan <- led
			}
		}
	}
}

func (bp *ButtonProcessor) initLEDs() {
	for _, buttonPinNumber := range BUTTONS {
		led := convertButtonToLed(buttonPinNumber)
		if !bp.buttonStates[buttonPinNumber] {
			led |= 0x80
		}
		bp.outChan <- led
	}
}

func convertButtonToLed(button byte) byte {

	var led = BUTTON_LED_RELATIONS[button]

	return led
}
