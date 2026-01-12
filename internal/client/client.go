package client

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	host        string
	port        int
	useTLS      bool
	logInterval time.Duration

	handshakeTimeout time.Duration
	pongWait         time.Duration
	pingPeriod       time.Duration
	writeWait        time.Duration
	logger           *zap.Logger
}

func New(host string, port int, useTLS bool, logger *zap.Logger) *Client {
	pongWait := 60 * time.Second
	return &Client{
		host:             host,
		port:             port,
		useTLS:           useTLS,
		logInterval:      30 * time.Second,
		handshakeTimeout: 10 * time.Second,
		pongWait:         pongWait,
		pingPeriod:       (pongWait * 9) / 10,
		writeWait:        10 * time.Second,
		logger:           logger,
	}
}

// Start opens exactly one WebSocket connection and blocks until:
// - ctx is cancelled, or
// - the connection fails/closes.
// No retry logic.
func (c *Client) Start(ctx context.Context) error {
	scheme := "ws"
	if c.useTLS {
		scheme = "wss"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", c.host, c.port),
		Path:   "/ws",
	}
	urlStr := u.String()

	dialer := websocket.Dialer{
		HandshakeTimeout: c.handshakeTimeout,
	}

	conn, resp, err := dialer.DialContext(ctx, urlStr, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial failed (http %s): %w", resp.Status, err)
		}
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	c.logger.Info("Connected", zap.String("url", urlStr))

	// Gorilla allows one concurrent reader + one concurrent writer.
	// We serialize ALL writes (pings + pongs + close) with this mutex.
	var writeMu sync.Mutex

	// Read deadline is driven by pong responses.
	conn.SetReadDeadline(time.Now().Add(c.pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	// If the server sends pings, respond with pongs (serialized with writeMu).
	conn.SetPingHandler(func(appData string) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(c.writeWait),
		)
	})

	done := make(chan struct{})
	defer close(done)

	// Close the connection when ctx is cancelled (best-effort close frame).
	go func() {
		select {
		case <-ctx.Done():
			writeMu.Lock()
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context cancelled"),
				time.Now().Add(c.writeWait),
			)
			writeMu.Unlock()
			_ = conn.Close()
		case <-done:
		}
	}()

	// Keepalive pings.
	go func() {
		t := time.NewTicker(c.pingPeriod)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				writeMu.Lock()
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(c.writeWait))
				writeMu.Unlock()
				if err != nil {
					c.logger.Error("Ping failed", zap.Error(err))
					_ = conn.Close()
					return
				}
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	// Periodic status logging (optional).
	go func() {
		t := time.NewTicker(c.logInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				c.logger.Info("Connection status", zap.String("status", "CONNECTED"), zap.String("url", urlStr))
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	// Read loop (blocks). If the server sends no application messages, that's fine:
	// pongs still count as reads and will extend the read deadline.
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return fmt.Errorf("websocket closed unexpectedly: %w", err)
			}
			return fmt.Errorf("websocket read failed: %w", err)
		}
	}
}
