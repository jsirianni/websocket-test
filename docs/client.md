# WebSocket Client

The WebSocket client connects to a WebSocket server and maintains a persistent connection with automatic reconnection.

## Features

- Automatic connection and reconnection
- Periodic connection status logging (every 30 seconds)
- Ping/pong keep-alive mechanism (every 15 seconds)
- Configurable server host and port
- Graceful shutdown handling
- Suitable for unattended operation

## Configuration

All configuration is done via command-line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `localhost` | WebSocket server host |
| `--port` | `3003` | WebSocket server port |
| `--connections` | `1` | Number of concurrent connections to establish |
| `--tls` | `false` | Use TLS (wss scheme) for WebSocket connection |

## Usage

### Running Locally

```bash
# Connect to default server (localhost:3003) with single connection
./client

# Connect to custom server
./client --host=example.com --port=8080

# Connect with TLS (wss scheme)
./client --host=example.com --port=443 --tls

# Connect with multiple concurrent connections (for load testing)
./client --connections=10
```

### Running in Docker

```bash
docker run ghcr.io/jsirianni/websocket-test-client:latest --host=server --port=3003

# With multiple connections for load testing
docker run ghcr.io/jsirianni/websocket-test-client:latest --host=server --port=3003 --connections=10

# Connect with TLS
docker run ghcr.io/jsirianni/websocket-test-client:latest --host=example.com --port=443 --tls
```

### Running in Kubernetes

The client is typically deployed in the same cluster as the server:

```bash
kubectl apply -f k8s/client-deployment.yaml
```

The Kubernetes deployment is pre-configured to connect to the server service.

## Connection Behavior

### Initial Connection

1. Client attempts to connect to the specified server
2. On success, logs "Connected to ws://host:port/ws"
3. Begins sending periodic pings and logging status

### Connection Loss

1. Client detects connection loss
2. Logs "Connection lost: <error>"
3. Waits 5 seconds (reconnect delay)
4. Attempts to reconnect
5. Repeats until successful or interrupted

### Keep-Alive

The client sends WebSocket ping frames every 15 seconds to:
- Keep the connection alive through firewalls/proxies
- Detect connection failures quickly
- Maintain NAT session state

## Logging

The client logs the following events:

- Connection attempts
- Successful connections
- Periodic connection status (every 30 seconds)
- Connection failures and reconnection attempts
- Shutdown signals

Example log output (single connection):

```
2026/01/12 10:00:00 [Connection 1] Connecting to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 1] Connected to ws://localhost:3003/ws
2026/01/12 10:00:30 [Connection 1] Connection status: CONNECTED to ws://localhost:3003/ws
2026/01/12 10:01:00 [Connection 1] Connection status: CONNECTED to ws://localhost:3003/ws
2026/01/12 10:01:15 [Connection 1] Connection lost: websocket: close 1006 (abnormal closure)
2026/01/12 10:01:15 [Connection 1] Connection failed: EOF. Retrying in 5s...
2026/01/12 10:01:20 [Connection 1] Connecting to ws://localhost:3003/ws
2026/01/12 10:01:20 [Connection 1] Connected to ws://localhost:3003/ws
```

Example log output (multiple connections):

```
2026/01/12 10:00:00 [Connection 1] Connecting to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 2] Connecting to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 3] Connecting to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 1] Connected to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 2] Connected to ws://localhost:3003/ws
2026/01/12 10:00:00 [Connection 3] Connected to ws://localhost:3003/ws
2026/01/12 10:00:30 [Connection 1] Connection status: CONNECTED to ws://localhost:3003/ws
2026/01/12 10:00:30 [Connection 2] Connection status: CONNECTED to ws://localhost:3003/ws
2026/01/12 10:00:30 [Connection 3] Connection status: CONNECTED to ws://localhost:3003/ws
```

## Graceful Shutdown

The client handles `SIGINT` and `SIGTERM` signals for graceful shutdown:

1. Signal received
2. Connection context is cancelled
3. WebSocket connection is closed
4. Process exits

## Unattended Operation

The client is designed to run unattended:

- Automatic reconnection on connection loss
- No user interaction required
- Continuous operation until explicitly stopped
- Detailed logging for monitoring

This makes it ideal for:
- Long-running connection tests
- Network reliability monitoring
- Load testing scenarios (use `--connections` flag to test with multiple connections from a single instance)
- Kubernetes deployments

