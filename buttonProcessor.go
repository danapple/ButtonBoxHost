package main

import (
	"context"
	"log"
	"sync"
)

var buttonStates = make([]bool, 100)

func initLeds() {
	for _, buttonPinNumber := range BUTTONS {
		led := convertButtonToLed(buttonPinNumber)
		if !buttonStates[buttonPinNumber] {
			led |= 0x80
		}
		outChan <- led
	}
}

func buttonProcessor(inChan <-chan byte, outChan chan<- byte, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	log.Printf("buttonProcessor starting\n")

	initLeds()

	for {
		select {
		case <-ctx.Done():
			{
				log.Printf("buttonProcessor Done\n")
				return
			}
		case readByte, more := <-inChan:
			{
				if !more {
					log.Printf("buttonProcessor no more\n")
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
				newState := !buttonStates[button]
				if !newState {
					led |= 0x80
				}
				buttonStates[button] = newState
				outChan <- led
			}
		}
	}
}

func convertButtonToLed(button byte) byte {
	var led byte = 0

	switch button {
	case WHITE_BUTTON:
		led = WHITE_BUTTON_LED
	case GREEN_BUTTON:
		led = GREEN_BUTTON_LED
	case YELLOW_BUTTON:
		led = YELLOW_BUTTON_LED
	case RED_BUTTON:
		led = RED_BUTTON_LED
	case SQUARE_BUTTON:
		led = SQUARE_BUTTON_LED
	default:
		led = 0
	}
	return led
}
