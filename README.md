# Webplow

A WebP image conversion API service powered by imgproxy, with multi-user token management.

## Quick Start (Docker Compose)

```bash
docker compose up -d --build

# Create your first token
docker exec webplow-webplow-1 webplow-token add "user1"

# Test
curl -X POST https://webplow.example.com/ \
  -H "X-API-Key: YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -o output.webp
```

## Token Management

```bash
# Add
docker exec webplow-webplow-1 webplow-token add "site-name"

# List
docker exec webplow-webplow-1 webplow-token list

# Delete
docker exec webplow-webplow-1 webplow-token delete <key>

# Reload tokens (no restart, no downtime)
docker kill -s HUP webplow-webplow-1
```

Token changes require a reload via `SIGHUP`. This is a hot reload — existing connections are not interrupted.

Use `make token-add` / `make token-list` / `make token-delete` for local development.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TOKEN_FILE` | `tokens.json` | Token file path |
| `LISTEN_ADDR` | `127.0.0.1:9000` | Listen address |
| `IMGPROXY_URL` | `http://127.0.0.1:48080` | imgproxy backend URL |
| `TEMP_DIR` | `/var/www/imgproxy/uploads` | Temp directory |
| `MAX_FILE_SIZE` | `20971520` | Max upload size (bytes) |
| `READ_TIMEOUT` | `30s` | Read timeout |
| `WRITE_TIMEOUT` | `60s` | Write timeout |
| `LOG_FILE` | (empty, no logging) | Access log file path |

When deploying with Docker Compose, environment variables are set in `docker-compose.yml`. No `.env` file needed.

## API

### Image Conversion

```bash
curl -X POST https://webplow.example.com/ \
  -H "X-API-Key: YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -o output.webp
```

### Batch Conversion

The API handles one image per request. For batch conversion, send parallel requests from the client:

```bash
# 10 images in parallel
ls *.jpg | xargs -P 10 -I{} \
  curl -s -X POST https://webplow.example.com/ \
    -H "X-API-Key: YOUR_TOKEN" \
    -F "file=@{}" -o "{}.webp"
```

Server-side imgproxy supports 32 concurrent conversions. No batch API needed.

### Health Check

```bash
curl https://webplow.example.com/health
# {"status":"ok"}
```

## Access Logging

When `LOG_FILE` is set, each request is logged:

```json
{"time":"2026-02-06T04:15:32Z","user":"felix","file":"test.png","in_bytes":70,"status":200,"ms":2}
```

Query examples:

```bash
# View log
docker exec webplow-webplow-1 cat /data/access.log

# Per-user stats (requires jq on host)
docker exec webplow-webplow-1 cat /data/access.log | \
  jq -s 'group_by(.user) | map({user: .[0].user, count: length})'
```

## Architecture

```
External Server → Nginx (443/SSL) → webplow (Go, :9000) → imgproxy (libvips, :8080)
                                          │                        │
                                      data volume             uploads volume
                                      tokens.json             temp images
                                      access.log
```

- Nginx: host-native systemd, TLS termination + connection buffering
- webplow + imgproxy: Docker Compose, `docker compose up -d` manages both

## Deployment

### Docker Compose (Production)

```bash
docker compose up -d --build
# Token management
docker exec webplow-webplow-1 webplow-token add "site-name"
docker kill -s HUP webplow-webplow-1
```

### With Nginx

```nginx
upstream webplow {
    server 127.0.0.1:9000;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name webplow.example.com;

    ssl_certificate     /etc/letsencrypt/live/webplow.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/webplow.example.com/privkey.pem;

    client_max_body_size 20m;

    location / {
        proxy_pass http://webplow;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Local Development

```bash
cp .env.example .env
make token-add
make run
```

### Host Deployment (systemd, alternative)

```bash
make deploy
sudo /opt/webplow/webplow-token add "user1"
sudo systemctl restart webplow
```

## Make Targets

```bash
make build          # Build webplow + webplow-token
make run            # Run locally
make token-add      # Add token (interactive)
make token-list     # List all tokens
make token-delete   # Delete token (interactive)
make health         # Health check
make deploy         # Host deploy (systemd)
make docker-up      # Docker start
make docker-down    # Docker stop
make clean          # Clean build artifacts
```

## Project Structure

```
webplow/
├── cmd/
│   ├── server/main.go          # API server entry point
│   └── token/main.go           # Token management CLI
├── internal/
│   ├── auth/store.go           # Token store (JSON file)
│   ├── config/config.go        # Environment-based config
│   └── handler/handler.go      # HTTP handler
├── configs/config.yaml         # Config reference
├── deployments/webplow.service # systemd service (alternative)
├── scripts/deploy.sh           # Host deploy script
├── Dockerfile                  # Container build
├── docker-compose.yml          # Container orchestration (production)
├── .env.example                # Local dev environment variables
└── Makefile
```
