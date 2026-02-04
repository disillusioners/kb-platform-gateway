# Go API Gateway - API Specification

## Base URL

```
Production: https://api.kb-platform.example.com
Development: http://localhost:8080
```

## Authentication

All endpoints except `/healthz` and `/readyz` require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### Get JWT Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**Response (200 OK)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-05T11:30:00Z"
}
```

## Documents

### Upload Document

Initiates document upload with presigned S3 URL and starts Temporal workflow.

```http
POST /api/v1/documents
Content-Type: multipart/form-data
Authorization: Bearer <token>

file: <binary data>
```

**Response (200 OK)**:
```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440000",
  "upload_url": "https://s3.amazonaws.com/bucket/key?signature=...",
  "status": "pending",
  "expires_at": "2026-02-04T12:30:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid file type or size
- `401 Unauthorized`: Invalid or missing token
- `500 Internal Server Error`: Failed to generate URL or start workflow

### Complete Upload

Signals that file upload is complete and triggers indexing.

```http
POST /api/v1/documents/{document_id}/complete
Authorization: Bearer <token>
```

**Response (200 OK)**:
```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "indexing"
}
```

**Error Responses**:
- `404 Not Found`: Document not found
- `409 Conflict`: Document already completed or failed

### List Documents

Retrieves list of all documents with their status.

```http
GET /api/v1/documents
Authorization: Bearer <token>
```

**Query Parameters**:
- `status` (optional): Filter by status (`pending`, `indexing`, `complete`, `failed`)
- `limit` (optional): Number of results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response (200 OK)**:
```json
{
  "documents": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "filename": "document.pdf",
      "file_size": 1048576,
      "status": "complete",
      "created_at": "2026-02-03T10:00:00Z",
      "indexed_at": "2026-02-03T10:01:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

### Get Document

Retrieves metadata for a specific document.

```http
GET /api/v1/documents/{document_id}
Authorization: Bearer <token>
```

**Response (200 OK)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf",
  "file_size": 1048576,
  "status": "complete",
  "created_at": "2026-02-03T10:00:00Z",
  "indexed_at": "2026-02-03T10:01:00Z",
  "error_message": null
}
```

**Error Responses**:
- `404 Not Found`: Document not found

### Delete Document

Deletes a document and all associated data (S3, Qdrant, Postgres).

```http
DELETE /api/v1/documents/{document_id}
Authorization: Bearer <token>
```

**Response (204 No Content)**

**Error Responses**:
- `404 Not Found`: Document not found
- `500 Internal Server Error`: Failed to delete from one or more services

## Conversations

### List Conversations

Retrieves all conversations.

```http
GET /api/v1/conversations
Authorization: Bearer <token>
```

**Query Parameters**:
- `limit` (optional): Number of results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response (200 OK)**:
```json
{
  "conversations": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "created_at": "2026-02-03T11:00:00Z",
      "updated_at": "2026-02-03T11:05:00Z",
      "message_count": 5
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

### Create Conversation

Creates a new conversation.

```http
POST /api/v1/conversations
Authorization: Bearer <token>
```

**Response (201 Created)**:
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "created_at": "2026-02-03T11:00:00Z",
  "updated_at": "2026-02-03T11:00:00Z"
}
```

### Get Conversation Messages

Retrieves all messages in a conversation.

```http
GET /api/v1/conversations/{conversation_id}/messages
Authorization: Bearer <token>
```

**Response (200 OK)**:
```json
{
  "messages": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "role": "user",
      "content": "What is LlamaIndex?",
      "timestamp": "2026-02-03T11:00:00Z",
      "metadata": {}
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440003",
      "role": "assistant",
      "content": "LlamaIndex is a data framework...",
      "timestamp": "2026-02-03T11:00:01Z",
      "metadata": {}
    }
  ]
}
```

**Error Responses**:
- `404 Not Found`: Conversation not found

## Queries

### Query (Streaming)

Performs a RAG query and streams response via SSE.

```http
POST /api/v1/query
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "What is LlamaIndex?",
  "conversation_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Response (200 OK, text/event-stream)**:
```
event: message
data: {"type":"start","id":"880e8400-e29b-41d4-a716-446655440004"}

event: chunk
data: {"content":"LlamaIndex is"}

event: chunk
data: {"content":" a data framework"}

event: message
data: {"type":"end","id":"880e8400-e29b-41d4-a716-446655440004"}
```

**Request Body**:
- `query` (string, required): The user query
- `conversation_id` (string, optional): Existing conversation ID. If not provided, creates new conversation.

**Error Responses**:
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Invalid or missing token
- `500 Internal Server Error`: Query processing failed

## Health Checks

### Health Check

```http
GET /healthz
```

**Response (200 OK)**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-04T11:30:00Z"
}
```

### Readiness Check

```http
GET /readyz
```

**Response (200 OK)**:
```json
{
  "status": "ready",
  "dependencies": {
    "python_core": "ok",
    "temporal": "ok",
    "opa": "ok"
  }
}
```

**Response (503 Service Unavailable)**:
```json
{
  "status": "not_ready",
  "dependencies": {
    "python_core": "error: connection refused",
    "temporal": "ok",
    "opa": "ok"
  }
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `AUTHENTICATION_ERROR` | 401 | Invalid or missing authentication |
| `AUTHORIZATION_ERROR` | 403 | Authorization denied |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists or invalid state |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service unavailable or dependent service down |
| `TIMEOUT` | 504 | Gateway timeout from backend service |

## Rate Limiting

Currently not implemented. Planned for future releases.

## Pagination

List endpoints use cursor-based pagination via `limit` and `offset` parameters.

- Maximum `limit`: 100
- Default `limit`: 50
- Default `offset`: 0
