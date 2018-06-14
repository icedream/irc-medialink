FROM golang:1.10-alpine AS builder

RUN apk add --no-cache \
	git

ARG GO_ROOT_IMPORT_PATH=github.com/icedream/irc-medialink

ENV CGO_ENABLED 0

COPY . "$GOPATH/src/$GO_ROOT_IMPORT_PATH"
RUN go get -v -d "$GO_ROOT_IMPORT_PATH"
RUN go build -ldflags '-extldflags "-static"' -o /irc-medialink "$GO_ROOT_IMPORT_PATH"

###

FROM scratch

COPY --from=builder /irc-medialink /
ENTRYPOINT ["/irc-medialink"]
