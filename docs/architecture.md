# Architecture

This document describes the architecture and design decisions of the WebSocket test project.

## Overview

The project consists of two main components:
1. **Server**: A WebSocket server that accepts connections and provides metrics
2. **Client**: A WebSocket client that maintains persistent connections

## Component Architecture

### Server Architecture

```
┌─────────────────────────────────────────┐
│         WebSocket Server                │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   HTTP Server (Port 3003)         │ │
│  │   ├─ /ws (WebSocket Upgrade)      │ │
│  │   └─ Connection Handler           │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Metrics Server (Port 9100)      │ │
│  │   └─ /metrics (Prometheus)        │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Connection Manager              │ │
│  │   ├─ Active Connections Map       │ │
│  │   ├─ Mutex for Thread Safety      │ │
│  │   └─ Periodic Logger              │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Prometheus Metrics              │ │
│  │   └─ Connection Count Gauge       │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### Key Design Decisions

1. **Separate HTTP Servers**: The WebSocket server and metrics server run on separate HTTP servers to isolate concerns and allow independent port configuration.

2. **Thread-Safe Connection Tracking**: A mutex protects the connections map to ensure thread-safe access from multiple goroutines.

3. **Echo Protocol**: The server echoes back any messages it receives, making it easy to test bidirectional communication.

4. **Graceful Shutdown**: Context-based shutdown allows for clean termination of both servers and all active connections.

### Client Architecture

```
┌─────────────────────────────────────────┐
│         WebSocket Client                │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Connection Manager              │ │
│  │   ├─ Connect                      │ │
│  │   ├─ Reconnect on Failure         │ │
│  │   └─ Graceful Disconnect          │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Keep-Alive Manager              │ │
│  │   ├─ Ping Timer (15s)             │ │
│  │   └─ Pong Handler                 │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Status Logger                   │ │
│  │   └─ Periodic Status (30s)        │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### Key Design Decisions

1. **Automatic Reconnection**: The client continuously attempts to reconnect on connection loss, with a 5-second delay between attempts.

2. **Keep-Alive Mechanism**: Periodic ping frames (every 15 seconds) detect connection failures quickly and keep the connection alive through NAT/firewalls.

3. **Unattended Operation**: The client is designed to run indefinitely without user interaction, making it suitable for long-term testing and monitoring.

4. **Context-Based Cancellation**: Shutdown signals are propagated through contexts to cleanly stop all goroutines.

## Communication Flow

### Connection Establishment

```
Client                          Server
  │                               │
  ├─── WebSocket Upgrade ────────>│
  │                               ├─ Log: "Connected: IP:Port"
  │                               ├─ Add to connections map
  │<──── Upgrade Success ─────────┤
  │                               │
  ├─ Log: "Connected to..."       │
  │                               │
```

### Keep-Alive (Ping/Pong)

```
Client                          Server
  │                               │
  ├────── Ping Frame ────────────>│
  │<───── Pong Frame ─────────────┤
  │                               │
  (Every 15 seconds)
```

### Periodic Logging

```
Server (every 30 seconds):
  - Log: "Active connections: N"
  - Update Prometheus metric

Client (every 30 seconds):
  - Log: "Connection status: CONNECTED to..."
```

### Connection Loss and Reconnect

```
Client                          Server
  │                               │
  │         (Connection Lost)     │
  │<────────── x ──────────────────
  │                               ├─ Log: "Disconnected: IP:Port"
  ├─ Log: "Connection lost"       ├─ Remove from connections map
  │                               │
  (Wait 5 seconds)
  │                               │
  ├─── Reconnect Attempt ────────>│
  │<──── New Connection ──────────┤
  │                               │
```

## Concurrency Model

### Server Concurrency

1. **Main Goroutine**: Runs the WebSocket HTTP server
2. **Metrics Goroutine**: Runs the Prometheus metrics HTTP server
3. **Logger Goroutine**: Periodically logs connection count
4. **Per-Connection Goroutines**: Each WebSocket connection runs in its own goroutine

Thread safety is ensured through:
- Mutex protection for the connections map
- Read/write locks for efficient concurrent access

### Client Concurrency

1. **Main Goroutine**: Connection management and reconnection loop
2. **Ping Goroutine**: Sends periodic ping frames
3. **Status Logger**: Logs connection status periodically

All goroutines respect context cancellation for clean shutdown.

## Metrics

### Prometheus Metrics

The server exposes a single gauge metric:

```
# HELP websocket_connections Current number of active WebSocket connections
# TYPE websocket_connections gauge
websocket_connections 5
```

This metric is updated in real-time as connections are established and closed.

