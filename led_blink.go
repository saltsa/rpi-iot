package main

import (
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio"
)

var pins = [3]int{25, 8, 7}

// INTERVAL the interval in ms
const INTERVAL = 50 * time.Millisecond

var stopchan = make(chan struct{})

func blink(id int) {
	log.Println("starting to blink pin", pins[id])
	err := rpio.Open()
	if err != nil {
		log.Fatalln("gpio open fail", err)
	}

	defer rpio.Close()

	state := rpio.Low

	pin := rpio.Pin(pins[id])
	for {
		select {
		default:
			if state == rpio.Low {
				pin.High()
				state = rpio.High
			} else {
				pin.Low()
				state = rpio.Low
			}
			time.Sleep(INTERVAL)
		case <-stopchan:
			return
		}
	}
}

func initAndBlink(led string) {

	err := rpio.Open()
	if err != nil {
		log.Fatalln("gpio open fail", err)
	}

	defer rpio.Close()

	log.Println("initialising pins....")

	for _, i := range pins {
		pin := rpio.Pin(i)
		pin.Output()
		pin.Low()
		time.Sleep(INTERVAL)
		pin.High()
		time.Sleep(INTERVAL)
		pin.Low()
	}

	log.Printf("init done, args: %+v", os.Args)

	switch led {
	case "red":
		go blink(0)
	case "green":
		go blink(1)
	case "blue":
		go blink(2)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	<-sigc

	log.Println("software exit, shut down leds")
	// Stopping go routine
	close(stopchan)
	for _, i := range pins {
		pin := rpio.Pin(i)
		pin.Low()
	}
	log.Println("all leds off")
}

func main() {
	if len(os.Args) > 1 {
		led := os.Args[1]
		initAndBlink(led)
	}
}
