FROM golang:1.12 as build

COPY main.go /usr/src/telegram-gateway/main.go
COPY go.mod /usr/src/telegram-gateway/go.mod
COPY go.sum /usr/src/telegram-gateway/go.sum

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN cd /usr/src/telegram-gateway \
  && go mod download \
  && go mod verify \
  && go build -v -o telegram-gateway -ldflags "-X main.buildTime=$(date +"%Y%m%d%H%M%S")"

FROM alpine:latest

COPY --from=build /usr/src/telegram-gateway/telegram-gateway /usr/local/bin/telegram-gateway
RUN apk add --no-cache ca-certificates

CMD telegram-gateway