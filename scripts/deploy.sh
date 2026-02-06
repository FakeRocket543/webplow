#!/bin/bash
set -e

APP_NAME="webplow"
BUILD_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_DIR="/opt/webplow"

echo "=== 建置 ==="
cd "$BUILD_DIR"
CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow ./cmd/server
CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow-token ./cmd/token
echo "✓ 建置完成"

echo "=== 部署到 $DEPLOY_DIR ==="
sudo mkdir -p "$DEPLOY_DIR"
sudo cp webplow webplow-token "$DEPLOY_DIR/"

if [ ! -f "$DEPLOY_DIR/.env" ]; then
    sudo cp .env.example "$DEPLOY_DIR/.env"
    sudo sed -i "s|TOKEN_FILE=tokens.json|TOKEN_FILE=$DEPLOY_DIR/tokens.json|" "$DEPLOY_DIR/.env"
fi

if [ ! -f "$DEPLOY_DIR/tokens.json" ]; then
    echo "[]" | sudo tee "$DEPLOY_DIR/tokens.json" > /dev/null
    echo "⚠ 請執行 $DEPLOY_DIR/webplow-token add <name> 建立第一組 token"
fi

sudo chown -R www-data:www-data "$DEPLOY_DIR"
sudo chmod 600 "$DEPLOY_DIR/.env" "$DEPLOY_DIR/tokens.json"
echo "✓ 部署完成"

echo "=== 安裝 systemd 服務 ==="
sudo cp deployments/webplow.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable webplow
echo "✓ 服務已安裝"

echo "=== 啟動服務 ==="
sudo systemctl restart webplow
sleep 1
sudo systemctl status webplow --no-pager
echo "✓ 完成！"
