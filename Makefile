BINARY=gserver
CLIENT_BINARY=gclient

VERSION=1.0.0
BUILD=`git rev-parse HEAD`

# ToDo: set verions stuffs in files
# Setup the -ldflags option for go build here, interpolate the variable values
# LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"


build:
	go build -o ${BINARY} server/*.go
	go build -o ${CLIENT_BINARY} client/*.go


run:
	./${BINARY}

install:
	go install

clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi

.PHONY: build run test install clean
