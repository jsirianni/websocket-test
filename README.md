# WebSocket Test

A simple WebSocket server and client for testing WebSocket connections, built with Go and Gorilla WebSocket.

## Features

- **Server**: WebSocket server with Prometheus metrics
- **Client**: Auto-reconnecting WebSocket client
- **Docker**: Minimal "from scratch" container images
- **Kubernetes**: Ready-to-use deployment manifests
- **Multi-arch**: Supports amd64 and arm64 architectures

## Quick Start

### Build and Run

```bash
# Build binaries
make build

# Terminal 1: Start server
./bin/server

# Terminal 2: Start client
./bin/client

# Terminal 3: Check metrics
curl http://localhost:9100/metrics | grep websocket
```

### Using Docker

```bash
# Create network
docker network create websocket-test

# Run server
docker run -d --name websocket-server --network websocket-test \
  -p 3003:3003 -p 9100:9100 \
  ghcr.io/jsirianni/websocket-test-server:latest

# Run client
docker run -d --name websocket-client --network websocket-test \
  ghcr.io/jsirianni/websocket-test-client:latest \
  --host=websocket-server

# View logs
docker logs -f websocket-server
docker logs -f websocket-client
```

### Using Kubernetes

```bash
kubectl apply -f k8s/
kubectl logs -f -l app=websocket-test-server
kubectl logs -f -l app=websocket-test-client
```

For detailed setup instructions, see the [Quick Start Guide](docs/quickstart.md).

## Configuration

### Server

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `0.0.0.0` | Server bind address |
| `--port` | `3003` | WebSocket port |
| `--metrics-port` | `9100` | Prometheus metrics port |

### Client

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `localhost` | Server hostname |
| `--port` | `3003` | Server port |
| `--connections` | `1` | Number of concurrent connections |
| `--tls` | `false` | Use TLS (wss scheme) for WebSocket connection |

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Quick Start Guide](docs/quickstart.md)** - Get up and running in minutes
- **[Server Documentation](docs/server.md)** - Server configuration and usage
- **[Client Documentation](docs/client.md)** - Client configuration and usage
- **[Metrics and Monitoring](docs/metrics.md)** - Prometheus metrics and alerting
- **[Development Guide](docs/development.md)** - Building, testing, and contributing
- **[Architecture](docs/architecture.md)** - Design and implementation details

## Building

Build binaries:
```bash
go build ./cmd/server
go build ./cmd/client
```

Build with GoReleaser:
```bash
goreleaser release --snapshot --clean
```

See the [Development Guide](docs/development.md) for more details.

## License

MIT License - see LICENSE file for details.

