test:
	helm lint --strict helm/telegram-gateway
	helm template helm/telegram-gateway | kubectl apply --dry-run -f -
build:
	docker build .