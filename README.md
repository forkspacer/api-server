# Forkspacer API Server

REST API server for managing Forkspacer workspaces and modules in Kubernetes environments.

## Overview

The Forkspacer API Server provides a REST API for creating, managing, and orchestrating workspace and module resources within Kubernetes clusters. It acts as a bridge between client applications and the Forkspacer Kubernetes operator.

## Features

- **Workspace Management**: Full CRUD operations for Forkspacer workspaces
- **Module Management**: Deploy and manage modules within workspaces
- **Kubeconfig Secret Management**: Store and manage Kubernetes connection credentials
- **Auto-hibernation Support**: Configure automatic workspace hibernation schedules
- **OpenAPI Documentation**: Interactive API documentation at `/api/v1/docs`
- **Comprehensive Validation**: DNS-compliant naming, YAML validation, and business rules

## Quick Start

### Prerequisites

- Go 1.25+
- Kubernetes cluster (v1.34+)
- Forkspacer CRDs installed in the cluster

### Installation

```bash
git clone https://github.com/forkspacer/api-server.git
cd api-server
go mod download
```

### Configuration

Configure using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DEV` | `true` | Enable development mode |
| `API_PORT` | `8421` | HTTP server port |

### Running

**Development:**
```bash
make dev
```

**Production:**
```bash
DEV=false go run ./cmd/main.go
```

**Docker:**
```bash
make docker-build
docker run -p 8421:8421 ghcr.io/forkspacer/api-server:v1.0.0
```

## API Documentation

Once running, visit:
- **Interactive Docs**: http://localhost:8421/api/v1/docs
- **OpenAPI Spec**: http://localhost:8421/api/v1/openapi.yaml

## Development

**Format and lint:**
```bash
make fmt
make lint
```

**Run tests:**
```bash
go test ./...
```

## Docker

```bash
# Build
make docker-build

# Multi-platform build
make docker-buildx PLATFORMS=linux/amd64,linux/arm64

# Push
make docker-push IMG=your-registry/api-server:tag
```

## Project Structure

```
cmd/          # Application entry point
pkg/
  api/        # HTTP API layer (handlers, routing, validation)
  services/   # Business logic and Kubernetes operations
  config/     # Configuration management
  utils/      # Utility functions
```

## Response Format

**Success:**
```json
{
  "success": {
    "code": "ok|created|deleted",
    "data": { ... }
  }
}
```

**Error:**
```json
{
  "error": {
    "code": "bad_request|...",
    "data": "Error details"
  }
}
```

## License

Part of the Forkspacer ecosystem.

## Support

- Issues: https://github.com/forkspacer/api-server/issues
- Docs: http://localhost:8421/api/v1/docs