# Multi-User Imgproxy API Gateway - Technical Specification

## 1. System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Rate Limiter  │    │  Auth Middleware│
│   (Nginx/HAProxy│────│  (Token Bucket) │────│   (JWT/API Key) │
│      Optional)  │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API Gateway (Go)                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Router    │  │ Middleware  │  │ Health Check│             │
│  │  (Gorilla)  │  │   Stack     │  │  Manager    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │Load Balancer│  │   Metrics   │  │   Config    │             │
│  │  (Weighted  │  │ Collector   │  │  Manager    │             │
│  │Round Robin) │  │             │  │             │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PostgreSQL Database                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │    Users    │  │   Servers   │  │   Metrics   │             │
│  │    Table    │  │    Table    │  │    Table    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Imgproxy Instances                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Imgproxy-1  │  │ Imgproxy-2  │  │ Imgproxy-N  │             │
│  │ (Weight: 3) │  │ (Weight: 2) │  │ (Weight: 1) │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## 2. Database Schema Design

```sql
-- Users table for authentication and rate limiting
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_hour INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Imgproxy servers configuration
CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    weight INTEGER DEFAULT 1,
    is_healthy BOOLEAN DEFAULT true,
    last_health_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    response_time_ms INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Request metrics and logging
CREATE TABLE request_metrics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    server_id INTEGER REFERENCES servers(id),
    request_path TEXT,
    response_status INTEGER,
    response_time_ms INTEGER,
    request_size_bytes BIGINT,
    response_size_bytes BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rate limiting buckets (for persistent rate limiting)
CREATE TABLE rate_limit_buckets (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    tokens_minute INTEGER DEFAULT 0,
    tokens_hour INTEGER DEFAULT 0,
    last_refill_minute TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_refill_hour TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_servers_healthy ON servers(is_healthy, weight);
CREATE INDEX idx_metrics_user_time ON request_metrics(user_id, created_at);
CREATE INDEX idx_rate_limit_user ON rate_limit_buckets(user_id);
```

## 3. API Endpoint Specifications

### Base URL: `/api/v1`

#### Authentication
- **Header**: `X-API-Key: <api_key>`
- **Alternative**: `Authorization: Bearer <api_key>`

#### Endpoints

```
GET /health
- Description: Gateway health check
- Response: 200 OK
- Body: {"status": "healthy", "servers": [...]}

GET /servers
- Description: List configured imgproxy servers
- Auth: Required
- Response: 200 OK
- Body: {"servers": [{"id": 1, "name": "server1", "healthy": true, "weight": 3}]}

POST /process
- Description: Process image through imgproxy
- Auth: Required
- Headers: Content-Type: application/json
- Body: {
    "url": "https://example.com/image.jpg",
    "operations": "resize:300:200/quality:80",
    "format": "webp"
  }
- Response: 200 OK (proxied imgproxy response)

GET /metrics
- Description: User-specific metrics
- Auth: Required
- Query: ?from=2024-01-01&to=2024-01-31
- Response: 200 OK
- Body: {"requests": 1500, "avg_response_time": 120, "errors": 5}

GET /status
- Description: Current rate limit status
- Auth: Required
- Response: 200 OK
- Body: {
    "rate_limit": {
      "minute": {"remaining": 45, "reset_at": "2024-01-01T12:01:00Z"},
      "hour": {"remaining": 850, "reset_at": "2024-01-01T13:00:00Z"}
    }
  }
```

## 4. Weighted Round-Robin Load Balancing Algorithm

```go
type Server struct {
    ID              int    `json:"id"`
    Name            string `json:"name"`
    URL             string `json:"url"`
    Weight          int    `json:"weight"`
    CurrentWeight   int    `json:"-"`
    IsHealthy       bool   `json:"is_healthy"`
    ResponseTimeMs  int    `json:"response_time_ms"`
}

type LoadBalancer struct {
    servers []*Server
    mutex   sync.RWMutex
}

func (lb *LoadBalancer) NextServer() *Server {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()
    
    if len(lb.servers) == 0 {
        return nil
    }
    
    // Filter healthy servers
    healthy := make([]*Server, 0)
    for _, server := range lb.servers {
        if server.IsHealthy {
            healthy = append(healthy, server)
        }
    }
    
    if len(healthy) == 0 {
        return nil
    }
    
    // Weighted round-robin selection
    var selected *Server
    totalWeight := 0
    
    for _, server := range healthy {
        server.CurrentWeight += server.Weight
        totalWeight += server.Weight
        
        if selected == nil || server.CurrentWeight > selected.CurrentWeight {
            selected = server
        }
    }
    
    if selected != nil {
        selected.CurrentWeight -= totalWeight
    }
    
    return selected
}
```

## 5. Rate Limiting Implementation

### Token Bucket Algorithm

