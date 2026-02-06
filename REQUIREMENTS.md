# ImgProxy Multi-User API Gateway Requirements

## 1. Executive Summary

Transform the existing single-user imgproxy-api Go program into a scalable, multi-tenant API gateway that supports multiple users, backend servers, and comprehensive observability.

**Current State:**
- Single API key: `a8e77f84e2729dc6728889d4be31b123`
- Single backend: `http://127.0.0.1:48080`
- Basic file upload and WebP conversion
- No user management or advanced features

**Target State:**
- Multi-user authentication system
- Load-balanced backend pool
- Per-user access controls and quotas
- Enterprise-grade observability

## 2. Functional Requirements

### 2.1 Multi-User Authentication (FR-001)

**FR-001.1: API Key Management**
- Generate unique API keys per user (minimum 32 characters, cryptographically secure)
- Support API key rotation without service interruption
- Store API keys with secure hashing (bcrypt/scrypt)
- API key format: `imgproxy_<user_id>_<random_string>`

**FR-001.2: User Management**
- Create, read, update, delete user accounts
- User attributes: ID, name, email, created_at, updated_at, status
- User status: active, suspended, deleted
- Admin interface for user management

**FR-001.3: Authentication Flow**
- Validate API key on every request via `X-API-Key` header
- Return appropriate HTTP status codes (401, 403)
- Support multiple API keys per user (primary/secondary)

### 2.2 Backend Server Management (FR-002)

**FR-002.1: Multiple Backend Support**
- Configure multiple imgproxy backend servers
- Backend attributes: URL, weight, status, health_check_url
- Minimum 2 backends, maximum 10 backends per deployment

**FR-002.2: Weighted Load Balancing**
- Implement weighted round-robin algorithm
- Support weight values 1-100 per backend
- Automatic failover to healthy backends
- Sticky sessions not required (stateless)

**FR-002.3: Backend Configuration**
```yaml
backends:
  - url: "http://127.0.0.1:48080"
    weight: 70
    health_check_path: "/health"
  - url: "http://127.0.0.1:48081" 
    weight: 30
    health_check_path: "/health"
```

### 2.3 Health Monitoring (FR-003)

**FR-003.1: Backend Health Checks**
- HTTP health checks every 30 seconds
- Configurable timeout (default: 5 seconds)
- Mark backend unhealthy after 3 consecutive failures
- Mark backend healthy after 2 consecutive successes

**FR-003.2: Health Check Endpoints**
- Backend health: `GET /health` → 200 OK
- Gateway health: `GET /api/v1/health` → JSON status
- Readiness probe: `GET /api/v1/ready`

### 2.4 Access Control (FR-004)

**FR-004.1: IP Whitelisting**
- Per-user IP whitelist configuration
- Support CIDR notation (e.g., `192.168.1.0/24`)
- Default: allow all IPs if whitelist empty
- Block requests from non-whitelisted IPs with 403

**FR-004.2: User Configuration**
```json
{
  "user_id": "user_123",
  "api_keys": ["imgproxy_user123_abc..."],
  "ip_whitelist": ["192.168.1.0/24", "10.0.0.1"],
  "rate_limits": {...},
  "quotas": {...}
}
```

### 2.5 Rate Limiting (FR-005)

**FR-005.1: Multi-Level Rate Limiting**
- IP-level: requests per IP address
- User-level: requests per API key/user
- API-level: global requests per endpoint

**FR-005.2: Rate Limit Configuration**
```yaml
rate_limits:
  ip_level:
    requests_per_minute: 100
    burst: 20
  user_level:
    requests_per_minute: 1000
    requests_per_hour: 10000
  api_level:
    requests_per_minute: 5000
```

**FR-005.3: Rate Limit Headers**
- `X-RateLimit-Limit`: limit value
- `X-RateLimit-Remaining`: remaining requests
- `X-RateLimit-Reset`: reset timestamp
- Return 429 when limits exceeded

### 2.6 Quota Management (FR-006)

**FR-006.1: User Quotas**
- Monthly request quota per user
- Monthly bandwidth quota (upload + download)
- Storage quota for cached/processed images
- Reset quotas on configurable schedule

**FR-006.2: Quota Tracking**
- Real-time quota consumption tracking
- Quota usage API endpoint: `GET /api/v1/users/{id}/quota`
- Block requests when quota exceeded (402 Payment Required)

**FR-006.3: Quota Configuration**
```json
{
  "monthly_requests": 100000,
  "monthly_bandwidth_mb": 10240,
  "storage_quota_mb": 1024,
  "reset_day": 1
}
```

### 2.7 API Endpoints (FR-007)

