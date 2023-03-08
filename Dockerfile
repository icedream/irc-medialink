FROM golang:1.20.2-alpine AS builder

RUN apk add --no-cache \
	git

ENV CGO_ENABLED 0

WORKDIR /usr/src/medialink
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG APPLICATION_NAME
RUN EXTRA_LDFLAGS='-extldflags -static' ./build.sh -o /irc-medialink
RUN cp *.tpl /

###

FROM alpine:3.17

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /irc-medialink /usr/local/bin
COPY --from=builder /*.tpl .
ENTRYPOINT ["irc-medialink"]
