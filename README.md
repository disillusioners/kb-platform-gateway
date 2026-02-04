# KB Platform - Go API Gateway

## Overview
This service acts as the main entry point for the Knowledge Base Platform. It handles:
- HTTP REST API for clients (including Flutter)
- SSE (Server-Sent Events) for streaming RAG responses
- Request routing to the Python Core Service via HTTP
- Authentication via `x-user-name` header (from upstream gateway)
- Document upload/download via S3
- Workflow orchestration via Temporal
- Data persistence via PostgreSQL

## Directory Structure
- `cmd/`: Application entry point
- `internal/`: Private service code (handlers, middleware, business logic)
  - `api/`: HTTP layer (handlers, routes, middleware)
  - `config/`: Configuration management
  - `models/`: Data models
  - `repository/`: Database abstraction layer
  - `services/`: External service clients (S3, Temporal, Python Core)

## Tech Stack
- Go 1.23+
- Gin Web Framework
- PostgreSQL (lib/pq)
- AWS S3 SDK v2
- Temporal Go SDK
- Zerolog (structured logging)
- Godotenv (environment configuration)

## Quick Start

### 1. Configuration

Copy and configure environment variables:

```bash
cp .env.example .env
# Edit .env with your settings
```

### 2. Database Setup

Run the schema:

```bash
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f schema.sql
```

### 3. Run

```bash
# Build
go build -o bin/gateway cmd/main.go

# Run
./bin/gateway
```

Or use Docker:

```bash
docker build -t kb-platform-gateway .
docker run --env-file .env -p 8080:8080 kb-platform-gateway
```

## Configuration

Environment variables are loaded from (highest to lowest priority):
1. System environment variables
2. `.env` file in working directory
3. Code defaults

See `.env.example` for all available configuration options.

## API Endpoints

### Health Checks
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check (verifies dependencies)

### Documents
- `POST /api/v1/documents` - Upload document (requires `x-user-name`)
- `GET /api/v1/documents` - List documents (requires `x-user-name`)
- `GET /api/v1/documents/:id` - Get document (requires `x-user-name`)
- `DELETE /api/v1/documents/:id` - Delete document (requires `x-user-name`)
- `POST /api/v1/documents/:id/complete` - Complete upload (requires `x-user-name`)

### Conversations
- `GET /api/v1/conversations` - List conversations (requires `x-user-name`)
- `POST /api/v1/conversations` - Create conversation (requires `x-user-name`)
- `GET /api/v1/conversations/:id/messages` - Get messages (requires `x-user-name`)

### Queries
- `POST /api/v1/query` - Query RAG system with SSE streaming (requires `x-user-name`)

For full API documentation, see [API.md](API.md).

## Development

### Dependencies

```bash
go mod download
go mod tidy
```

### Build

```bash
go build ./...
```

### Run Tests

```bash
go test ./...
```

## Documentation

- [API.md](API.md) - Full API specification
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - Implementation details and changes
- [schema.sql](schema.sql) - Database schema
