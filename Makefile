export GO111MODULE = on

test:
	@scripts/validate-license.sh
	go mod tidy
	go fmt ./cmd
	go test ./cmd
	golangci-lint run --allow-parallel-runners -v --enable-all --disable testpackage,funlen --fix
build:
	docker build . -t paskalmaksim/telegram-gateway:1.0.4
build-all:
	@scripts/build-all.sh