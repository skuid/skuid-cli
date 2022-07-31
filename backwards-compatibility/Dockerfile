FROM golang:1.18-alpine3.16 as build

ARG HOST
ARG USERNAME
ARG PASSWORD

RUN apk update
RUN apk add git

# todo: grab the images instead
WORKDIR /
RUN git clone https://github.com/skuid/skuid-cli skuid-cli
RUN git clone -b PLIN-3600 https://github.com/skuid/skuid-cli tides

WORKDIR /skuid-cli
RUN go build -o run . 

WORKDIR /tides
RUN go build -o run . 

FROM build as run

RUN apk add diffutils && diff --version

WORKDIR /

RUN ./skuid-cli/run retrieve \
		--host $HOST \
		--username $USERNAME \
		--password $PASSWORD \
		--dir old \
		--verbose

RUN ./tides/run retrieve \
		--host $HOST \
		--username $USERNAME \
		--password $PASSWORD \
		--directory new \
		--verbose 
		
RUN find . -type f -name "*.json" -exec sh -c 'cat $0 | jq > $0.pretty' {} \;

CMD diff -r -x '*.json' old new