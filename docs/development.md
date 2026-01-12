# Development Guide

This document covers the development workflow for the WebSocket test project.

## Prerequisites

- Go 1.21 or later
- Docker (for building container images)
- GoReleaser (for releases)
- kubectl (for Kubernetes deployment)

## Project Structure

```
websocket-test/
├── cmd/
│   ├── server/          # Server entry point
│   │   └── main.go
│   └── client/          # Client entry point
│       └── main.go
├── server/              # Server implementation
│   ├── server.go        # WebSocket server logic
│   └── metrics.go       # Prometheus metrics
├── client/              # Client implementation
│   └── client.go        # WebSocket client logic
├── k8s/                 # Kubernetes manifests
├── docs/                # Documentation
├── .github/workflows/   # CI/CD pipelines
├── Dockerfile.server    # Server container image
├── Dockerfile.client    # Client container image
├── .goreleaser.yaml     # GoReleaser configuration
└── go.mod              # Go module definition
```

## Building

### Building Binaries

```bash
# Build server
go build -o server ./cmd/server

# Build client
go build -o client ./cmd/client

# Build both (timezone data is automatically embedded via import)
go build -o server ./cmd/server
go build -o client ./cmd/client
```

### Building Docker Images Manually

```bash
# Build server image
go build -o server ./cmd/server
docker build -t websocket-test-server:dev -f Dockerfile.server .

# Build client image (multi-stage build copies CA certs from ubuntu)
go build -o client ./cmd/client
docker build -t websocket-test-client:dev -f Dockerfile.client .
```

### Building with GoReleaser

```bash
# Build snapshot (no release)
goreleaser release --snapshot --clean

# Full release (requires git tag)
git tag -a v0.1.0 -m "Release v0.1.0"
goreleaser release --clean
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run vet
go vet ./...
```

## Running Locally

### Terminal 1: Start the server

```bash
go run ./cmd/server --host=127.0.0.1 --port=3003 --metrics-port=9100
```

### Terminal 2: Start the client

```bash
go run ./cmd/client --host=127.0.0.1 --port=3003
```

### Terminal 3: Check metrics

```bash
curl http://localhost:9100/metrics | grep websocket
```

## Development Workflow

### Making Changes

1. Create a feature branch
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes

3. Run tests and linters
   ```bash
   go test ./...
   go vet ./...
   ```

4. Commit and push
   ```bash
   git add .
   git commit -m "Add my feature"
   git push origin feature/my-feature
   ```

5. Create a pull request

### Releasing

1. Ensure all tests pass
   ```bash
   go test ./...
   ```

2. Create and push a version tag
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

3. GitHub Actions will automatically:
   - Build binaries for multiple platforms
   - Create Docker images for amd64 and arm64
   - Push images to ghcr.io
   - Create a GitHub release with artifacts

## Continuous Integration

### CI Pipeline (`.github/workflows/ci.yaml`)

Runs on every push and pull request:
- Builds server and client binaries
- Runs tests
- Runs go vet

### Release Pipeline (`.github/workflows/release.yaml`)

Runs on version tags (e.g., `v0.1.0`):
- Builds binaries for multiple platforms
- Creates multi-arch Docker images
- Pushes images to GitHub Container Registry
- Creates GitHub release with artifacts

## Kubernetes Deployment

### Deploy to Kubernetes

```bash
# Deploy server
kubectl apply -f k8s/server-deployment.yaml
kubectl apply -f k8s/server-service.yaml

# Deploy client
kubectl apply -f k8s/client-deployment.yaml

# Check status
kubectl get pods
kubectl logs -l app=websocket-test-server
kubectl logs -l app=websocket-test-client
```

### Access Metrics in Kubernetes

```bash
# Port forward metrics
kubectl port-forward svc/websocket-test-server 9100:9100

# Check metrics
curl http://localhost:9100/metrics
```

## Code Organization

### Server Package (`server/`)

- `server.go`: Main WebSocket server implementation
  - Connection management
  - WebSocket upgrade handling
  - Periodic logging
  - Message echo logic

- `metrics.go`: Prometheus metrics
  - Metrics registration
  - Metrics server setup
  - Connection count gauge

### Client Package (`client/`)

- `client.go`: WebSocket client implementation
  - Connection management
  - Automatic reconnection
  - Keep-alive ping/pong
  - Status logging

## Dependencies

Key dependencies:
- `github.com/gorilla/websocket`: WebSocket implementation
- `github.com/prometheus/client_golang`: Prometheus metrics

To update dependencies:

```bash
go get -u ./...
go mod tidy
```

## GoReleaser Configuration

The `.goreleaser.yaml` file configures:

- Multi-platform binary builds (Linux, macOS, Windows)
- Multi-architecture support (amd64, arm64)
- Docker image builds with multi-arch manifests
- GitHub release creation
- Changelog generation
- Timezone data embedding via `import _ "time/tzdata"`

The timezone data import ensures binaries include timezone information for proper time handling in containers.

## Troubleshooting

### Build Issues

If you encounter certificate errors during `go get`:
```bash
# Update CA certificates
go clean -modcache
go mod download
```

### Docker Build Issues

Ensure binaries are built before Docker build:
```bash
CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
CGO_ENABLED=0 GOOS=linux go build -o client ./cmd/client
```

### Kubernetes Issues

Check logs for errors:
```bash
kubectl logs -l app=websocket-test-server --tail=50
kubectl logs -l app=websocket-test-client --tail=50
```

Check pod status:
```bash
kubectl describe pod -l app=websocket-test-server
kubectl describe pod -l app=websocket-test-client
```