**FR-007.1: Image Processing (Existing)**
- `POST /api/v1/upload` - Upload and process image
- `GET /api/v1/convert` - Convert image format
- Maintain backward compatibility with existing endpoints

**FR-007.2: Management APIs**
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `GET /api/v1/users/{id}/quota` - Get quota usage
- `GET /api/v1/backends` - List backend status
- `GET /api/v1/metrics` - Get system metrics

## 3. Non-Functional Requirements

### 3.1 Performance (NFR-001)

- **Response Time**: 95th percentile < 500ms for image processing
- **Throughput**: Support 1000 concurrent requests
- **Availability**: 99.9% uptime (8.76 hours downtime/year)
- **Scalability**: Horizontal scaling via load balancer

### 3.2 Security (NFR-002)

- **Authentication**: Secure API key validation
- **Authorization**: Role-based access control
- **Data Protection**: TLS 1.3 for all communications
- **Input Validation**: Sanitize all user inputs
- **Rate Limiting**: Prevent abuse and DoS attacks

### 3.3 Observability (NFR-003)

**NFR-003.1: Metrics Collection**
- Request count, duration, status codes
- Backend health and response times
- User quota consumption
- Rate limit violations
- Error rates and types

**NFR-003.2: Metrics Endpoints**
- Prometheus metrics: `GET /metrics`
- JSON metrics: `GET /api/v1/metrics`

**NFR-003.3: Key Metrics**
```
# Request metrics
http_requests_total{method, endpoint, status, user_id}
http_request_duration_seconds{method, endpoint}

# Backend metrics  
backend_requests_total{backend_url, status}
backend_health_status{backend_url}
backend_response_time_seconds{backend_url}

# User metrics
user_requests_total{user_id}
user_quota_usage{user_id, quota_type}
user_rate_limit_violations{user_id, limit_type}

# System metrics
active_connections
memory_usage_bytes
cpu_usage_percent
```

**NFR-003.4: Logging**
- Structured JSON logging
- Log levels: DEBUG, INFO, WARN, ERROR
- Request/response logging with correlation IDs
- Security event logging (auth failures, rate limits)

### 3.4 Configuration (NFR-004)

**NFR-004.1: Configuration Management**
- YAML configuration files
- Environment variable overrides
- Hot-reload for non-critical settings
- Configuration validation on startup

**NFR-004.2: Sample Configuration**
```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: "postgres"
  dsn: "postgres://user:pass@localhost/imgproxy_gateway"

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

observability:
  metrics_enabled: true
  prometheus_path: "/metrics"
  log_level: "info"
```

## 4. Technical Architecture

### 4.1 Components

- **API Gateway**: Main HTTP server and request router
- **Authentication Service**: API key validation and user management
- **Load Balancer**: Backend selection and health checking
- **Rate Limiter**: Multi-level rate limiting engine
- **Quota Manager**: Usage tracking and enforcement
- **Metrics Collector**: Observability and monitoring
- **Configuration Manager**: Settings and hot-reload

### 4.2 Data Storage

- **Primary Database**: PostgreSQL for user data, quotas, configuration
- **Cache Layer**: Redis for rate limiting, session data
- **Metrics Storage**: Prometheus for time-series metrics

### 4.3 Dependencies

- **Go**: 1.21+ for main application
- **PostgreSQL**: 14+ for persistent storage
- **Redis**: 7+ for caching and rate limiting
- **Prometheus**: Metrics collection (optional)

## 5. Migration Strategy

### 5.1 Phase 1: Core Infrastructure
- Multi-user authentication system
- Database schema and user management
- Basic load balancing

### 5.2: Phase 2: Advanced Features  
- Rate limiting and quota management
- IP whitelisting
- Health checks

### 5.3: Phase 3: Observability
- Metrics collection
- Monitoring dashboards
- Alerting rules

### 5.4: Backward Compatibility
- Existing API key `a8e77f84e2729dc6728889d4be31b123` migrated as admin user
- Existing endpoints maintain same behavior
- Gradual migration path for existing clients

## 6. Success Criteria

- [ ] Support 100+ concurrent users
- [ ] 99.9% uptime with multiple backends
- [ ] Sub-500ms response times for 95% of requests
- [ ] Complete observability with Prometheus metrics
- [ ] Zero-downtime deployments
- [ ] Comprehensive rate limiting and quota enforcement
- [ ] Secure multi-tenant architecture

## 7. Out of Scope

- Image caching/CDN functionality
- Advanced image processing features
- User billing/payment integration
- Multi-region deployment
- GraphQL API support
- WebSocket connections
