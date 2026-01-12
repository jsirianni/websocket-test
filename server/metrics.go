package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics holds Prometheus metrics
type Metrics struct {
	connectionCount prometheus.Gauge
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	m := &Metrics{
		connectionCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Current number of active WebSocket connections",
		}),
	}

	prometheus.MustRegister(m.connectionCount)
	return m
}

// SetConnectionCount updates the connection count metric
func (m *Metrics) SetConnectionCount(count int) {
	m.connectionCount.Set(float64(count))
}

// StartMetricsServer starts the Prometheus metrics HTTP server
func StartMetricsServer(ctx context.Context, port int, logger *zap.Logger) error {
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", port)
	logger.Info("Starting metrics server", zap.String("address", addr))

	server := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down metrics server", zap.Error(err))
		}
	}()

	return server.ListenAndServe()
}

