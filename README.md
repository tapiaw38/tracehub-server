# TraceHub Server

**TraceHub Server** is a high-performance backend for collecting, storing, and analyzing traces/logs from distributed systems. Built with Go, PostgreSQL, and Gin framework following hexagonal architecture principles.

## 🏗️ Architecture

TraceHub follows a clean hexagonal architecture (Ports & Adapters):

```
tracehub-server/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/              # Business entities (Project, Trace, Error)
│   ├── usecases/            # Business logic
│   ├── adapters/
│   │   ├── web/             # HTTP handlers (Gin)
│   │   ├── datasources/     # Database repositories
│   │   └── websocket/       # Real-time streaming
│   └── platform/
│       ├── config/          # Configuration management
│       ├── database/        # Database connection
│       ├── auth/            # API key authentication
│       └── errors/          # Error handling
├── migrations/              # Database migrations
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## ✨ Features

- **Trace Ingestion**: High-throughput trace/log collection with batch support
- **Real-time Querying**: Filter and search traces by level, service, timestamp, etc.
- **PostgreSQL Storage**: Efficient indexed storage with automatic cleanup
- **API Key Authentication**: Secure project-based access control
- **Docker Ready**: Fully containerized with docker-compose
- **Auto Migrations**: Database schema managed with golang-migrate
- **Hexagonal Architecture**: Clean separation of concerns for maintainability

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone and configure**:
```bash
git clone https://github.com/tapiaw38/tracehub-server.git
cd tracehub-server
cp .env.example .env
# Edit .env with your settings
```

2. **Start the stack**:
```bash
make docker-up
```

The server will be available at `http://localhost:8080`

3. **Check health**:
```bash
curl http://localhost:8080/api/v1/health
```

## 📡 API Reference

### Projects

#### Create Project
```bash
POST /api/v1/projects
Content-Type: application/json

{
  "name": "my-backend",
  "description": "Production backend service",
  "language": "javascript",
  "repo_url": "https://github.com/user/repo"
}
```

#### List Projects
```bash
GET /api/v1/projects?limit=50&offset=0
```

### Traces

#### Ingest Single Trace
```bash
POST /api/v1/traces?project_id=<uuid>
Authorization: Bearer th_your_api_key
Content-Type: application/json

{
  "level": "error",
  "message": "Database connection failed",
  "timestamp": "2026-01-10T12:00:00Z",
  "source": "database.js:45",
  "stack_trace": "Error: Connection refused...",
  "context": {
    "user_id": "123",
    "request_id": "abc-def"
  },
  "service_name": "api-server",
  "environment": "production"
}
```

#### Query Traces
```bash
GET /api/v1/traces?project_id=<uuid>&level=error&limit=100
Authorization: Bearer th_your_api_key
```

## 🔑 Authentication

TraceHub uses API key authentication. Include the API key in the `Authorization` header:

```bash
Authorization: Bearer th_your_api_key
```

## 🛠️ Development

```bash
make help           # Show all commands
make build          # Build binary
make run            # Run locally
make dev            # Run with hot-reload
make docker-up      # Start Docker stack
make test           # Run tests
```

## 🤝 Integration

### JavaScript/Node.js SDK
See: https://github.com/tapiaw38/tracehub-javascript-sdk

### CLI Client
See: https://github.com/tapiaw38/tracehub-cli

## 📝 License

MIT License

---

Built with ❤️ using Go, PostgreSQL, and Gin
