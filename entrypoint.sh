#!/bin/bash

env GOOS=$(echo "$OS" | tr "[:upper:]" "[:lower:]") GOARCH=amd64 go build -o dist/$(echo "$OS" | tr "[:upper:]" "[:lower:]")

# Select right go binary for runner os
$GITHUB_ACTION_PATH/dist/$(echo "$OS" | tr "[:upper:]" "[:lower:]")
