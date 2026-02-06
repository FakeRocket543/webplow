# ImgProxy Multi-User API Gateway - Project Structure

## 📁 Proposed Directory Structure

```
imgproxy-gateway/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Main application entry point
│   └── migrate/
│       └── main.go                 # Database migration tool
│
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration loading and validation
│   │   └── config_test.go
│   │
│   ├── models/
│   │   ├── user.go                 # User model
│   │   ├── backend.go              # Backend server model
│   │   ├── quota.go                # Quota model
│   │   └── api_key.go              # API key model
│   │
│   ├── db/
│   │   ├── db.go                   # Database connection pool
│   │   ├── migrate.go              # Migration runner
│   │   └── queries.go              # SQL queries
│   │
│   ├── auth/
│   │   ├── auth.go                 # Authentication logic
│   │   ├── api_key.go              # API key validation
│   │   └── middleware.go           # Auth middleware
│   │
│   ├── loadbalancer/
│   │   ├── loadbalancer.go         # Weighted round-robin implementation
│   │   ├── backend.go              # Backend management
│   │   └── health.go               # Health check system
│   │
│   ├── ratelimit/
│   │   ├── ratelimit.go            # Rate limiting engine
│   │   ├── token_bucket.go         # Token bucket algorithm
│   │   ├── ip_limiter.go           # IP-level limiter
│   │   ├── user_limiter.go         # User-level limiter
│   │   └── middleware.go           # Rate limit middleware
│   │
│   ├── quota/
│   │   ├── quota.go                # Quota management
│   │   ├── tracker.go              # Usage tracking
│   │   └── middleware.go           # Quota check middleware
│   │
│   ├── ipwhitelist/
│   │   ├── whitelist.go            # IP whitelist logic
│   │   ├── cidr.go                 # CIDR matching
│   │   └── middleware.go           # IP check middleware
│   │
│   ├── metrics/
│   │   ├── metrics.go              # Prometheus metrics
│   │   ├── collector.go            # Custom collectors
│   │   └── middleware.go           # Metrics middleware
│   │
│   ├── handlers/
│   │   ├── upload.go               # Image upload handler
│   │   ├── health.go               # Health check handler
│   │   ├── users.go                # User management handlers
│   │   ├── backends.go             # Backend status handlers
│   │   └── metrics.go              # Metrics endpoint handler
│   │
│   ├── middleware/
│   │   ├── logging.go              # Request logging
│   │   ├── recovery.go             # Panic recovery
│   │   ├── cors.go                 # CORS handling
│   │   └── request_id.go           # Request ID generation
│   │
│   └── server/
│       ├── server.go               # HTTP server setup
│       ├── router.go               # Route definitions
│       └── shutdown.go             # Graceful shutdown
│
├── migrations/
│   ├── 001_initial_schema.sql      # Initial database schema
│   ├── 002_add_api_keys.sql        # API keys table
│   ├── 003_add_quotas.sql          # Quotas table
│   └── 004_add_indexes.sql         # Performance indexes
│
├── configs/
│   ├── config.yaml                 # Default configuration
│   ├── config.dev.yaml             # Development config
│   └── config.prod.yaml            # Production config
│
├── scripts/
│   ├── setup.sh                    # Initial setup script
│   ├── migrate.sh                  # Migration helper
│   └── test.sh                     # Test runner
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile              # Application container
│   │   └── docker-compose.yml      # Local development stack
│   │
│   └── kubernetes/
│       ├── deployment.yaml         # K8s deployment
│       ├── service.yaml            # K8s service
│       └── configmap.yaml          # K8s config
│
├── docs/
│   ├── api/
│   │   ├── openapi.yaml            # OpenAPI specification
│   │   └── postman.json            # Postman collection
│   │
│   └── architecture/
│       ├── diagrams/               # Architecture diagrams
│       └── decisions/              # ADRs (Architecture Decision Records)
│
├── tests/
│   ├── integration/
│   │   ├── auth_test.go            # Auth integration tests
│   │   ├── loadbalancer_test.go    # Load balancer tests
│   │   └── ratelimit_test.go       # Rate limit tests
│   │
│   └── load/
│       ├── k6_script.js            # k6 load test script
│       └── locust_test.py          # Locust load test
│
├── .env.example                    # Environment variables template
├── .gitignore                      # Git ignore rules
├── go.mod                          # Go module definition
├── go.sum                          # Go module checksums
├── Makefile                        # Build and test commands
└── README.md                       # Project documentation
```

## 📦 Package Organization

### `/cmd`
Application entry points. Each subdirectory is a separate executable.

### `/internal`
Private application code. Cannot be imported by other projects.

### `/migrations`
Database schema migrations in SQL format.

### `/configs`
Configuration file templates for different environments.

### `/scripts`
Scripts for setup, deployment, and maintenance.

### `/deployments`
Container and orchestration configurations.

### `/docs`
Project documentation, API specs, and architecture diagrams.

### `/tests`
Integration and load tests (unit tests live next to code).

## 🔧 Key Files

### Configuration Files

