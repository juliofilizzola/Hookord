# Stage 1: Builder
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -X main.AppVersion=$(git describe --tags --always --dirty)" \
    -o /Hookord ./cmd/hookord

FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.source="https://github.com/juliofilizzola/hookord"
LABEL org.opencontainers.image.version="0.0.1"

COPY --from=builder /hookord /hookord
COPY configs/config.example.yaml /configs/config.yaml

EXPOSE 8080
USER 1000:1000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/hookord"]
CMD ["--config", "/configs/config.yaml"]
