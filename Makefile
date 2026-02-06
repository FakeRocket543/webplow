.PHONY: build run test deploy clean docker-up docker-down

# Build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow ./cmd/server
	CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow-token ./cmd/token

# Run locally
run:
	go run ./cmd/server

# Token management
token-add:
	@read -p "Name: " name; go run ./cmd/token add "$$name"

token-list:
	go run ./cmd/token list

token-delete:
	@read -p "Key: " key; go run ./cmd/token delete "$$key"

# Reload tokens (Docker, no restart)
reload:
	docker kill -s HUP webplow-webplow-1

# Test conversion
test:
	@echo "Usage: curl -X POST http://127.0.0.1:9000/ -H 'X-API-Key: <token>' -F 'file=@test.jpg' -o output.webp"

# Health check
health:
	curl -s http://127.0.0.1:9000/health | python3 -m json.tool

# Host deploy
deploy:
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh

# Docker
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# Clean
clean:
	rm -f webplow webplow-token output.webp
