# Raspberry PI IoT things

## Environment
```
export IOT_REGION=europe-west1
export IOT_REGISTRY=<registryname>
export IOT_DEVICE_NAME=<devicename>
export IOT_MQTT_HOST=tls://mqtt.googleapis.com:8883
export IOT_PROJECT_ID=<google-cloud-project-id-here>
```
## Led blinker

Blinks led. Either red, green or blue color is supported.

## Monitor

Detects motions from motion sensor and sends data to the cloud. Also blinks leds.

Data is sent to Cloud IoT service in JSON format using MQTT transport.

### Dependencies
```
go get
```

### Installation

First, compile:
```
./compile_rpi.sh
```

Then copy the binary and `ec_private.pem` to Raspberry PI or other device.

And finally set environment variables and run!


## Certificates and device creation for Cloud IoT connectivity

Create them:
```
./gen_keys.sh
```