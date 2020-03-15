test:
	go test *.go
build:
	docker build .
	
.PHONY: test-style
test-style:
	GO111MODULE=on golangci-lint run
	@scripts/validate-license.sh