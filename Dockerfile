FROM alpine:3.13.0

LABEL MAINTAINER 'janog-netcon'

RUN apk add --no-cache ca-certificates && update-ca-certificates
# build by goreleaser
ADD netcon /

ENTRYPOINT ["/netcon"]
