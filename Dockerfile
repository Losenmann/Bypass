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
COPY ./src/ /
WORKDIR /opt/bypass
RUN mkdir -p /opt/bypass/redis \
    && apk add --no-cache \
        bash \
        iproute2 \
        nftables \
        openresolv \
        redis \
        curl \
    && pip3 install --no-cache-dir --break-system-packages -r /etc/requirements.txt \
    && chmod 0755 /entrypoint.sh
COPY --from=awg --chmod=0755 /usr/bin/amneziawg-go /usr/bin/
COPY --from=awg-tools --chmod=0755 /usr/bin/awg /usr/bin/awg-quick /usr/bin/wg /usr/bin/wg-quick /usr/bin/
ENTRYPOINT ["/entrypoint.sh"]
