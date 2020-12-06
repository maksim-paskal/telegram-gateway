export GO111MODULE = on

test:
	@scripts/validate-license.sh
	go mod tidy
	go fmt ./cmd
	go test ./cmd
	golangci-lint run --allow-parallel-runners -v --enable-all --disable testpackage,funlen --fix
build:
	docker build . -t paskalmaksim/telegram-gateway:dev
build-all:
	@scripts/build-all.sh
run:
	GOFLAGS="-trimpath" go build -v -o /tmp/telegram-gateway ./cmd && /tmp/telegram-gateway --log.level=DEBUG $(args)