#!/usr/bin/env bash
set -ex

TARGET=192.168.2.38
FILE=smartmeter

env GOOS=linux GOARCH=arm GOARM=6 go build -o $FILE

ssh root@$TARGET service $FILE stop
scp $FILE root@$TARGET:/usr/bin/
ssh root@$TARGET service $FILE start

rm $FILE
