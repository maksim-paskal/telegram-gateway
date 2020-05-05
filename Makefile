export GO111MODULE = on

test:
	go test *.go
	golangci-lint run
	@scripts/validate-license.sh
build:
	docker build . -t paskalmaksim/telegram-gateway:1.0.4
build-all:
	@scripts/build-all.sh