```go
type TokenBucket struct {
    capacity     int64
    tokens       int64
    refillRate   int64
    lastRefill   time.Time
    mutex        sync.Mutex
}

func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
    return &TokenBucket{
        capacity:   capacity,
        tokens:     capacity,
        refillRate: refillRate,
        lastRefill: time.Now(),
    }
}

func (tb *TokenBucket) Allow() bool {
    tb.mutex.Lock()
    defer tb.mutex.Unlock()
    
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)
    
    // Refill tokens
    tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate
    tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
    tb.lastRefill = now
    
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    
    return false
}

type RateLimiter struct {
    buckets map[string]*TokenBucket
    mutex   sync.RWMutex
    db      *sql.DB
}

func (rl *RateLimiter) CheckLimit(userID string, minuteLimit, hourLimit int64) error {
    rl.mutex.RLock()
    minuteBucket, exists := rl.buckets[userID+":minute"]
    rl.mutex.RUnlock()
    
    if !exists {
        rl.mutex.Lock()
        minuteBucket = NewTokenBucket(minuteLimit, minuteLimit/60)
        rl.buckets[userID+":minute"] = minuteBucket
        rl.mutex.Unlock()
    }
    
    if !minuteBucket.Allow() {
        return errors.New("rate limit exceeded: minute")
    }
    
    // Similar logic for hour bucket...
    
    return nil
}
```

## 6. Health Check Mechanism

```go
type HealthChecker struct {
    servers   []*Server
    interval  time.Duration
    timeout   time.Duration
    db        *sql.DB
}

func (hc *HealthChecker) Start() {
    ticker := time.NewTicker(hc.interval)
    go func() {
        for range ticker.C {
            hc.checkAllServers()
        }
    }()
}

func (hc *HealthChecker) checkAllServers() {
    var wg sync.WaitGroup
    
    for _, server := range hc.servers {
        wg.Add(1)
        go func(s *Server) {
            defer wg.Done()
            hc.checkServer(s)
        }(server)
    }
    
    wg.Wait()
}

func (hc *HealthChecker) checkServer(server *Server) {
    start := time.Now()
    
    client := &http.Client{Timeout: hc.timeout}
    resp, err := client.Get(server.URL + "/health")
    
    duration := time.Since(start)
    isHealthy := err == nil && resp.StatusCode == 200
    
    // Update server status
    server.IsHealthy = isHealthy
    server.ResponseTimeMs = int(duration.Milliseconds())
    
    if err != nil {
        server.ErrorCount++
    }
    
    // Update database
    _, err = hc.db.Exec(`
        UPDATE servers 
        SET is_healthy = $1, response_time_ms = $2, error_count = $3, last_health_check = NOW()
        WHERE id = $4`,
        isHealthy, server.ResponseTimeMs, server.ErrorCount, server.ID)
    
    if resp != nil {
        resp.Body.Close()
    }
}
```

## 7. Configuration File Structure

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

database:
  host: "localhost"
  port: 5432
  name: "imgproxy_gateway"
  user: "postgres"
  password: "password"
  max_connections: 25
  max_idle_connections: 5
  connection_lifetime: "5m"

rate_limiting:
  default_minute_limit: 60
  default_hour_limit: 1000
  cleanup_interval: "1m"

health_check:
  interval: "30s"
  timeout: "5s"
  failure_threshold: 3

load_balancer:
  algorithm: "weighted_round_robin"
  
imgproxy_servers:
  - name: "primary"
    url: "http://imgproxy-1:8080"
    weight: 3
  - name: "secondary"
    url: "http://imgproxy-2:8080"
    weight: 2

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  retention_days: 30
```

```go
type Config struct {
    Server struct {
        Port         int           `yaml:"port"`
        Host         string        `yaml:"host"`
        ReadTimeout  time.Duration `yaml:"read_timeout"`
        WriteTimeout time.Duration `yaml:"write_timeout"`
        IdleTimeout  time.Duration `yaml:"idle_timeout"`
    } `yaml:"server"`
    
    Database struct {
        Host               string        `yaml:"host"`
        Port               int           `yaml:"port"`
        Name               string        `yaml:"name"`
        User               string        `yaml:"user"`
        Password           string        `yaml:"password"`
        MaxConnections     int           `yaml:"max_connections"`
        MaxIdleConnections int           `yaml:"max_idle_connections"`
        ConnectionLifetime time.Duration `yaml:"connection_lifetime"`
    } `yaml:"database"`
    
    RateLimiting struct {
        DefaultMinuteLimit int           `yaml:"default_minute_limit"`
        DefaultHourLimit   int           `yaml:"default_hour_limit"`
        CleanupInterval    time.Duration `yaml:"cleanup_interval"`
    } `yaml:"rate_limiting"`
    
    HealthCheck struct {
        Interval         time.Duration `yaml:"interval"`
        Timeout          time.Duration `yaml:"timeout"`
        FailureThreshold int           `yaml:"failure_threshold"`
    } `yaml:"health_check"`
    
    ImgproxyServers []struct {
        Name   string `yaml:"name"`
        URL    string `yaml:"url"`
        Weight int    `yaml:"weight"`
    } `yaml:"imgproxy_servers"`
}
```

## 8. Error Handling and HTTP Status Codes

```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e APIError) Error() string {
    return e.Message
}

