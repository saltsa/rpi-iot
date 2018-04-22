#!/bin/sh

set -eu

openssl ecparam -genkey -name prime256v1 -noout -out ec_private.pem
openssl ec -in ec_private.pem -pubout -out ec_public.pem

echo "To delete old token use command:"
echo "gcloud iot devices credentials delete --region $IOT_REGION  --registry $IOT_REGISTRY --device $IOT_DEVICE_NAME 0"
echo "creating new token"

gcloud iot devices credentials create --path ec_public.pem --region $IOT_REGION  --registry $IOT_REGISTRY --device $IOT_DEVICE_NAME --type es256-pem
echo "credentials done"

echo "deploy software and ec_public.pem file"



