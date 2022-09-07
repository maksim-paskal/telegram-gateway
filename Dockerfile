FROM alpine:latest

COPY ./telegram-gateway /usr/local/bin/telegram-gateway
RUN apk upgrade && apk add --no-cache ca-certificates

USER 30001

ENTRYPOINT [ "/usr/local/bin/telegram-gateway" ]