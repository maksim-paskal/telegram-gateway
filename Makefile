tag=dev
image=paskalmaksim/telegram-gateway:$(tag)
KUBECONFIG=$(HOME)/.kube/dev

test:
	@scripts/validate-license.sh
	go mod tidy
	go fmt ./cmd
	go test -race ./cmd
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v
test-release:
	git tag -d `git tag -l "helm-chart-*"`
	go run github.com/goreleaser/goreleaser@latest release --debug --snapshot --skip-publish --clean
build:
	git tag -d `git tag -l "helm-chart-*"`
	go run github.com/goreleaser/goreleaser@latest build --clean --skip-validate
	mv ./dist/gateway_linux_amd64_v1/telegram-gateway telegram-gateway
	docker build --pull --push . -t $(image)
promote-to-beta:
	git tag -d `git tag -l "helm-chart-*"`
	go run github.com/goreleaser/goreleaser@latest release --clean --snapshot
	docker push paskalmaksim/telegram-gateway:beta-amd64
	docker push paskalmaksim/telegram-gateway:beta-arm64
	docker manifest create --amend paskalmaksim/telegram-gateway:beta \
	paskalmaksim/telegram-gateway:beta-arm64 \
	paskalmaksim/telegram-gateway:beta-amd64
	docker manifest push --purge paskalmaksim/telegram-gateway:beta
push:
	docker push $(image)
run:
	go run -race -v ./cmd --log.level=DEBUG -server.address=127.0.0.1:9090 --log.pretty $(args)
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
deploy:
	helm upgrade telegram-gateway ./charts/telegram-gateway \
	--install \
	--create-namespace \
	--namespace telegram-gateway-test \
	--set registry.image=$(image) \
	--set registry.imagePullPolicy=Always
clean:
	kubectl delete namespace telegram-gateway-test