### Metric Collection

External systems (like Prometheus) can scrape the `/metrics` endpoint:

```
┌───────────┐          ┌──────────────┐
│ Prometheus│─ scrape ─>│ Server:9100  │
│  Server   │<─metrics──│   /metrics   │
└───────────┘          └──────────────┘
```

## Container Images

### From Scratch Images

Both server and client use `FROM scratch` base images for:
- Minimal attack surface
- Small image size (< 20 MB)
- Fast startup time
- No unnecessary dependencies

### Multi-Architecture Support

Images are built for:
- `linux/amd64`
- `linux/arm64`

Multi-arch manifests allow `docker pull` to automatically select the correct architecture.

## Kubernetes Deployment

### Pod Architecture

```
┌─────────────────────────────────────────┐
│         Kubernetes Cluster              │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Server Deployment              │   │
│  │  ├─ Pod (Replica 1)             │   │
│  │  │  └─ Container: server        │   │
│  │  │     ├─ Port 3003 (WebSocket) │   │
│  │  │     └─ Port 9100 (Metrics)   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Server Service                 │   │
│  │  ├─ ClusterIP                   │   │
│  │  ├─ Port 3003 → Pod:3003        │   │
│  │  └─ Port 9100 → Pod:9100        │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Client Deployment              │   │
│  │  ├─ Pod (Replica 1)             │   │
│  │  │  └─ Container: client        │   │
│  │  │     └─ Connects to Service   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Service Discovery

The client uses the Kubernetes service name (`websocket-test-server`) to connect to the server, leveraging Kubernetes DNS for service discovery.

## Error Handling

### Server Error Handling

1. **Connection Upgrade Failures**: Logged and connection rejected
2. **Read/Write Errors**: Connection closed and removed from tracking
3. **Shutdown Errors**: Logged but don't prevent shutdown

### Client Error Handling

1. **Connection Failures**: Logged and automatic reconnection attempted
2. **Read/Write Errors**: Connection closed and reconnection triggered
3. **Context Cancellation**: Clean shutdown without error logging

## Security Considerations

### Current Implementation

- **No Authentication**: The server accepts all connections without authentication
- **No TLS**: Connections use plain WebSocket (ws://)
- **Accept All Origins**: CORS is disabled (CheckOrigin returns true)

### Production Recommendations

For production use, consider adding:
1. **TLS/WSS**: Use secure WebSocket connections
2. **Authentication**: Token-based or certificate-based authentication
3. **Origin Checking**: Validate WebSocket origin headers
4. **Rate Limiting**: Prevent connection flooding
5. **Network Policies**: Kubernetes network policies to restrict access

## Performance Characteristics

### Server Performance

- **Connection Overhead**: ~1 KB memory per connection
- **CPU Usage**: Minimal (mostly idle waiting for messages)
- **Network**: Low bandwidth (only ping/pong and echo messages)

### Client Performance

- **Reconnection Overhead**: 5-second delay between attempts
- **Keep-Alive Traffic**: Ping frame every 15 seconds (~20 bytes)
- **Logging Overhead**: Minimal (one log entry every 30 seconds)

### Scalability

- Server can handle thousands of concurrent connections
- Bottlenecks:
  - File descriptors (OS limit)
  - Memory (connection tracking map)
  - CPU (minimal with current echo implementation)

## Build Process

### GoReleaser Build Pipeline

```
┌─────────────────────────────────────────┐
│         GoReleaser Pipeline             │
│                                         │
│  1. Build Binaries                      │
│     ├─ linux/amd64                      │
│     ├─ linux/arm64                      │
│     ├─ darwin/amd64                     │
│     ├─ darwin/arm64                     │
│     ├─ windows/amd64                    │
│     └─ windows/arm64                    │
│                                         │
│  2. Create Archives                     │
│     └─ tar.gz / zip                     │
│                                         │
│  3. Build Docker Images                 │
│     ├─ server (amd64, arm64)            │
│     └─ client (amd64, arm64)            │
│                                         │
│  4. Create Manifests                    │
│     ├─ server:version                   │
│     ├─ server:latest                    │
│     ├─ client:version                   │
│     └─ client:latest                    │
│                                         │
│  5. Push to Registry                    │
│     └─ ghcr.io/jsirianni/*               │
│                                         │
│  6. Create GitHub Release               │
│     ├─ Binaries                         │
│     ├─ Archives                         │
│     └─ Checksums                        │
└─────────────────────────────────────────┘
```

Timezone information is embedded in the binaries via the `import _ "time/tzdata"` statement in the main packages, which is essential for proper time handling in "from scratch" Docker images.

