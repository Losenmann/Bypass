FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine3.22 AS awg
ARG ARG_AWG_VERSION=v0.2.13
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ADD https://github.com/amnezia-vpn/amneziawg-go.git#${ARG_AWG_VERSION} /tmp/awg
WORKDIR /tmp/awg
RUN go mod download \
    && go mod verify \
    && go build -ldflags '-s -w -extldflags "-fno-PIC -static"' -v -o /usr/bin

FROM alpine:3.22.1 AS awg-tools
ARG ARG_AWG_TOOLS_VERSION=v1.0.20250903
ADD https://github.com/amnezia-vpn/amneziawg-tools.git#${ARG_AWG_TOOLS_VERSION} /tmp/awg-tools
WORKDIR /tmp/awg-tools
# Patch https://lists.zx2c4.com/pipermail/wireguard/2023-February/007936.html
RUN sed -i '/net.ipv4.conf.all.src_valid_mark/s/&&/\&\& [[ $(sysctl -n net.ipv4.conf.all.src_valid_mark) != 1 ]] \&\& /' /tmp/awg-tools/src/wg-quick/linux.bash
RUN apk add --no-cache \
        bash \
        make \
        gcc \
        libc-dev \
        linux-headers \
    && make -C ./src \
    && make -C ./src install \
    && ln -s /usr/bin/awg /usr/bin/wg \
    && ln -s /usr/bin/awg-quick /usr/bin/wg-quick

FROM python:3.13.7-alpine3.22 AS app
ARG ARG_AWG_EXPORTER_REDIS_HOST="127.0.0.1"
ENV AWG_EXPORTER_REDIS_HOST=${ARG_AWG_EXPORTER_REDIS_HOST}
ADD --chmod=0755 https://raw.githubusercontent.com/amnezia-vpn/amneziawg-exporter/refs/heads/main/exporter.py /usr/bin/awg-exporter
WORKDIR /opt/app
RUN mkdir -p /opt/app/data /etc/redis \
    && apk add --no-cache \
        bash \
        iproute2 \
        nftables \
        openresolv \
        redis \
        curl \
    && pip3 install --no-cache-dir --break-system-packages -r requirements.txt \
    && printf "%s\n%s\n%s\n%s\n%s\n" \
        "#!/bin/bash" \
        "awg-quick up wg0; sleep 3" \
        "nft -f /etc/nftables.conf" \
        "redis-server /etc/redis/redis.conf" \
        "awg-exporter" \
        > /entrypoint.sh \
    && printf "%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n" \
        "bind 0.0.0.0" \
        "protected-mode no" \
        "port 6379" \
        "tcp-backlog 511" \
        "timeout 0" \
        "tcp-keepalive 300" \
        "daemonize yes" \
        "pidfile /run/redis.pid" \
        "loglevel warning" \
        "databases 16" \
        "always-show-logo no" \
        "set-proc-title no" \
        "save 3600 1" \
        "stop-writes-on-bgsave-error no" \
        "rdbcompression yes" \
        "rdbchecksum yes" \
        "dir /opt/app/data" \
        > /etc/redis/redis.conf \
    && echo "\
IyEvdXNyL3NiaW4vbmZ0IC1mCgp0YWJsZSBpcCBuYXQgewogICAgY2hhaW4gUE9TVFJPVVRJTkcg\
ewogICAgICAgIHR5cGUgbmF0IGhvb2sgcG9zdHJvdXRpbmcgcHJpb3JpdHkgc3JjbmF0OyBwb2xp\
Y3kgYWNjZXB0OwogICAgICAgIG9pZm5hbWUgIndnMCIgY291bnRlciBtYXNxdWVyYWRlCiAgICB9\
Cn0KCnRhYmxlIGlwIG1hbmdsZSB7CiAgICBjaGFpbiBQT1NUUk9VVElORyB7CiAgICAgICAgdHlw\
ZSBmaWx0ZXIgaG9vayBwb3N0cm91dGluZyBwcmlvcml0eSBtYW5nbGU7IHBvbGljeSBhY2NlcHQ7\
CiAgICAgICAgaXAgcHJvdG9jb2wgdWRwIG1ldGEgbWFyayAweDAwMDBjYTZjIGNvdW50ZXIgY29t\
bWVudCAiQ09OTk1BUksiCiAgICB9CgogICAgY2hhaW4gUFJFUk9VVElORyB7CiAgICAgICAgdHlw\
ZSBmaWx0ZXIgaG9vayBwcmVyb3V0aW5nIHByaW9yaXR5IG1hbmdsZTsgcG9saWN5IGFjY2VwdDsK\
ICAgICAgICBpcCBwcm90b2NvbCB1ZHAgY291bnRlciBjb21tZW50ICJDT05OTUFSSyIKICAgIH0K\
fQo=" | base64 -d > /etc/nftables.conf \
    && chmod 0755 /entrypoint.sh
COPY --from=awg --chmod=0755 /usr/bin/amneziawg-go /usr/bin/
COPY --from=awg-tools --chmod=0755 /usr/bin/awg /usr/bin/awg-quick /usr/bin/wg /usr/bin/wg-quick /usr/bin/
COPY ./healtz.sh /usr/bin/awg-health
ENTRYPOINT ["/entrypoint.sh"]
