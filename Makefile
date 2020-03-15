build:
	helm package charts/*
	helm --url https://test repo index .