// Standard error responses
var (
    ErrUnauthorized     = APIError{401, "Unauthorized", "Invalid or missing API key"}
    ErrRateLimitExceeded = APIError{429, "Rate Limit Exceeded", "Too many requests"}
    ErrServerUnavailable = APIError{503, "Service Unavailable", "No healthy imgproxy servers"}
    ErrBadRequest       = APIError{400, "Bad Request", "Invalid request parameters"}
    ErrInternalError    = APIError{500, "Internal Server Error", ""}
)

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
    var apiErr APIError
    
    switch e := err.(type) {
    case APIError:
        apiErr = e
    case *url.Error:
        apiErr = APIError{502, "Bad Gateway", "Upstream server error"}
    case context.DeadlineExceeded:
        apiErr = APIError{504, "Gateway Timeout", "Request timeout"}
    default:
        apiErr = ErrInternalError
        apiErr.Details = err.Error()
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(apiErr.Code)
    json.NewEncoder(w).Encode(apiErr)
}

// HTTP Status Code Mapping:
// 200 - Successful image processing
// 400 - Invalid request parameters
// 401 - Authentication failed
// 403 - Forbidden (user inactive)
// 429 - Rate limit exceeded
// 500 - Internal server error
// 502 - Imgproxy server error
// 503 - No healthy servers available
// 504 - Request timeout
```

## 9. Security Considerations

### API Key Management
```go
func generateAPIKey() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return base64.URLEncoding.EncodeToString(bytes)
}

func hashAPIKey(key string) string {
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:])
}
```

### Input Validation
```go
type ProcessRequest struct {
    URL        string `json:"url" validate:"required,url"`
    Operations string `json:"operations" validate:"required,max=500"`
    Format     string `json:"format" validate:"omitempty,oneof=jpeg png webp avif"`
}

func validateRequest(req *ProcessRequest) error {
    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        return err
    }
    
    // Additional URL validation
    if !isAllowedDomain(req.URL) {
        return errors.New("domain not allowed")
    }
    
    return nil
}
```

### Security Headers
```go
func securityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        next.ServeHTTP(w, r)
    })
}
```

## 10. Performance Optimization Strategies

### Connection Pooling
```go
func setupDatabase(config *Config) (*sql.DB, error) {
    db, err := sql.Open("postgres", buildDSN(config))
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(config.Database.MaxConnections)
    db.SetMaxIdleConns(config.Database.MaxIdleConnections)
    db.SetConnMaxLifetime(config.Database.ConnectionLifetime)
    
    return db, nil
}
```

### HTTP Client Optimization
```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  true, // Let imgproxy handle compression
    },
}
```

### Caching Strategy
```go
type CacheEntry struct {
    Data      []byte
    Headers   map[string]string
    ExpiresAt time.Time
}

type MemoryCache struct {
    entries map[string]*CacheEntry
    mutex   sync.RWMutex
    maxSize int64
    size    int64
}

func (c *MemoryCache) Get(key string) (*CacheEntry, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.entries[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry, true
}
```

### Metrics Collection
```go
type Metrics struct {
    RequestsTotal     prometheus.CounterVec
    RequestDuration   prometheus.HistogramVec
    ActiveConnections prometheus.Gauge
}

func initMetrics() *Metrics {
    return &Metrics{
        RequestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "gateway_requests_total",
                Help: "Total number of requests",
            },
            []string{"method", "status", "user_id"},
        ),
        RequestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "gateway_request_duration_seconds",
                Help: "Request duration in seconds",
            },
            []string{"method", "status"},
        ),
    }
}
```

## Implementation Notes

1. **Graceful Shutdown**: Implement proper signal handling for graceful server shutdown
2. **Circuit Breaker**: Add circuit breaker pattern for failing imgproxy servers
3. **Distributed Rate Limiting**: Consider Redis for distributed rate limiting across multiple gateway instances
4. **Monitoring**: Integrate with Prometheus/Grafana for comprehensive monitoring
5. **Logging**: Use structured logging with correlation IDs for request tracing
6. **Testing**: Implement comprehensive unit and integration tests
7. **Documentation**: Generate API documentation using OpenAPI/Swagger

This specification provides a solid foundation for implementing a production-ready multi-user imgproxy API gateway in Go.