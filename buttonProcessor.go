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

	var led = BUTTON_LED_RELATIONS[button]

	return led
}
