# Webplow

透過 imgproxy 將上傳圖片轉換為 WebP 格式的 API 服務，支援多用戶 token 管理。

## 快速開始

```bash
# 1. 複製環境設定
cp .env.example .env

# 2. 建立第一組 token
make token-add
# 或直接: go run ./cmd/token add "user1"

# 3. 啟動
make run
```

## Token 管理

```bash
# 新增 token
./webplow-token add "user1"
# Token created for "user1":
# a1b2c3d4e5f6...

# 列出所有 token
./webplow-token list
# NAME   KEY                                       CREATED
# user1  a1b2c3d4e5f6...                           2026-02-06 02:28

# 刪除 token
./webplow-token delete a1b2c3d4e5f6...

# 開發時也可用 make
make token-add
make token-list
make token-delete
```

Token 存放在 `tokens.json`（可透過 `TOKEN_FILE` 環境變數指定路徑）。

## 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `TOKEN_FILE` | `tokens.json` | Token 存放路徑 |
| `LISTEN_ADDR` | `127.0.0.1:9000` | 監聽地址 |
| `IMGPROXY_URL` | `http://127.0.0.1:48080` | imgproxy 後端地址 |
| `TEMP_DIR` | `/var/www/imgproxy/uploads` | 暫存目錄 |
| `MAX_FILE_SIZE` | `20971520` | 上傳大小上限（bytes） |
| `READ_TIMEOUT` | `30s` | 讀取超時 |
| `WRITE_TIMEOUT` | `60s` | 寫入超時 |

## API 使用

### 圖片轉換

```bash
curl -X POST http://127.0.0.1:9000/ \
  -H "X-API-Key: YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -o output.webp
```

### 健康檢查

```bash
curl http://127.0.0.1:9000/health
# {"status":"ok"}
```

## 部署方式

### Docker Compose（推薦）

```bash
cp .env.example .env
docker compose up -d --build
# 進入容器建立 token
docker compose exec webplow webplow-token add "user1"
```

### 主機部署（systemd）

```bash
make deploy
# 首次部署會提示建立 token
sudo /opt/webplow/webplow-token add "user1"
sudo systemctl restart webplow
```

### 搭配 Nginx（面對外部流量時建議）

```nginx
upstream webplow {
    server 127.0.0.1:9000;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name img.example.com;

    ssl_certificate     /etc/letsencrypt/live/img.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/img.example.com/privkey.pem;

    client_max_body_size 20m;

    location / {
        proxy_pass http://webplow;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

內網呼叫可省略 Nginx，直接連 Go 服務即可。

## Make 指令

```bash
make build          # 編譯 webplow + webplow-token
make run            # 本地執行
make token-add      # 新增 token（互動式）
make token-list     # 列出所有 token
make token-delete   # 刪除 token（互動式）
make health         # 健康檢查
make deploy         # 主機部署
make docker-up      # Docker 啟動
make docker-down    # Docker 停止
make clean          # 清理產出物
```

## 專案結構

```
webplow/
├── cmd/
│   ├── server/main.go          # API 服務進入點
│   └── token/main.go           # Token 管理 CLI
├── internal/
│   ├── auth/store.go           # Token 存取（JSON 檔）
│   ├── config/config.go        # 環境變數配置
│   └── handler/handler.go      # HTTP handler
├── configs/config.yaml         # 配置參考文件
├── deployments/webplow.service # systemd 服務
├── scripts/deploy.sh           # 部署腳本
├── Dockerfile                  # 容器建置
├── docker-compose.yml          # 容器編排
├── .env.example                # 環境變數範本
└── Makefile
```
