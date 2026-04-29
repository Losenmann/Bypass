FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine3.23 AS builder
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0
WORKDIR /opt/src
RUN apk add upx
COPY . ./
RUN go mod download
RUN go mod verify
RUN go build -ldflags '-s -w -extldflags "-fno-PIC -static"' -v -o ./filter
RUN upx --best --lzma ./filter

FROM scratch AS app
COPY --from=builder /opt/src/filter .
ENTRYPOINT ["/filter"]