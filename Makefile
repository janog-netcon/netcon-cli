SHELL = /bin/bash

FORCE:
.PHONY: FORCE

build: FORCE
	goreleaser --snapshot --skip-publish --rm-dist

go-build:
	go build ./cmd/netcon/netcon.go
