# Quick Reference Guide - ImgProxy Multi-User API Gateway

## 🎯 30-Second Overview

Transform single-user imgproxy API → Multi-tenant gateway with:
- ✅ Multiple users with API keys
- ✅ Load balancing across multiple backends (weighted)
- ✅ 3-level rate limiting (IP/User/API)
- ✅ Health checks & auto-failover
- ✅ IP whitelisting per user
- ✅ Quota management
- ✅ Prometheus metrics

## 📂 Document Guide

| Document | Purpose | Read When |
|----------|---------|-----------|
| **PROJECT_SUMMARY.md** | High-level overview | Start here first |
| **REQUIREMENTS.md** | What to build | Planning & stakeholder review |
| **TECHNICAL_SPECIFICATION.md** | How to build it | During implementation |
| **IMPLEMENTATION_TASKS.md** | Step-by-step tasks | Daily development work |

## 🔄 Development Workflow

```
1. Read PROJECT_SUMMARY.md (15 min)
   ↓
2. Review REQUIREMENTS.md (30 min)
   ↓
3. Study TECHNICAL_SPECIFICATION.md (1 hour)
   ↓
4. Start IMPLEMENTATION_TASKS.md Phase 1
   ↓
5. Implement → Test → Deploy each phase
```

## 🏗️ Architecture at a Glance

```
Client Request
    ↓
[API Gateway - Go]
    ├─ Auth Middleware (API Key validation)
    ├─ Rate Limiter (IP/User/API levels)
    ├─ IP Whitelist Check
    ├─ Quota Check
    └─ Load Balancer (Weighted Round-Robin)
        ↓
    [Imgproxy Backends]
    ├─ Server 1 (Weight: 70)
    ├─ Server 2 (Weight: 30)
    └─ Server N (Weight: 10)

Data Stores:
├─ PostgreSQL (Users, Config, Logs)
├─ Redis (Rate Limits, Cache)
└─ Prometheus (Metrics)
```

## 📊 Key Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Response Time (p95) | < 500ms | TBD |
| Throughput | 1000 req/s | TBD |
| Availability | 99.9% | TBD |
| Concurrent Users | 100+ | 1 |

## 🔑 Core Components

### 1. Authentication
```go
// API Key format: imgproxy_<user_id>_<random>
X-API-Key: imgproxy_user123_a8e77f84e2729dc6728889d4be31b123
```

### 2. Load Balancing
```yaml
backends:
  - url: http://127.0.0.1:48080
    weight: 70  # Handles 70% of traffic
  - url: http://127.0.0.1:48081
    weight: 30  # Handles 30% of traffic
```

### 3. Rate Limiting
```yaml
rate_limits:
  ip_level: 100/min      # Per IP address
  user_level: 1000/min   # Per user
  api_level: 5000/min    # Global
```

### 4. Health Checks
```
Every 30s → GET /health
  ├─ 3 failures → Mark unhealthy
  └─ 2 successes → Mark healthy
```

## 📋 Implementation Checklist

### Phase 1: Foundation (Week 1-3)
- [ ] F1.1 - Project structure
- [ ] F1.2 - Configuration management
- [ ] F1.3 - Database schema
- [ ] F1.4 - Database connection
- [ ] F1.5 - Basic HTTP server

### Phase 2: Core Features (Week 4-7)
- [ ] C2.1 - User management models
- [ ] C2.2 - API key authentication
- [ ] C2.3 - Backend configuration
- [ ] C2.4 - Health check system
- [ ] C2.5 - Weighted load balancer
- [ ] C2.6 - Image processing handler

### Phase 3: Advanced Features (Week 8-10)
- [ ] A3.1 - Multi-level rate limiting
- [ ] A3.2 - IP whitelisting
- [ ] A3.3 - Quota management
- [ ] A3.4 - Admin API endpoints
- [ ] A3.5 - Configuration hot-reload

### Phase 4: Observability (Week 11-12)
- [ ] O4.1 - Prometheus metrics
- [ ] O4.2 - Structured logging
- [ ] O4.3 - Request tracing
- [ ] O4.4 - Monitoring dashboard

### Phase 5: Testing & Deployment (Week 13)
- [ ] T5.1 - Integration tests
- [ ] T5.2 - Load testing
- [ ] T5.3 - Security audit
- [ ] T5.4 - Documentation
- [ ] T5.5 - Production deployment

