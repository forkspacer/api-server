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
- Kubernetes cluster (v1.20+)
- Forkspacer operator and CRDs installed in the cluster
  - If you haven't set up the operator yet, follow the [Development Guide](https://forkspacer.com/development/overview) to run it in dev mode

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
| `KUBECONFIG` | `~/.kube/config` | Path to Kubernetes config file |

**Kubernetes Connection:**

The API server connects to Kubernetes using your kubeconfig. It supports:
- **Local development**: Uses `KUBECONFIG` or `~/.kube/config`
- **In-cluster**: Automatically detects when running inside a Kubernetes pod

**RBAC Requirements:**

The API server requires permissions to manage Forkspacer resources. When running locally, it uses your current kubeconfig context's credentials.

### Running

**Development:**
```bash
make dev
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

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Support

- Issues: https://github.com/forkspacer/api-server/issues
- Docs: http://localhost:8421/api/v1/docs