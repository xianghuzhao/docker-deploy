FROM golang:alpine as builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0
ARG GOPROXY

ARG DOCKER_REMOTE_VERSION

COPY . /pro

WORKDIR /pro

RUN go build -v -ldflags "-X main.version=${DOCKER_REMOTE_VERSION}" -o docker-remote


FROM docker as docker


FROM alpine:latest

RUN apk add --no-cache openssh-client

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY --from=builder /pro/docker-remote /usr/local/bin/docker-remote

CMD ["/usr/local/bin/docker-remote"]
