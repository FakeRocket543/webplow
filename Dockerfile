FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /webplow ./cmd/server && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /webplow-token ./cmd/token

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1000 webplow && \
    mkdir -p /data /uploads && \
    chown webplow:webplow /data /uploads
COPY --from=builder /webplow /webplow-token /usr/local/bin/
USER webplow
ENTRYPOINT ["webplow"]
