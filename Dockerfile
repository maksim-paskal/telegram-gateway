FROM golang:1.16 as build

COPY ./cmd /usr/src/telegram-gateway/cmd
COPY go.* /usr/src/telegram-gateway/
COPY .git /usr/src/telegram-gateway/

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GOFLAGS="-trimpath"

RUN cd /usr/src/telegram-gateway \
  && go mod download -x \
  && go mod verify \
  && go build -v -o telegram-gateway -ldflags \
  "-X main.gitVersion=$(git describe --tags `git rev-list --tags --max-count=1`)-$(date +%Y%m%d%H%M%S)-$(git log -n1 --pretty='%h')" \
  ./cmd

RUN /usr/src/telegram-gateway/telegram-gateway --version

FROM alpine:3

COPY --from=build /usr/src/telegram-gateway/telegram-gateway /usr/local/bin/telegram-gateway
RUN apk add --no-cache ca-certificates

CMD telegram-gateway