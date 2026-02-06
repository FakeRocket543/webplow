# Webplow

透過 imgproxy 將上傳圖片轉換為 WebP 格式的 API 服務，支援多用戶 token 管理。

## 快速開始（Docker Compose）

```bash
docker compose up -d --build

# 建立第一組 token
docker exec webp_api-webplow-1 webplow-token add "user1"

# 測試
curl -X POST https://webplow.lcn.tw/ \
  -H "X-API-Key: YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -o output.webp
```

## Token 管理

```bash
# 新增
docker exec webp_api-webplow-1 webplow-token add "site-name"

# 列出
docker exec webp_api-webplow-1 webplow-token list

# 刪除
docker exec webp_api-webplow-1 webplow-token delete <key>

# 新增/刪除 token 後需重啟載入
docker compose restart webplow
```

本地開發時也可用 `make token-add` / `make token-list` / `make token-delete`。

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
| `LOG_FILE` | （空，不記錄） | 用量記錄檔路徑 |

Docker Compose 部署時環境變數已在 `docker-compose.yml` 中設定，不需要 `.env`。

## API

### 圖片轉換

```bash
curl -X POST https://webplow.lcn.tw/ \
  -H "X-API-Key: YOUR_TOKEN" \
  -F "file=@image.jpg" \
  -o output.webp
```

### 健康檢查

```bash
curl https://webplow.lcn.tw/health
# {"status":"ok"}
```

## 用量記錄

設定 `LOG_FILE` 後，每筆請求會記錄：

```json
{"time":"2026-02-06T04:15:32Z","user":"felix","file":"test.png","in_bytes":70,"status":200,"ms":2}
```

查詢範例：

```bash
# 查看 log
docker exec webp_api-webplow-1 cat /data/access.log

# 各用戶統計（需要主機上有 jq）
docker exec webp_api-webplow-1 cat /data/access.log | \
  jq -s 'group_by(.user) | map({user: .[0].user, count: length})'
```

## 架構

```
外部 Server → Nginx (443/SSL) → webplow (Go, :9000) → imgproxy (libvips, :8080)
                                      │                        │
                                  data volume             uploads volume
                                  tokens.json             暫存圖片
                                  access.log
```

- Nginx：主機原生 systemd，TLS 終止 + 連線緩衝
- webplow + imgproxy：Docker Compose，`docker compose up -d` 一鍵管理

## 部署方式

### Docker Compose（生產環境）

```bash
docker compose up -d --build
# Token 管理
docker exec webp_api-webplow-1 webplow-token add "site-name"
docker compose restart webplow
```

### 搭配 Nginx

```nginx
upstream webplow {
    server 127.0.0.1:9000;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name webplow.lcn.tw;

    ssl_certificate     /etc/letsencrypt/live/webplow.lcn.tw/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/webplow.lcn.tw/privkey.pem;

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

### 本地開發

```bash
cp .env.example .env
make token-add
make run
```

### 主機部署（systemd，備選）

```bash
make deploy
sudo /opt/webplow/webplow-token add "user1"
sudo systemctl restart webplow
```

## Make 指令

```bash
make build          # 編譯 webplow + webplow-token
make run            # 本地執行
make token-add      # 新增 token（互動式）
make token-list     # 列出所有 token
make token-delete   # 刪除 token（互動式）
make health         # 健康檢查
make deploy         # 主機部署（systemd）
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
├── deployments/webplow.service # systemd 服務（備選）
├── scripts/deploy.sh           # 主機部署腳本
├── Dockerfile                  # 容器建置
├── docker-compose.yml          # 容器編排（生產環境）
├── .env.example                # 本地開發用環境變數
└── Makefile
```