**config.yaml**
```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: postgres
  dsn: "postgres://user:pass@localhost/imgproxy_gateway"
  max_connections: 25

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

backends:
  - url: "http://127.0.0.1:48080"
    weight: 70
  - url: "http://127.0.0.1:48081"
    weight: 30

rate_limits:
  ip_level:
    requests_per_minute: 100
  user_level:
    requests_per_minute: 1000
```

**.env.example**
```bash
# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@localhost/imgproxy_gateway

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Observability
LOG_LEVEL=info
METRICS_ENABLED=true
```

**Makefile**
```makefile
.PHONY: build test run migrate docker-up docker-down

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

run:
	go run cmd/server/main.go

migrate:
	go run cmd/migrate/main.go

docker-up:
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker/docker-compose.yml down

lint:
	golangci-lint run

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
```

## 🗄️ Database Schema Files

**migrations/001_initial_schema.sql**
```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- API keys table
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);

-- Backends table
CREATE TABLE backends (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    weight INTEGER DEFAULT 1,
    is_healthy BOOLEAN DEFAULT true,
    last_health_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rate limits table
CREATE TABLE rate_limits (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    requests_per_minute INTEGER DEFAULT 1000,
    requests_per_hour INTEGER DEFAULT 10000,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quotas table
CREATE TABLE quotas (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    monthly_requests INTEGER DEFAULT 100000,
    monthly_bandwidth_mb INTEGER DEFAULT 10240,
    current_requests INTEGER DEFAULT 0,
    current_bandwidth_mb INTEGER DEFAULT 0,
    reset_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- IP whitelist table
CREATE TABLE ip_whitelist (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    ip_cidr VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Request logs table
CREATE TABLE request_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    backend_id INTEGER REFERENCES backends(id),
    request_path TEXT,
    response_status INTEGER,
    response_time_ms INTEGER,
    request_size_bytes BIGINT,
    response_size_bytes BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_backends_healthy ON backends(is_healthy, weight);
CREATE INDEX idx_ip_whitelist_user ON ip_whitelist(user_id);
CREATE INDEX idx_request_logs_user_time ON request_logs(user_id, created_at);
```

## 🐳 Docker Configuration

**deployments/docker/docker-compose.yml**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: imgproxy_gateway
      POSTGRES_USER: imgproxy
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus

  imgproxy-1:
    image: darthsim/imgproxy:latest
    environment:
      IMGPROXY_LOCAL_FILESYSTEM_ROOT: /var/www/imgproxy/uploads
    ports:
      - "48080:8080"
    volumes:
      - ../../uploads:/var/www/imgproxy/uploads

  imgproxy-2:
    image: darthsim/imgproxy:latest
    environment:
      IMGPROXY_LOCAL_FILESYSTEM_ROOT: /var/www/imgproxy/uploads
    ports:
      - "48081:8080"
    volumes:
      - ../../uploads:/var/www/imgproxy/uploads

  gateway:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - imgproxy-1
      - imgproxy-2
    environment:
      DB_DSN: postgres://imgproxy:secret@postgres/imgproxy_gateway
      REDIS_ADDR: redis:6379
    volumes:
      - ../../configs/config.yaml:/app/config.yaml

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
```

## 📊 Monitoring Configuration

**deployments/docker/prometheus.yml**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'imgproxy-gateway'
    static_configs:
      - targets: ['gateway:8080']
    metrics_path: '/metrics'
```

## 🧪 Test Structure

### Unit Tests
Located next to source files:
```
internal/loadbalancer/
├── loadbalancer.go
└── loadbalancer_test.go
```

### Integration Tests
Located in `/tests/integration/`:
```go
// tests/integration/auth_test.go
func TestAuthenticationFlow(t *testing.T) {
    // Test complete auth flow
}
```

### Load Tests
Located in `/tests/load/`:
```javascript
// tests/load/k6_script.js
import http from 'k6/http';
import { check } from 'k6';

export default function() {
    const res = http.post('http://localhost:8080/api/v1/upload', ...);
    check(res, { 'status is 200': (r) => r.status === 200 });
}
```

## 🚀 Build and Run Commands

```bash
# Setup
make setup              # Initialize project
make migrate            # Run database migrations

# Development
make run                # Run server locally
make test               # Run all tests
make lint               # Run linter

# Docker
make docker-up          # Start all services
make docker-down        # Stop all services

# Production
make build              # Build binary
make docker-build       # Build Docker image
```

## 📝 Code Organization Principles

### 1. Separation of Concerns
- Each package has a single responsibility
- Clear boundaries between layers

### 2. Dependency Injection
- Dependencies passed as parameters
- Easy to test and mock

### 3. Interface-Based Design
- Define interfaces for key components
- Enable testing and flexibility

### 4. Minimal Coupling
- Packages don't depend on each other unnecessarily
- Use interfaces for cross-package communication

### 5. Standard Project Layout
- Follows Go community conventions
- Easy for new developers to navigate

## 🎯 Implementation Order

1. **Phase 1**: Setup structure, config, database
2. **Phase 2**: Core features (auth, load balancing)
3. **Phase 3**: Advanced features (rate limiting, quotas)
4. **Phase 4**: Observability (metrics, logging)
5. **Phase 5**: Testing and deployment

## 📚 Additional Resources

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

**Note**: This structure follows Go best practices and is designed for scalability and maintainability.
