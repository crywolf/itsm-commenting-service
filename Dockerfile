# syntax=docker/dockerfile:experimental
FROM golang:1.16.5-alpine3.12 as build_base
ARG GOPROXY
ARG GOPRIVATE='github.com/KompiTech/*'
RUN apk add --no-cache --update alpine-sdk make git openssl gcc openssh
RUN mkdir /src
RUN mkdir /root/.ssh/
RUN touch /root/.ssh/known_hosts
RUN git config --global url."git@github.com:KompiTech".insteadOf https://github.com/KompiTech
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts
COPY go.mod /src
COPY go.sum /src
WORKDIR /src
RUN --mount=type=ssh go mod download

FROM build_base as build
COPY . /src
WORKDIR /src
RUN --mount=type=ssh make build-linux

FROM alpine:3.12
RUN apk update && apk add ca-certificates tzdata && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=build /src/build/itsm-commenting-service.linux /
ENTRYPOINT ["/itsm-commenting-service.linux"]
CMD [ "--config-path", "/etc/itsm-commenting-service" ]