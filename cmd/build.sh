#!/bin/sh
echo "Build go application"
GOOS=linux GOARCH=amd64 go build -o cdr-pusher main.go
