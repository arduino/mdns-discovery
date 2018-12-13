#!/bin/bash

# Build for every architecture and emit json and shasums

rm -rf distrib

# TODO: replace me with proper go modules stuff
export GOPATH=$PWD

export CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -o distrib/linux64/mdns-discovery
GOOS=linux GOARCH=386 go build -o distrib/linux32/mdns-discovery
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/mdns-discovery
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/mdns-discovery
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows/mdns-discovery.exe
GOOS=darwin GOARCH=amd64 go build -o distrib/darwin/mdns-discovery

cd distrib
zip -r ../mdns-discovery-${VERSION}.zip *
cd ..

shasum mdns-discovery-${VERSION}.zip
sha256sum mdns-discovery-${VERSION}.zip
ls -la mdns-discovery-${VERSION}.zip
