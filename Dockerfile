FROM golang:1.15 as build

COPY ./cmd /usr/src/telegram-gateway/cmd
COPY go.* /usr/src/telegram-gateway/

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GOFLAGS="-trimpath"

RUN cd /usr/src/telegram-gateway \
  && go mod download \
  && go mod verify \
  && go build -v -o telegram-gateway -ldflags "-X main.buildTime=$(date +"%Y%m%d%H%M%S")" ./cmd

FROM alpine:latest

COPY --from=build /usr/src/telegram-gateway/telegram-gateway /usr/local/bin/telegram-gateway
RUN apk add --no-cache ca-certificates

CMD telegram-gateway