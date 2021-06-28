FROM golang:1.14-alpine

RUN apk add --no-cache bash \
    bash-completion \
    bash-doc \
    coreutils \
    git \
    make

WORKDIR /go/app

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8888