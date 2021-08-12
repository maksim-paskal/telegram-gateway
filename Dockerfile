FROM alpine:3.4

COPY ./telegram-gateway /usr/local/bin/telegram-gateway
RUN apk add --no-cache ca-certificates

CMD telegram-gateway