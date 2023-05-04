FROM golang:1.19.5-alpine AS builder

RUN apk add --no-cache \
	git

ENV CGO_ENABLED 0

WORKDIR /usr/src/medialink
# download dependencies (separate cache layer)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=secret,id=mynetrc,dst=/root/.netrc --mount=type=secret,id=myknownhosts,dst=/root/.ssh/known_hosts --mount=type=ssh go mod download
# compile rest of code and install to /target for copying to final image
COPY ./ ./
ARG APPLICATION_NAME
RUN --mount=type=cache,target=/root/.cache/go-build \
	EXTRA_LDFLAGS='-extldflags -static' ./build.sh -o /irc-medialink

###

FROM alpine:3.17

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /irc-medialink /usr/local/bin
COPY --from=builder /*.tpl .
ENTRYPOINT ["irc-medialink"]
