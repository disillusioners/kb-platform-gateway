# KB Platform - Go API Gateway

## Overview
This service acts as the main entry point for the Knowledge Base Platform. It handles:
- HTTP REST API for clients (including Flutter)
- SSE (Server-Sent Events) for streaming RAG responses
- JWT Authentication
- Request routing to the Python Core Service via gRPC

## Directory Structure
- `cmd/server`: Entrypoint for the application
- `internal/`: Private service code (handlers, middleware, business logic)
- `pkg/`: Public shared code (if any)

## Tech Stack
- Go 1.21+
- Gin/Echo (TBD)
- gRPC
