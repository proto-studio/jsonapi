ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: test test-docker bench coverage coverage-html reportcard generated

test:
	go test ./...

race:
	go test -race ./...

coverage:
	go test -coverpkg=./pkg/... -coverprofile=coverage.out ./pkg/...
	go tool cover -func coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out

reportcard:
	goreportcard-cli -v
