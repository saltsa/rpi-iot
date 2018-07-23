[![CircleCI](https://circleci.com/gh/saltsa/rpi-iot.svg?style=svg)](https://circleci.com/gh/saltsa/rpi-iot)

# Raspberry PI IoT things

## Environment
```
export IOT_REGION=europe-west1
export IOT_REGISTRY=<registryname>
export IOT_DEVICE_NAME=<devicename>
export IOT_MQTT_HOST=tls://mqtt.googleapis.com:8883
export IOT_PROJECT_ID=<google-cloud-project-id-here>
export IOT_DEBUG=true # to enable debug
```
## Led blinker

Blinks led. Either red, green or blue color is supported.

## Monitor

Detects motions from motion sensor and sends data to the cloud. Also blinks leds.

Data is sent to Cloud IoT service in JSON format using MQTT transport.

### Dependencies

This uses go 1.11.


```
go get
```

### Installation

First, compile:
```
./compile_rpi.sh 
```

Alternatively, load binary directly from CI.

Then copy the binary and `ec_private.pem` to Raspberry PI.

And finally set environment variables and run! For more permanent solution,
there's `rpi-mon.service` file for systemd.


## Certificates and device creation for Cloud IoT connectivity

Create them:
```
./gen_keys.sh
```
