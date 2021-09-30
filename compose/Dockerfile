FROM golang:1.15-alpine
LABEL authors="Jared Hooper, Kenny O'Neal, Amy Dewaal, Adam Cofer, Joshua Goodlett"
LABEL company="Skuid"

ENV SRC_DIR /go/src/github.com/skuid/skuid-cli

ADD . $SRC_DIR
WORKDIR $SRC_DIR

RUN apk add --no-cache --update gcc musl-dev git

# Get dependencies
RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o /tmp/dist/skuid_linux_amd64
RUN GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o /tmp/dist/skuid_darwin_amd64
RUN GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o /tmp/dist/skuid_windows_amd64.exe
