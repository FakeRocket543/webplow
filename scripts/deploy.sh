#!/bin/bash
set -e

APP_NAME="webplow"
BUILD_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_DIR="/opt/webplow"

echo "=== Building ==="
cd "$BUILD_DIR"
CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow ./cmd/server
CGO_ENABLED=0 go build -ldflags="-s -w" -o webplow-token ./cmd/token
echo "✓ Build complete"

echo "=== Deploying to $DEPLOY_DIR ==="
sudo mkdir -p "$DEPLOY_DIR"
sudo cp webplow webplow-token "$DEPLOY_DIR/"

if [ ! -f "$DEPLOY_DIR/.env" ]; then
    sudo cp .env.example "$DEPLOY_DIR/.env"
    sudo sed -i "s|TOKEN_FILE=tokens.json|TOKEN_FILE=$DEPLOY_DIR/tokens.json|" "$DEPLOY_DIR/.env"
fi

if [ ! -f "$DEPLOY_DIR/tokens.json" ]; then
    echo "[]" | sudo tee "$DEPLOY_DIR/tokens.json" > /dev/null
    echo "⚠ Run $DEPLOY_DIR/webplow-token add <name> to create your first token"
fi

sudo chown -R www-data:www-data "$DEPLOY_DIR"
sudo chmod 600 "$DEPLOY_DIR/.env" "$DEPLOY_DIR/tokens.json"
echo "✓ Deploy complete"

echo "=== Installing systemd service ==="
sudo cp deployments/webplow.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable webplow
echo "✓ Service installed"

echo "=== Starting service ==="
sudo systemctl restart webplow
sleep 1
sudo systemctl status webplow --no-pager
echo "✓ Done!"
