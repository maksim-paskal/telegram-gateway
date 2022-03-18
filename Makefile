tag=dev
image=paskalmaksim/telegram-gateway:$(tag)

test:
	@scripts/validate-license.sh
	go mod tidy
	go fmt ./cmd
	go test -race ./cmd
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v
test-release:
	go run github.com/goreleaser/goreleaser@latest release --snapshot --skip-publish --rm-dist
build:
	go run github.com/goreleaser/goreleaser@latest build --rm-dist --skip-validate
	mv ./dist/telegram-gateway_linux_amd64/telegram-gateway telegram-gateway
	docker build --pull . -t $(image)
push:
	docker push $(image)
run:
	go run -race -v ./cmd --log.level=DEBUG --log.pretty $(args)
testProm:
	curl -H "Content-Type: application/json" --data @scripts/test-data-prom.json http://localhost:9090/prom
testSentry:
	curl -H "Content-Type: application/json" --data @scripts/test-data-sentry.json http://localhost:9090/sentry
heap:
	go tool pprof -http=127.0.0.1:8080 http://localhost:9090/debug/pprof/heap
scan:
	@trivy image \
	-ignore-unfixed --no-progress --severity HIGH,CRITICAL \
	$(image)