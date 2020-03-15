export GO111MODULE = on

test:
	go test *.go
	golangci-lint run
	@scripts/validate-license.sh
build:
	docker build .