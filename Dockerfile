FROM golang:1.10-alpine AS builder

RUN apk add --no-cache \
	git

ARG GO_ROOT_IMPORT_PATH=github.com/icedream/irc-medialink

ENV CGO_ENABLED 0

COPY . "$GOPATH/src/$GO_ROOT_IMPORT_PATH"
WORKDIR "$GOPATH/src/$GO_ROOT_IMPORT_PATH"
RUN go get -v -d
RUN go build -ldflags '-extldflags "-static"' -o /irc-medialink
RUN cp *.tpl /

###

FROM alpine:3.7

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /irc-medialink /usr/local/bin
COPY --from=builder /*.tpl .
ENTRYPOINT ["irc-medialink"]
