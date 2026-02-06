.PHONY: build run test deploy clean docker-up docker-down

# 建置
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow ./cmd/server
	CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow-token ./cmd/token

# 本地執行
run:
	go run ./cmd/server

# Token 管理
token-add:
	@read -p "Name: " name; go run ./cmd/token add "$$name"

token-list:
	go run ./cmd/token list

token-delete:
	@read -p "Key: " key; go run ./cmd/token delete "$$key"

# 測試轉換
test:
	@echo "Usage: curl -X POST http://127.0.0.1:9000/ -H 'X-API-Key: <token>' -F 'file=@test.jpg' -o output.webp"

# 健康檢查
health:
	curl -s http://127.0.0.1:9000/health | python3 -m json.tool

# 部署到主機
deploy:
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh

# Docker
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# 清理
clean:
	rm -f webplow webplow-token output.webp
