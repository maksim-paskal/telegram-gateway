test:
	@scripts/validate-license.sh
	go mod tidy
	go fmt ./cmd
	go test -race ./cmd
	golangci-lint run -v
build:
	docker build . -t paskalmaksim/telegram-gateway:dev
push:
	docker push paskalmaksim/telegram-gateway:dev
run:
	go run -race -v ./cmd --log.level=DEBUG --log.pretty $(args)
testProm:
	curl -H "Content-Type: application/json" --data @scripts/test-data-prom.json http://localhost:9090/prom
testSentry:
	curl -H "Content-Type: application/json" --data @scripts/test-data-sentry.json http://localhost:9090/sentry
heap:
	go tool pprof -http=127.0.0.1:8080 http://localhost:9090/debug/pprof/heap