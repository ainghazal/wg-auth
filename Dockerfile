# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
ENV CGO_ENABLED=0

WORKDIR /src
COPY . /src
RUN go build -o /bin/wgauth ./cmd/dummy-auth

FROM alpine:3

COPY --from=builder /bin/wgauth /bin/wgauth

RUN apk add --no-cache \
    findutils openresolv iptables ip6tables iproute2 wireguard-tools

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x entrypoint.sh
EXPOSE 51820

ENTRYPOINT ["/entrypoint.sh"]
