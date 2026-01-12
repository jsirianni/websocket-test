# Metrics and Monitoring

The WebSocket server exposes Prometheus metrics for monitoring connection health and performance.

## Metrics Endpoint

- **URL**: `http://<server>:9100/metrics`
- **Format**: Prometheus text-based exposition format
- **Protocol**: HTTP/1.1

## Available Metrics

### `websocket_connections`

**Type:** Gauge

**Description:** Current number of active WebSocket connections

**Labels:** None

**Example:**
```
# HELP websocket_connections Current number of active WebSocket connections
# TYPE websocket_connections gauge
websocket_connections 5
```

This metric is updated in real-time:
- Incremented when a client connects
- Decremented when a client disconnects

## Querying Metrics

### Using curl

```bash
# Get all metrics
curl http://localhost:9100/metrics

# Filter for WebSocket metrics only
curl http://localhost:9100/metrics | grep websocket

# Get just the value
curl -s http://localhost:9100/metrics | grep '^websocket_connections ' | awk '{print $2}'
```

### Using wget

```bash
wget -qO- http://localhost:9100/metrics | grep websocket
```

### Using httpie

```bash
http http://localhost:9100/metrics | grep websocket
```

## Prometheus Configuration

### Scrape Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'websocket-test'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9100']
        labels:
          service: 'websocket-test'
          environment: 'production'
```

### Service Discovery (Kubernetes)

For Kubernetes deployments, use pod annotations:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9100"
    prometheus.io/path: "/metrics"
```

Prometheus Operator PodMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: websocket-test-server
  labels:
    app: websocket-test-server
spec:
  selector:
    matchLabels:
      app: websocket-test-server
  podMetricsEndpoints:
  - port: metrics
    interval: 15s
```

Prometheus Operator ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: websocket-test-server
  labels:
    app: websocket-test-server
spec:
  selector:
    matchLabels:
      app: websocket-test-server
  endpoints:
  - port: metrics
    interval: 15s
```

## Alerting Rules

### Prometheus Alerting Rules

Example alerting rules for `prometheus.yml` or a separate rules file:

```yaml
groups:
  - name: websocket_test
    interval: 30s
    rules:
      # Alert when there are no connections for 5 minutes
      - alert: NoWebSocketConnections
        expr: websocket_connections == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "No WebSocket connections"
          description: "The WebSocket server has had zero connections for 5 minutes."

      # Alert when connections exceed threshold
      - alert: TooManyWebSocketConnections
        expr: websocket_connections > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High WebSocket connection count"
          description: "The WebSocket server has {{ $value }} connections (threshold: 1000)."

      # Alert when metrics endpoint is down
      - alert: WebSocketMetricsDown
        expr: up{job="websocket-test"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "WebSocket metrics endpoint is down"
          description: "The WebSocket server metrics endpoint has been unreachable for 2 minutes."
```

## Grafana Dashboard

### Example Dashboard JSON

Create a Grafana dashboard with these panels:

#### Panel 1: Active Connections (Time Series)

**PromQL Query:**
```promql
websocket_connections
```

**Panel Type:** Time series

**Legend:** `Active Connections`

#### Panel 2: Current Connection Count (Stat)

**PromQL Query:**
```promql
websocket_connections
```

**Panel Type:** Stat

**Options:**
- Show: Value
- Color mode: Background

#### Panel 3: Connection Rate (Graph)

**PromQL Query:**
```promql
rate(websocket_connections[5m])
```

**Panel Type:** Time series

**Legend:** `Connection Rate (5m)`

#### Panel 4: Connection Change (Bar Gauge)

**PromQL Query:**
```promql
delta(websocket_connections[1h])
```

**Panel Type:** Bar gauge

**Legend:** `Change (1h)`

### Import Dashboard

1. Open Grafana
2. Go to Dashboards → Import
3. Paste the JSON or upload the file
4. Select your Prometheus data source
5. Click Import

## Monitoring Best Practices

### Recommended Scrape Interval

- **Development**: 15-30 seconds
- **Production**: 30-60 seconds
- **High-load**: 10-15 seconds

### Retention

Configure appropriate retention in Prometheus:

```yaml
# prometheus.yml
storage:
  tsdb:
    retention.time: 15d  # Keep 15 days of data
    retention.size: 50GB # Or 50GB max
```

### HA Setup

For high availability:

1. **Multiple Prometheus instances** scraping the same targets
2. **Thanos** or **Cortex** for long-term storage
3. **Federation** to aggregate metrics from multiple Prometheus servers

### Alertmanager Integration

Configure Alertmanager to send notifications:

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'websocket-team'

receivers:
  - name: 'websocket-team'
    email_configs:
      - to: 'team@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/XXX'
        channel: '#websocket-alerts'
```

## Metrics Collection in Kubernetes

### Using Prometheus Operator

The Prometheus Operator automatically discovers services with the correct labels:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: websocket-test-server
  labels:
    app: websocket-test-server
spec:
  selector:
    app: websocket-test-server
  ports:
  - name: metrics
    port: 9100
    targetPort: 9100
```

Then create a ServiceMonitor:

```bash
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: websocket-test-server
  labels:
    release: prometheus  # Must match your Prometheus Operator release
spec:
  selector:
    matchLabels:
      app: websocket-test-server
  endpoints:
  - port: metrics
    interval: 30s
EOF
```

### Verify Prometheus is Scraping

1. Open Prometheus UI
2. Go to Status → Targets
3. Look for the `websocket-test` job
4. Verify status is "UP"

Or query directly:

```promql
up{job="websocket-test"}
```

## Troubleshooting

### Metrics Endpoint Not Accessible

**Check if metrics server is running:**

```bash
netstat -an | grep 9100
```

**Check firewall:**

```bash
# Allow incoming connections on port 9100
sudo ufw allow 9100/tcp
```

### Metrics Not Updating

**Verify metrics are changing:**

```bash
# Get current value
curl -s http://localhost:9100/metrics | grep websocket_connections

# Connect a client, then check again
./bin/client &
sleep 2
curl -s http://localhost:9100/metrics | grep websocket_connections
```

### Prometheus Not Scraping

**Check Prometheus logs:**

```bash
# Docker
docker logs prometheus

# Kubernetes
kubectl logs -n monitoring prometheus-server-0
```

**Common issues:**
- Incorrect target configuration
- Network connectivity
- DNS resolution issues
- Authentication/TLS errors

### High Cardinality

The current metrics implementation has **no labels**, so cardinality is not an issue. If you add labels in the future:

- Keep labels low-cardinality (< 10 values)
- Avoid user IDs or IP addresses as labels
- Use labels for grouping, not identification

## Future Metrics

Potential metrics to add:

```
# Message statistics
websocket_messages_received_total (counter)
websocket_messages_sent_total (counter)
websocket_message_errors_total (counter)

# Connection duration
websocket_connection_duration_seconds (histogram)

# Data transfer
websocket_bytes_received_total (counter)
websocket_bytes_sent_total (counter)

# Performance
websocket_message_processing_duration_seconds (histogram)
websocket_concurrent_handlers (gauge)
```

These can be added by extending the `Metrics` struct in `server/metrics.go`.

