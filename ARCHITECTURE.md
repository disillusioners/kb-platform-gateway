# Go API Gateway - Architecture

## Service Design

The Gateway service serves as the API entry point for the Knowledge Base platform, handling HTTP requests, authentication, and routing to backend services.

## Core Responsibilities

### 1. HTTP API Layer
- RESTful API endpoints for all client interactions
- Request validation and error handling
- Response formatting

### 2. Authentication & Authorization
- JWT token generation and validation
- API key management
- Integration with OPA for policy enforcement

### 3. Streaming Support
- Server-Sent Events (SSE) for query responses
- Real-time streaming from Python core to client
- Connection management and timeout handling

### 4. Request Routing
- Forwarding requests to Python LlamaIndex core
- gRPC or REST communication with internal services
- Service discovery integration

### 5. Observability
- Request logging
- Health and readiness checks
- Error tracking

## Architecture Diagram

```
┌─────────────┐     HTTP/SSE     ┌──────────────────┐     gRPC/HTTP     ┌──────────────────────┐
│   Client    │ ◄──────────────► │  Go API Gateway   │ ◄───────────────► │ Python LlamaIndex    │
│  (Flutter)  │                  │                   │                  │      Core Service     │
└─────────────┘                  └──────────────────┘                  └──────────────────────┘
                                         │
                                         │ OPA
                                         ▼
                                   ┌──────────────┐
                                   │ OPA Server   │
                                   └──────────────┘
```

## Component Architecture

### HTTP Layer
```
HTTP Request → Middleware Chain → Handler → Service Client → Response
                           │
                           ├── Auth Middleware (JWT)
                           ├── Logging Middleware
                           ├── CORS Middleware
                           └── OPA Authorization
```

### Streaming Layer (SSE)
```
Client Request → SSE Handler → Subscribe to Core Stream → Forward Events → Client
```

### Auth Flow
```
1. Client sends credentials to /auth/login
2. Gateway validates credentials
3. Gateway generates JWT token
4. Gateway returns token to client
5. Client includes token in Authorization header
6. Gateway validates token on each request
7. Gateway forwards valid requests to services
```

## Request Flow Examples

### Document Upload
```
1. Client: POST /api/v1/documents (multipart form)
2. Gateway: Validate JWT
3. Gateway: Generate S3 presigned URL
4. Gateway: Start Temporal UploadWorkflow
5. Gateway: Return presigned URL to client
6. Client: Upload directly to S3
7. Client: POST /api/v1/documents/{id}/complete
8. Gateway: Send signal to Temporal workflow
```

### Query (Streaming)
```
1. Client: POST /api/v1/query (SSE)
2. Gateway: Validate JWT
3. Gateway: Forward to Python core
4. Python core: Perform RAG query
5. Python core: Stream response chunks via SSE
6. Gateway: Forward chunks to client
7. Client: Display streaming response
```

## Configuration Management

### Configuration Sources
1. Environment variables (primary)
2. Config files (for development)

### Key Configuration
```go
type Config struct {
    Server   ServerConfig
    JWT      JWTConfig
    Services ServicesConfig
}

type ServerConfig struct {
    Host string
    Port int
}

type JWTConfig struct {
    Secret     string
    Expiration time.Duration
}

type ServicesConfig struct {
    PythonCore string
    Temporal   string
}
```

## Security

### Authentication
- JWT tokens signed with HS256
- Token expiration: 24h (configurable)
- No refresh tokens (simplified for single-user system)

### Authorization
- OPA policies for fine-grained access control
- Default deny-all policy
- Namespace-based isolation in K8S

### Transport Security
- HTTPS in production
- Internal communication via service mesh (optional)

## Error Handling

### Error Categories
1. **Client Errors (4xx)**
   - 400 Bad Request: Invalid input
   - 401 Unauthorized: Missing/invalid JWT
   - 403 Forbidden: Authorization failed
   - 404 Not Found: Resource not found
   - 429 Too Many Requests: Rate limited (future)

2. **Server Errors (5xx)**
   - 500 Internal Server Error: Unexpected error
   - 502 Bad Gateway: Python core unavailable
   - 503 Service Unavailable: Temporal unavailable
   - 504 Gateway Timeout: Timeout from backend

### Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request format",
    "details": {...}
  }
}
```

## Performance Considerations

### Resource Usage
- Memory: ~50Mi base + 1Mi per active connection
- CPU: <10m idle, scales with request rate

### Optimization Strategies
1. Connection pooling to Python core
2. SSE connection reuse where possible
3. Async request processing
4. Response compression (gzip)

### Scalability
- Horizontal scaling via K8S replicas
- Stateless design enables easy scaling
- Load balancer distributes connections

## Deployment Considerations

### Container Requirements
- Base image: `golang:1.21-alpine`
- Multi-stage build for small image size
- Binary only in final image

### K8S Configuration
- Probes:
  - Liveness: `/healthz` (every 10s)
  - Readiness: `/readyz` (every 5s)
- Resources:
  - Request: 128Mi RAM, 100m CPU
  - Limit: 256Mi RAM, 200m CPU
- Replicas: 2 minimum

### Network Policies
- Allow ingress from Envoy Gateway
- Allow egress to Python Llama Core
- Allow egress to OPA server
- Deny all other traffic

## Future Enhancements

### Planned Features
1. Rate limiting
2. Request caching
3. gRPC streaming as alternative to SSE
4. WebSocket support for bidirectional streaming
5. API versioning
6. Swagger/OpenAPI documentation

### Potential Optimizations
1. Connection pooling optimization
2. Response streaming from backend
3. Load testing and benchmarking
4. Metrics export (Prometheus)