## 🚀 Quick Start Commands

```bash
# 1. Setup database
psql -U postgres -c "CREATE DATABASE imgproxy_gateway;"
psql -U postgres imgproxy_gateway < migrations/001_initial_schema.sql

# 2. Configure environment
cp .env.example .env
# Edit .env with your settings

# 3. Run migrations
go run cmd/migrate/main.go

# 4. Start server
go run cmd/server/main.go

# 5. Test health endpoint
curl http://localhost:8080/api/v1/health

# 6. Test with API key
curl -X POST http://localhost:8080/api/v1/upload \
  -H "X-API-Key: your-api-key" \
  -F "file=@image.jpg"
```

## 🔧 Configuration Template

```yaml
# config.yaml
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
    health_check_interval: 30s
    timeout: 5s
  - url: "http://127.0.0.1:48081"
    weight: 30
    health_check_interval: 30s
    timeout: 5s

rate_limits:
  ip_level:
    requests_per_minute: 100
    burst: 20
  user_level:
    requests_per_minute: 1000
    requests_per_hour: 10000
  api_level:
    requests_per_minute: 5000

quotas:
  default:
    monthly_requests: 100000
    monthly_bandwidth_mb: 10240

observability:
  metrics_enabled: true
  prometheus_path: "/metrics"
  log_level: "info"
  log_format: "json"
```

## 📡 API Endpoints

### Image Processing
```
POST /api/v1/upload
  Headers: X-API-Key
  Body: multipart/form-data (file)
  Response: image/webp
```

### Management
```
GET  /api/v1/health          # System health
GET  /api/v1/ready           # Readiness probe
GET  /api/v1/users/{id}      # User details
PUT  /api/v1/users/{id}      # Update user
GET  /api/v1/users/{id}/quota # Quota usage
GET  /api/v1/backends        # Backend status
GET  /api/v1/metrics         # JSON metrics
GET  /metrics                # Prometheus metrics
```

## 🔒 Security Checklist

- [ ] API keys stored with bcrypt hashing
- [ ] TLS 1.3 enabled for production
- [ ] Rate limiting configured
- [ ] IP whitelisting tested
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (parameterized queries)
- [ ] CORS configured properly
- [ ] Security headers set (HSTS, CSP, etc.)

## 📊 Monitoring Queries

### Prometheus Queries
```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Response time (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Backend health
backend_health_status

# Rate limit violations
rate(user_rate_limit_violations[5m])
```

## 🐛 Troubleshooting

### Issue: High latency
```bash
# Check backend health
curl http://localhost:8080/api/v1/backends

# Check metrics
curl http://localhost:8080/api/v1/metrics | jq '.backend_response_times'
```

### Issue: Rate limit errors
```bash
# Check user rate limits
curl http://localhost:8080/api/v1/users/{id} | jq '.rate_limits'

# Check Redis connection
redis-cli ping
```

### Issue: Authentication failures
```bash
# Verify API key in database
psql imgproxy_gateway -c "SELECT * FROM api_keys WHERE key_hash = crypt('your-key', key_hash);"
```

## 📚 Learning Resources

### Go Libraries Used
- `github.com/gorilla/mux` - HTTP routing
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/go-redis/redis/v8` - Redis client
- `github.com/prometheus/client_golang` - Metrics
- `github.com/spf13/viper` - Configuration
- `go.uber.org/zap` - Structured logging

### Algorithms Implemented
- **Weighted Round-Robin**: Load distribution
- **Token Bucket**: Rate limiting
- **Circuit Breaker**: Health checks
- **CIDR Matching**: IP whitelisting

## 🎓 Best Practices Applied

1. **Minimal Code**: Only essential features
2. **Star Topology**: Subagents coordinated centrally
3. **Separation of Concerns**: Middleware pattern
4. **Fail Fast**: Early validation
5. **Graceful Degradation**: Fallback to healthy backends
6. **Observability First**: Metrics from day one

## 📞 Getting Help

1. Check `TECHNICAL_SPECIFICATION.md` for implementation details
2. Review `IMPLEMENTATION_TASKS.md` for task dependencies
3. Refer to code comments in generated files
4. Test incrementally after each task

---

**Quick Tip**: Start with Phase 1, Task F1.1. Don't skip ahead!

**Estimated Timeline**: 8-13 weeks for full implementation

**Current Status**: Planning Complete ✅ | Implementation Ready 🚀
