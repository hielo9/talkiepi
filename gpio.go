package talkiepi

import (
	"fmt"
	"time"

	"github.com/dchote/gpio"
	"github.com/stianeikeland/go-rpio"
)

func (b *Talkiepi) initGPIO() {
	// we need to pull in rpio to pullup our button pin
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		b.GPIOEnabled = false
		return
	} else {
		b.GPIOEnabled = true
	}

	ButtonPinPullUp := rpio.Pin(ButtonPin)
	ButtonPinPullUp.PullUp()

	rpio.Close()
	
	var upTime int = 0
	var downTime int = 0
	var click int = 0

	// unfortunately the gpio watcher stuff doesnt work for me in this context, so we have to poll the button instead
	b.Button = gpio.NewInput(ButtonPin)
	go func() {
		for {
			currentState, err := b.Button.Read()

			if currentState != b.ButtonState && err == nil {
				b.ButtonState = currentState

				if b.Stream != nil {
					if b.ButtonState == 1 {
						fmt.Printf("Button is released\n")
						upTime = 0
						if downTime<250 {
							click = click+1
							fmt.Printf("add a click %d\n", click)
							if click==2 {  // this is a double click and we need to toggle the connection
								if b.IsConnected == false {
									fmt.Printf("click2 false\n")
									b.Connect()
								} else {
									fmt.Printf("click2 false\n")
									b.Client.Disconnect()
								}
						} else {
							fmt.Printf("stop transmitting\n")
							b.TransmitStop()
						}
					} else {
						fmt.Printf("Button is pressed\n")
						// Let's see what's intended before transmitting
						//b.TransmitStart()
						if upTime>500 {  // this obviously wasn't a double click
							click = 0
							downTime = 0
							fmt.Printf("over 500\n")
						}
					}
				}

			} else {
				if currentState == 1 {
					if upTime < 1000 {
						upTime = upTime+10
					}
				} else {
					if downTime < 1000 {
						downTime = downTime+10
					}
					if downTime>250 && b.Stream != nil && err==nil {  // all right, it's down long enough to assume this is transmission
						fmt.Printf("start transmission\n")
						b.TransmitStart()
						click = 0
						downTime = downTime + 1  // let's make sure if someone releases right at 250 ms, we don't keep coming in here (but I think I might have solved this by making sure currentState != 1
					}
				}
			}
				
				

			time.Sleep(10 * time.Millisecond)
		}
	}()

	// then we can do our gpio stuff
	if !SeeedStudio {
		b.OnlineLED = gpio.NewOutput(OnlineLEDPin, false)
		b.ParticipantsLED = gpio.NewOutput(ParticipantsLEDPin, false)
		b.TransmitLED = gpio.NewOutput(TransmitLEDPin, false)
	}
}

func (b *Talkiepi) LEDOn(LED gpio.Pin) {
	if b.GPIOEnabled == false {
		return
	}

	LED.High()
}

func (b *Talkiepi) LEDOff(LED gpio.Pin) {
	if b.GPIOEnabled == false {
		return
	}

	LED.Low()
}

func (b *Talkiepi) LEDOffAll() {
	if b.GPIOEnabled == false {
		return
	}

	b.LEDOff(b.OnlineLED)
	b.LEDOff(b.ParticipantsLED)
	b.LEDOff(b.TransmitLED)
}
