FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /chat_app
RUN mkdir -p /chat_app/proto

WORKDIR /chat_app

COPY ./proto/service.pb.go /chat_app/proto
COPY ./main.go /chat_app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o chat_app .

CMD ./chat_app