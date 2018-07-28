package main

import (
	"math/rand"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio"
)

var pins = [3]int{25, 8, 7}

// INTERVAL the interval in ms
const INTERVAL = 100 * time.Millisecond

var blinkLock sync.Mutex

var stopchan = make(chan struct{})

func nameToIdx(name string) int {
	switch name {
	case "red":
		return 0
	case "green":
		return 1
	case "blue":
		return 2
	default:
		return rand.Int() % 3
	}
}

func continuousBlink(id int) {
	log.Println("starting to blink pin", pins[id])

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

func blinkLed(color string) {
	blinkLock.Lock()
	defer blinkLock.Unlock()
	id := nameToIdx(color)
	pin := rpio.Pin(pins[id])
	pin.High()
	time.Sleep(INTERVAL)
	pin.Low()
	time.Sleep(INTERVAL)
}

func initBlink() {

	err := rpio.Open()
	if err != nil {
		log.Fatalln("gpio open fail", err)
	}

	//defer rpio.Close()

	log.Println("initialising pins....")

	for _, i := range pins {
		log.Printf("setting pin %d to output and low", i)
		pin := rpio.Pin(i)
		pin.Output()
		pin.Low()
	}

	// skip constant blinking
	/*
		defer func() {
			for _, i := range pins {
				pin := rpio.Pin(i)
				pin.Low()
			}
			log.Println("all leds off")
		}()

		log.Printf("init done, args: %+v", os.Args)

		switch led {
		case "red":
			blink(0)
		case "green":
			blink(1)
		case "blue":
			blink(2)
		}

		log.Println("blinking done")
	*/

	//time.Sleep(10 * time.Second)
	// Stopping go routine
	//close(stopchan)
}

func maintest() {
	if len(os.Args) > 1 {
		led := os.Args[1]
		initBlink()
		continuousBlink(nameToIdx(led))
	}
}
