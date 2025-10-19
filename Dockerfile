FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine3.22 AS awg
ARG TARGETOS
ARG TARGETARCH
ARG ARG_BYPASS_VERSION=3.0.0
ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}
WORKDIR /opt/src
COPY . ./
RUN go mod download \
    && go mod verify \
    && go build -ldflags '-s -w -extldflags "-fno-PIC -static"' -v -o /usr/bin/bypass
