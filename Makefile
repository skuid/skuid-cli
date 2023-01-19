SHA = $(shell git rev-parse --short HEAD)
GO_PKGS = $$(go list ./... | grep -v vendor)
REPOSITORY = 095427547185.dkr.ecr.us-west-2.amazonaws.com/skuid/skuid
TAG = latest
VOL_PATH=/go/src/github.com/skuid/skuid-cli
GO_VERSION=1.18
IMAGE ?= golang

ARCH=amd64
OS=darwin

VERSION=`cat .version`
LDFLAGS=-ldflags="-w -X github.com/skuid/skuid-cli/version.Name=$(VERSION)"

.PHONY: setup fmt vendored

test:
	go test -v -short ./...

bench:
	go test -benchmem -v -bench=. ./...

setup:
	go get -u github.com/kardianos/govendor

fmt:
	go fmt $(GO_PKGS)

docker-build: fmt
	docker run --rm \
		-e GOOS=$(OS) \
		-e GOARCH=$(ARCH) \
		-v $$(pwd):$(VOL_PATH) -w $(VOL_PATH) $(IMAGE):$(GO_VERSION) \
		go build -v -a -tags netgo -installsuffix netgo $(LDFLAGS)
		
build:
	go build .

vendored:
	test $$(govendor list +e |wc -l | awk '{print $$1}') -lt 1

docker:
	docker run --rm -v $$(pwd):$(VOL_PATH) -u jenkins-slave -w $(VOL_PATH) $(IMAGE):$(GO_VERSION) go build -v -a -tags netgo -installsuffix netgo $(LDFLAGS)
	docker build -t $(REPOSITORY):$(TAG) .

push:
	$$(aws ecr get-login --region us-west-2)
	docker push $(REPOSITORY):$(TAG)

# specifically used for downloads via github.
release:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o skuid-cli_linux_amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o skuid-cli_darwin_amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o skuid-cli_windows_amd64.exe

# this command creates a directory .coverage 
# then outputs coverage data into .coverage/cover.out, 
# then generates a readable html file in .coverage/coverage.html
cover:
	@mkdir .coverage || echo "hidden coverage folder exists"
	@go test -short -v -cover ./... -coverprofile .coverage/coverage.out
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
