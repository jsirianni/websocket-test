# WebSocket Server

The WebSocket server accepts WebSocket connections and provides Prometheus metrics.

## Features

- WebSocket server with configurable host and port
- Connection logging with remote address
- Periodic connection count logging (every 30 seconds)
- Prometheus metrics endpoint
- Graceful shutdown handling

## Configuration

All configuration is done via command-line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `0.0.0.0` | Host address to bind the WebSocket server to |
| `--port` | `3003` | Port for the WebSocket server |
| `--metrics-port` | `9100` | Port for the Prometheus metrics endpoint |

## Usage

### Running Locally

```bash
# Run with default settings
./server

# Run with custom host and port
./server --host=127.0.0.1 --port=8080 --metrics-port=9090
```

### Running in Docker

```bash
docker run -p 3003:3003 -p 9100:9100 ghcr.io/jsirianni/websocket-test-server:latest
```

### Running in Kubernetes

```bash
kubectl apply -f k8s/server-deployment.yaml
kubectl apply -f k8s/server-service.yaml
```

## Endpoints

### WebSocket Endpoint

- **URL**: `ws://<host>:<port>/ws`
- **Protocol**: WebSocket
- **Authentication**: None

The server accepts all incoming WebSocket connections and echoes back any messages received.

### Metrics Endpoint

- **URL**: `http://<host>:<metrics-port>/metrics`
- **Format**: Prometheus text format

#### Metrics Exposed

| Metric | Type | Description |
|--------|------|-------------|
| `websocket_connections` | Gauge | Current number of active WebSocket connections |

## Logging

The server logs the following events:

- Server startup with address
- WebSocket connection established (with remote address)
- WebSocket disconnection (with remote address)
- Periodic connection count (every 30 seconds)
- Errors during connection handling

Example log output:

```
2026/01/12 10:00:00 Starting WebSocket server on 0.0.0.0:3003
2026/01/12 10:00:00 Starting metrics server on :9100
2026/01/12 10:00:05 WebSocket connected: 192.168.1.100:54321
2026/01/12 10:00:35 Active connections: 1
2026/01/12 10:01:05 Active connections: 1
2026/01/12 10:01:20 WebSocket disconnected: 192.168.1.100:54321
```

## Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals for graceful shutdown:

1. Signal received
2. Server stops accepting new connections
3. Existing connections are closed
4. Metrics server shuts down
5. Process exits

Shutdown timeout is 5 seconds for both the WebSocket and metrics servers.

