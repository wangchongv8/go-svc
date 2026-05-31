FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
ARG TARGETARCH
ARG SERVICE_MAIN
ARG SERVICE_CONF_DIR

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build -o /app/bin ./${SERVICE_MAIN}

FROM alpine:3.21
ARG SERVICE_CONF_DIR

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/bin /app/bin
COPY ${SERVICE_CONF_DIR} /app/etc/
WORKDIR /app

ENTRYPOINT ["/app/bin"]
CMD ["-f", "/app/etc/config.yaml"]
