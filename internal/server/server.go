package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accept all connections
	},
}

// Server represents the WebSocket server
type Server struct {
	host        string
	port        int
	connections map[*websocket.Conn]string
	connMutex   sync.RWMutex
	metrics     *Metrics
	logger      *zap.Logger
}

// New creates a new WebSocket server
func New(host string, port int, metrics *Metrics, logger *zap.Logger) *Server {
	return &Server{
		host:        host,
		port:        port,
		connections: make(map[*websocket.Conn]string),
		metrics:     metrics,
		logger:      logger,
	}
}

// Start starts the WebSocket server
func (s *Server) Start(ctx context.Context) error {
	http.HandleFunc("/ws", s.handleWebSocket)

	// Start periodic connection logging
	go s.logConnectionsPeriodically(ctx)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.logger.Info("Starting WebSocket server", zap.String("address", addr))

	server := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down server", zap.Error(err))
		}
	}()

	return server.ListenAndServe()
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	remoteAddr := conn.RemoteAddr().String()
	s.logger.Info("WebSocket connected", zap.String("remote_addr", remoteAddr))

	s.connMutex.Lock()
	s.connections[conn] = remoteAddr
	s.updateMetrics()
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, conn)
		s.updateMetrics()
		s.connMutex.Unlock()
		conn.Close()
		s.logger.Info("WebSocket disconnected", zap.String("remote_addr", remoteAddr))
	}()

	// Keep connection alive and handle messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("Error reading message", zap.String("remote_addr", remoteAddr), zap.Error(err))
			}
			break
		}

		// Echo back the message
		if err := conn.WriteMessage(messageType, message); err != nil {
			s.logger.Error("Error writing message", zap.String("remote_addr", remoteAddr), zap.Error(err))
			break
		}
	}
}

func (s *Server) logConnectionsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.connMutex.RLock()
			count := len(s.connections)
			s.connMutex.RUnlock()
			s.logger.Info("Active connections", zap.Int("count", count))
		}
	}
}

func (s *Server) updateMetrics() {
	s.metrics.SetConnectionCount(len(s.connections))
}

// GetConnectionCount returns the current number of active connections
func (s *Server) GetConnectionCount() int {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()
	return len(s.connections)
}
