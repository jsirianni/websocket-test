# Quick Start Guide

This guide will help you get the WebSocket test tools running quickly.

## Prerequisites

Choose one of the following:
- Go 1.21+ (for building from source)
- Docker (for running containers)
- kubectl and a Kubernetes cluster (for Kubernetes deployment)

## Option 1: Build and Run from Source

### Step 1: Clone and Build

```bash
# Clone the repository
git clone https://github.com/jsirianni/websocket-test.git
cd websocket-test

# Build both binaries
make build

# Or build individually
make server  # builds bin/server
make client  # builds bin/client
```

### Step 2: Run the Server

Open a terminal and start the server:

```bash
./bin/server
```

You should see:
```
2026/01/12 10:00:00 Starting WebSocket server on 0.0.0.0:3003
2026/01/12 10:00:00 Starting metrics server on :9100
```

### Step 3: Run the Client

Open another terminal and start the client:

```bash
./bin/client
```

You should see:
```
2026/01/12 10:00:05 Connecting to ws://localhost:3003/ws
2026/01/12 10:00:05 Connected to ws://localhost:3003/ws
```

The server will log:
```
2026/01/12 10:00:05 WebSocket connected: 127.0.0.1:54321
```

### Step 4: Check Metrics

In a third terminal:

```bash
curl http://localhost:9100/metrics | grep websocket
```

You should see:
```
# HELP websocket_connections Current number of active WebSocket connections
# TYPE websocket_connections gauge
websocket_connections 1
```

### Step 5: Watch Periodic Logs

Both server and client will log status every 30 seconds:

**Server:**
```
2026/01/12 10:00:30 Active connections: 1
2026/01/12 10:01:00 Active connections: 1
```

**Client:**
```
2026/01/12 10:00:35 Connection status: CONNECTED to ws://localhost:3003/ws
2026/01/12 10:01:05 Connection status: CONNECTED to ws://localhost:3003/ws
```

### Step 6: Test Reconnection

Stop the server (Ctrl+C) and watch the client automatically reconnect:

**Client logs:**
```
2026/01/12 10:01:15 Connection lost: EOF
2026/01/12 10:01:15 Connection failed: EOF. Retrying in 5s...
2026/01/12 10:01:20 Connecting to ws://localhost:3003/ws
2026/01/12 10:01:20 Connection failed: dial tcp 127.0.0.1:3003: connection refused. Retrying in 5s...
```

Restart the server and the client will reconnect:
```
2026/01/12 10:01:45 Connecting to ws://localhost:3003/ws
2026/01/12 10:01:45 Connected to ws://localhost:3003/ws
```

## Option 2: Run with Docker

### Step 1: Pull Images

```bash
docker pull ghcr.io/jsirianni/websocket-test-server:latest
docker pull ghcr.io/jsirianni/websocket-test-client:latest
```

### Step 2: Create a Network

```bash
docker network create websocket-test
```

### Step 3: Run the Server

```bash
docker run -d \
  --name websocket-server \
  --network websocket-test \
  -p 3003:3003 \
  -p 9100:9100 \
  ghcr.io/jsirianni/websocket-test-server:latest
```

### Step 4: Run the Client

```bash
docker run -d \
  --name websocket-client \
  --network websocket-test \
  ghcr.io/jsirianni/websocket-test-client:latest \
  --host=websocket-server
```

### Step 5: Check Logs

```bash
# Server logs
docker logs -f websocket-server

# Client logs
docker logs -f websocket-client
```

### Step 6: Check Metrics

```bash
curl http://localhost:9100/metrics | grep websocket
```

### Cleanup

```bash
docker stop websocket-server websocket-client
docker rm websocket-server websocket-client
docker network rm websocket-test
```

## Option 3: Deploy to Kubernetes

### Step 1: Apply Manifests

```bash
kubectl apply -f k8s/
```

This creates:
- Server deployment (1 replica)
- Server service (ClusterIP)
- Client deployment (1 replica)

### Step 2: Check Status

```bash
# Check pods
kubectl get pods -l app=websocket-test-server
kubectl get pods -l app=websocket-test-client

# Check service
kubectl get svc websocket-test-server
```

### Step 3: View Logs

```bash
# Server logs
kubectl logs -f -l app=websocket-test-server

# Client logs
kubectl logs -f -l app=websocket-test-client
```

### Step 4: Access Metrics

Forward the metrics port to your local machine:

```bash
kubectl port-forward svc/websocket-test-server 9100:9100
```

Then in another terminal:

```bash
curl http://localhost:9100/metrics | grep websocket
```

### Step 5: Scale the Client

Test with multiple clients:

```bash
kubectl scale deployment websocket-test-client --replicas=5
```

Watch the connection count increase:

```bash
curl http://localhost:9100/metrics | grep websocket_connections
```

### Cleanup

```bash
kubectl delete -f k8s/
```

## Configuration Examples

### Custom Server Port

```bash
./bin/server --host=127.0.0.1 --port=8080 --metrics-port=9090
```

### Connect to Remote Server

```bash
./bin/client --host=example.com --port=8080
```

### Load Testing with Multiple Connections

```bash
# Connect with 10 concurrent connections from a single client instance
./bin/client --connections=10
```

### Docker with Custom Configuration

```bash
docker run -d \
  --name websocket-server \
  -p 8080:8080 \
  -p 9090:9090 \
  ghcr.io/jsirianni/websocket-test-server:latest \
  --host=0.0.0.0 \
  --port=8080 \
  --metrics-port=9090
```

## Monitoring with Prometheus

If you have Prometheus running, add this scrape config:

```yaml
scrape_configs:
  - job_name: 'websocket-test'
    static_configs:
      - targets: ['localhost:9100']
```

Or in Kubernetes, add these annotations to the server deployment:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "9100"
  prometheus.io/path: "/metrics"
```

## Troubleshooting

### Server won't start

**Error:** `bind: address already in use`

**Solution:** Another process is using the port. Either stop that process or use a different port:

```bash
./bin/server --port=8080 --metrics-port=9090
```

### Client can't connect

**Error:** `Connection failed: dial tcp: connection refused`

**Solution:** 
1. Verify the server is running
2. Check firewall settings
3. Verify the correct host and port:

```bash
./bin/client --host=127.0.0.1 --port=3003
```

### Docker client can't connect to host

**Solution:** Use the special Docker host:

```bash
# On Docker for Mac/Windows
docker run ghcr.io/jsirianni/websocket-test-client:latest \
  --host=host.docker.internal --port=3003

# On Linux, use host networking
docker run --network host ghcr.io/jsirianni/websocket-test-client:latest
```

### Kubernetes pods not starting

Check the events:

```bash
kubectl describe pod -l app=websocket-test-server
kubectl describe pod -l app=websocket-test-client
```

Common issues:
- Image pull errors: Check registry authentication
- CrashLoopBackOff: Check logs with `kubectl logs`

## Next Steps

- Read the [Server Documentation](server.md) for detailed configuration
- Read the [Client Documentation](client.md) for advanced usage
- See the [Development Guide](development.md) for building and contributing
- Check out the [Architecture](architecture.md) to understand how it works

