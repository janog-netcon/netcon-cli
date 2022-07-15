FROM golang:1.18-alpine

LABEL MAINTAINER 'janog-netcon'

WORKDIR /app

COPY . /app

RUN go build ./cmd/netcon/netcon.go

ENTRYPOINT ["/app/netcon"]
