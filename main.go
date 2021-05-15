package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpio "github.com/stianeikeland/go-rpio"

	"github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttInterval = 15 * time.Second
)

var msgHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.Printf("TOPIC: %s", msg.Topic())
	log.Printf("ID: %d", msg.MessageID())
	log.Printf("MSG: %s", msg.Payload())
}

var configHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	// TODO: Update config if necessary
	log.Printf("got config: %s", msg.Payload())
}

func createToken() string {
	// "iat": time.Now().Add(-10 * time.Minute).Unix(),
	// "exp": time.Now().Add(-9 * time.Minute).Unix(),

	ept := 20 * time.Minute
	log.Printf("creating new jwt, expiring in %s", ept)
	iat := time.Now().Unix()
	exp := time.Now().Add(20 * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": iat,
		"exp": exp,
		"aud": projectID,
	})

	keyStr, err := ioutil.ReadFile("ec_private.pem")
	if err != nil {
		log.Println("failed to read priv key:", err)
		return ""
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(keyStr)
	if err != nil {
		log.Println("failed to parse ec priv key:", err)
		return ""
	}

	out, err := token.SignedString(key)
	if err != nil {
		log.Println("failed to sign:", err)
		return ""
	}
	return out
}

func cp() (user, pass string) {
	token := createToken()
	log.Debugf("New creds: %s", token)
	return "user", token
}

type dataStruct struct {
	Motion      bool      `json:"motion"`
	MotionCount uint64    `json:"motionCount"`
	Time        time.Time `json:"time"`
	PublishedAt time.Time `json:"publishedAt"`
}

func newDataStruct() dataStruct {
	d := dataStruct{}
	d.Time = time.Now()
	d.MotionCount = motionCount
	return d
}

type myLogger struct{}

func (myLogger) Println(v ...interface{})               { log.Debugln(v) }
func (myLogger) Printf(format string, v ...interface{}) { log.Debugf(format, v) }

func startMqtt() {

	// MQTT.DEBUG = myLogger{}
	MQTT.CRITICAL = myLogger{}
	MQTT.ERROR = myLogger{}
	MQTT.WARN = myLogger{}

	opts := MQTT.NewClientOptions().AddBroker(mqttHost)
	opts.SetClientID(getMqttClientID())
	opts.SetDefaultPublishHandler(msgHandler)
	opts.SetCredentialsProvider(cp)

	opts.SetProtocolVersion(4)
	c := MQTT.NewClient(opts)

	log.Println("connecting to:", mqttHost)
	log.Println("client id:", getMqttClientID())

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln(token.Error())
	}

	//time.Sleep(3 * time.Second)
	var topic = fmt.Sprintf("/devices/%s/config", deviceID)
	log.Println("subscribing to config topic:", topic)

	if token := c.Subscribe(topic, 1, configHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	stateTopic := fmt.Sprintf("/devices/%s/state", deviceID)
	telemetryTopic := fmt.Sprintf("/devices/%s/events", deviceID)

	for i := uint64(0); ; i++ {
		data := motionState.Load().(dataStruct)
		data.PublishedAt = time.Now()
		motionState.Store(data)

		textObj, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("json marshal fail: %s", err)
		}
		text := string(textObj)

		var token MQTT.Token

		if i%10 == 0 {
			stateText := fmt.Sprintf("motionCount=%d", data.MotionCount)
			log.Printf("Sending state to topic %s: %s", stateTopic, stateText)
			token = c.Publish(stateTopic, 1, false, stateText)
			token.Wait()
		}

		log.Debugf("publishing %d to %s: %s", i, telemetryTopic, text)
		token = c.Publish(telemetryTopic, 1, false, text)
		token.Wait()

		// this is blocking 200 ms or something
		go blinkLed("blue")

		time.Sleep(mqttInterval)
	}
}

var gpioReady bool
var motionPin rpio.Pin

var mqttHost string
var projectID string
var location string
var registry string
var deviceID string

func getMqttClientID() string {
	return "projects/" + projectID +
		"/locations/" + location +
		"/registries/" + registry +
		"/devices/" + deviceID
}

func init() {

	viper.SetDefault("region", "europe-west1")
	viper.SetDefault("mqtt_host", "tls://mqtt.googleapis.com:8883")

	// check the environment variables
	viper.SetEnvPrefix("IOT")
	viper.AutomaticEnv()
	mqttHost = viper.GetString("mqtt_host")
	projectID = viper.GetString("project_id")
	location = viper.GetString("region")
	registry = viper.GetString("registry")
	deviceID = viper.GetString("device_name")

	debug := viper.GetBool("debug")

	if debug {
		log.Println("Enabling debug logging")
		log.SetLevel(log.DebugLevel)
	}

	err := rpio.Open()
	if err != nil {
		log.Println("gpio open fail", err)
	} else {
		gpioReady = true
	}

	if gpioReady {
		log.Println("configuring motion pin for input mode")
		motionPin = rpio.Pin(23)
		motionPin.Input()
	}
	motionState.Store(newDataStruct())
}

// getMotion reads motion from motion sensors. True if detected, false otherwise
func getMotion() bool {
	v := motionPin.Read()
	if v != 0 {
		return true
	}
	return false
}

var motionState atomic.Value
var motionCount uint64

// updateMotionStruct updates motion struct when necessary, running goroutine
// it also reads motion from the sensor
func updateMotionStruct() {
	for {
		ds := newDataStruct()
		ds.Motion = getMotion()

		if ds.Motion {
			atomic.AddUint64(&motionCount, 1)
		}
		oldDs := motionState.Load().(dataStruct)

		// update only if data published or (if not published)
		// when motion is detected or the value has changed one minute ago and
		// state changed
		if !oldDs.PublishedAt.IsZero() ||
			(time.Since(oldDs.Time) > time.Minute && oldDs.Motion != ds.Motion) ||
			ds.Motion {
			motionState.Store(ds)
			go blinkLed("red")
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {

	// first configure gpio
	initBlink()

	// start blinking thread

	go updateMotionStruct()

	// wait for it and then try to exit
	// doesn't work correctly with mqtt atm, so it panics
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go startMqtt()

	log.Println("waiting signal...")

	<-sigc

	log.Println("got signal")

	// stopchan in led_blink.go
	close(stopchan)
	log.Println("led shut down")
	time.Sleep(1 * time.Second)
}
