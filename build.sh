#!/bin/bash

mkdir -p dist

SRC="./cmd/lumen/"

echo "Building binaries..."

GOOS=linux GOARCH=amd64 go build -o dist/lumen-linux-amd64 $SRC

GOOS=windows GOARCH=amd64 go build -o dist/lumen-windows-amd64.exe $SRC

GOOS=darwin GOARCH=amd64 go build -o dist/lumen-darwin-amd64 $SRC

GOOS=darwin GOARCH=arm64 go build -o dist/lumen-darwin-arm64 $SRC

echo "Done! Check the /dist folder."
