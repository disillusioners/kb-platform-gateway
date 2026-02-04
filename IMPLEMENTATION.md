# KB Platform Gateway - Implementation Update

## Changes Summary

This update implements production-ready features for KB Platform Gateway:

### 0. Environment Configuration
- **Added**: `.env` file support using `github.com/joho/godotenv`
- **Priority**: System env vars → `.env` file → Code defaults
- **Created**: `.env.example` template with all configuration options
- **Created**: `.gitignore` to prevent committing `.env` file with credentials

Configuration loading priority:
1. System environment variables (highest priority)
2. Values from `.env` file in working directory
3. Hard-coded defaults (lowest priority)

### 1. Authentication Changes
- **Removed**: JWT token generation and validation
- **Added**: Uses `x-user-name` header from upstream gateway (e.g., Envoy)
- **Removed**: `/api/v1/auth/login` endpoint
- **Updated**: Auth middleware now validates `x-user-name` header

### 2. Database Persistence
- **Added**: Repository pattern with interfaces for easy database switching
- **Implemented**: PostgreSQL repository for:
  - Documents (CRUD operations)
  - Conversations (list, create, get)
  - Messages (create, list by conversation)
- **Schema**: Created `schema.sql` for table setup

### 3. S3 Integration
- **Added**: AWS S3 SDK v2
- **Implemented**: Presigned URL generation for uploads/downloads
- **Features**:
  - Support for S3-compatible services (via Endpoint config)
  - Configurable presigned URL expiration

### 4. Temporal Workflow Integration
- **Added**: Temporal Go SDK
- **Implemented**: Workflow client for:
  - Upload workflow initiation
  - Index workflow initiation
  - Workflow status queries
  - Workflow cancellation

### 5. UUID Generation
- **Fixed**: Now uses `github.com/google/uuid` library
- **Before**: Hardcoded UUIDs causing duplicates

### 6. SSE Hub Cleanup
- **Removed**: Unused SSE Hub stub implementation
- **Result**: Cleaner codebase, actual SSE streaming still works via Python Core

## New Dependencies

```go
github.com/aws/aws-sdk-go-v2 v1.41.1
github.com/aws/aws-sdk-go-v2/config v1.32.7
github.com/aws/aws-sdk-go-v2/service/s3 v1.96.0
github.com/google/uuid v1.6.0
github.com/lib/pq v1.10.9
go.temporal.io/api v1.62.0
go.temporal.io/sdk v1.39.0
```

## Configuration

### Environment Variables

The gateway now supports `.env` file configuration. Simply:

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration values

3. The gateway will automatically load `.env` on startup

**Note**: System environment variables override `.env` values if both are set.

New environment variables:

### Database
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=kb_user
DB_PASSWORD=kb_password
DB_NAME=kb_platform
DB_SSLMODE=disable
```

### S3
```bash
S3_BUCKET=kb-documents
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=your-access-key
S3_SECRET_ACCESS_KEY=your-secret-key
S3_ENDPOINT=https://s3.amazonaws.com  # Optional for S3-compatible services
```

### Temporal
```bash
TEMPORAL_HOST=temporal
TEMPORAL_PORT=7233
TEMPORAL_NAMESPACE=default
```

### Python Core (existing)
```bash
PYTHON_CORE_HOST=python-llama-core
PYTHON_CORE_PORT=8000
```

### Server (existing)
```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=debug  # or release
```

## Database Setup

Run the schema to create tables:

```bash
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f schema.sql
```

## Architecture

```
┌──────────────┐         x-user-name        ┌──────────────────┐
│ Envoy/Gateway│ ◄──────────────────────── │ Go API Gateway  │
└──────────────┘                           └────────┬─────────┘
                                                   │
                                   ┌───────────────┼───────────────┐
                                   ▼               ▼               ▼
                            ┌─────────┐    ┌────────┐    ┌─────────────┐
                            │PostgreSQL│    │   S3   │    │  Temporal   │
                            └─────────┘    └────────┘    └─────────────┘
                                                   │
                                                   ▼
                                    ┌──────────────────────────────┐
                                    │   Python LlamaIndex Core   │
                                    └──────────────────────────────┘
```

## API Changes

### Removed Endpoints
- `POST /api/v1/auth/login` - No longer needed

### Updated Endpoints

#### Upload Document
```http
POST /api/v1/documents
Header: x-user-name: <username>
Form: file=<binary>
```

Response now includes real S3 presigned URL:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "upload_url": "https://s3.amazonaws.com/...",
  "s3_key": "documents/550e8400/...",
  "filename": "document.pdf",
  "file_size": 1048576,
  "status": "pending",
  "created_at": "2026-02-05T00:00:00Z"
}
```

#### Complete Upload
```http
POST /api/v1/documents/{id}/complete
Header: x-user-name: <username>
```

Now triggers Temporal upload workflow.

#### List Documents
```http
GET /api/v1/documents?status=complete&limit=50&offset=0
Header: x-user-name: <username>
```

Returns persisted documents from PostgreSQL.

#### List/Create Conversations
```http
GET /api/v1/conversations?limit=50&offset=0
POST /api/v1/conversations
Header: x-user-name: <username>
```

Operations now persisted to PostgreSQL.

## Repository Interface

The `repository.Repository` interface allows swapping database implementations:

```go
type Repository interface {
    DocumentRepository
    ConversationRepository
    MessageRepository
}
```

To use a different database:
1. Create a new struct implementing `Repository`
2. Initialize it in `main.go` instead of `NewPostgresRepository`

## Error Handling

All handlers return consistent error responses:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

Error codes:
- `VALIDATION_ERROR` - Invalid input
- `AUTHENTICATION_ERROR` - Missing x-user-name header
- `INTERNAL_ERROR` - Server error
- `NOT_FOUND` - Resource not found

## Quick Start

### Development Setup

1. Copy and configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. Run database schema:
   ```bash
   psql -h localhost -U kb_user -d kb_platform -f schema.sql
   ```

3. Build and run:
   ```bash
   go build -o bin/gateway cmd/main.go
   ./bin/gateway
   ```

### Docker Setup

```bash
docker build -t kb-platform-gateway .
docker run --env-file .env -p 8080:8080 kb-platform-gateway
```

## Testing

Builds project:
```bash
go build ./...
```

Run tests (when implemented):
```bash
go test ./...
```

## Configuration Files

- **`.env.example`**: Template for environment variables
- **`.gitignore`**: Prevents committing `.env` file
- **`schema.sql`**: Database schema for PostgreSQL

## Migration from Old Version

1. Update environment variables (add DB, S3, Temporal configs)
2. Run `schema.sql` on your PostgreSQL database
3. Update upstream gateway (Envoy) to:
   - Remove JWT generation
   - Add `x-user-name` header with decoded username
4. Update client applications to:
   - Remove `/auth/login` call
   - Include `x-user-name` header in requests (or rely on upstream gateway)
