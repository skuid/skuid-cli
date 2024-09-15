FROM golang:1.23.1-alpine as build

# arguments to authenticate with the site we're retrieving from
ARG HOST
ARG USERNAME
ARG PASSWORD

RUN apk update
RUN apk add git

WORKDIR /
RUN git clone https://github.com/skuid/skuid-cli skuid-cli

# build each version
WORKDIR /skuid-cli
RUN go build -o run . 

FROM build as run

# use the right `diff` command
RUN apk add diffutils && diff --version

WORKDIR /

# retrieve using the old version
RUN ./skuid-cli/run retrieve \
		--host $HOST \
		--username $USERNAME \
		--password $PASSWORD \
		--dir site \
		--verbose

# so the problem with what I was seeing was that there are a ton of json files
# whose only differences are newlines and spacing. So, in order to circumvent this,
# let's format all the json files
RUN find . -type f -name "*.json" -exec sh -c 'cat $0 | jq > $0.pretty' {} \;

# ignore all the json files. instead we'll end up comparing `*.json.pretty` instead.
# CMD diff -r -x '*.json' old new