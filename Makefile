BINARY=bjs-tracker
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w"

.PHONY: build build-pi build-all clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

build-pi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(BINARY)-linux-arm .

build-all: build build-pi
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

clean:
	rm -f $(BINARY) $(BINARY)-linux-*
