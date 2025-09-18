package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"aktuell/pkg/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *models.ServerMessage
	register   chan *Client
	unregister chan *Client
	logger     *logrus.Logger
	wsServer   *WebSocketServer
	mu         sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	hub           *Hub
	conn          *websocket.Conn
	send          chan *models.ServerMessage
	subscriptions map[string]*models.Subscription
	closed        bool // Track if connection has been closed
	mu            sync.RWMutex
}

// WebSocketServer wraps the HTTP server and WebSocket hub
type WebSocketServer struct {
	hub              *Hub
	server           *http.Server
	logger           *logrus.Logger
	validator        models.SubscriptionValidator
	snapshotStreamer models.SnapshotStreamer
	actualAddr       string     // Store the actual listening address
	addrMu           sync.Mutex // Protect actualAddr field
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096, // Increased from 1024
	WriteBufferSize: 4096, // Increased from 1024
	CheckOrigin:     checkOrigin,
}

// getDefaultAllowedOrigins returns default origins based on environment
func getDefaultAllowedOrigins() []string {
	env := os.Getenv("AKTUELL_ENV")
	if env == "production" {
		// In production, only allow explicitly configured origins
		return []string{}
	}

	// Development defaults
	return []string{
		"http://localhost:3000",  // React development server
		"http://localhost:8080",  // Alternative dev server
		"https://localhost:3000", // HTTPS dev server
	}
}

// checkOrigin validates WebSocket connection origins for security
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// In development mode (when AKTUELL_ENV != "production"), allow localhost origins
	env := os.Getenv("AKTUELL_ENV")
	if env != "production" {
		if origin == "" {
			return true // Allow same-origin requests
		}

		// Allow any localhost or 127.0.0.1 origin in development
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			return true
		}
	}

	// Priority 1: Check custom allowed origins from environment variable
	if customOrigins := os.Getenv("AKTUELL_ALLOWED_ORIGINS"); customOrigins != "" {
		customList := strings.Split(customOrigins, ",")
		for _, allowed := range customList {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}

		// In production, if AKTUELL_ALLOWED_ORIGINS is set, ONLY use those origins
		if env == "production" {
			log.Printf("WebSocket connection rejected from origin: %s (not in AKTUELL_ALLOWED_ORIGINS)", origin)
			return false
		}
	}

	// Priority 2: Check default allowed origins (only used in development)
	defaultOrigins := getDefaultAllowedOrigins()
	for _, allowed := range defaultOrigins {
		if origin == allowed {
			return true
		}
	}

	// Log rejected origins for security monitoring
	log.Printf("WebSocket connection rejected from origin: %s (remote: %s)", origin, r.RemoteAddr)
	return false
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(addr string, logger *logrus.Logger) *WebSocketServer {
	ws := &WebSocketServer{
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}

	hub := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *models.ServerMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		wsServer:   ws, // Set the reference back to the WebSocket server
	}

	ws.hub = hub

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.handleWebSocket)
	mux.HandleFunc("/health", handleHealth)

	ws.server.Handler = mux

	return ws
}

// SetValidator sets the subscription validator for the WebSocket server
func (ws *WebSocketServer) SetValidator(validator models.SubscriptionValidator) {
	ws.validator = validator
}

// SetSnapshotStreamer sets the snapshot streamer for the WebSocket server
func (ws *WebSocketServer) SetSnapshotStreamer(streamer models.SnapshotStreamer) {
	ws.snapshotStreamer = streamer
}

// Start starts the WebSocket server and hub
func (ws *WebSocketServer) Start() error {
	// Start the hub in a goroutine
	go ws.hub.run()

	ws.logger.WithField("addr", ws.server.Addr).Info("Starting WebSocket server")

	// Create a listener to get the actual address
	listener, err := net.Listen("tcp", ws.server.Addr)
	if err != nil {
		return err
	}

	// Store the actual listening address with proper synchronization
	ws.addrMu.Lock()
	ws.actualAddr = listener.Addr().String()
	ws.addrMu.Unlock()

	return ws.server.Serve(listener)
}

// Stop stops the WebSocket server
func (ws *WebSocketServer) Stop() error {
	return ws.server.Close()
}

// GetAddr returns the server's actual listening address
func (ws *WebSocketServer) GetAddr() string {
	ws.addrMu.Lock()
	addr := ws.actualAddr
	ws.addrMu.Unlock()

	if addr != "" {
		return addr
	}
	if ws.server.Addr == "" {
		return "localhost:8080" // default fallback
	}
	return ws.server.Addr
}

// BroadcastChange broadcasts a change event to all subscribed clients
func (ws *WebSocketServer) BroadcastChange(change *models.ChangeEvent) {
	message := &models.ServerMessage{
		Type:   models.MessageTypeChange,
		Change: change,
	}
	ws.hub.broadcast <- message
}

// GetHub returns the WebSocket hub
func (ws *WebSocketServer) GetHub() *Hub {
	return ws.hub
}

// run starts the hub's main loop
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			h.logger.WithFields(logrus.Fields{
				"client_id":     client.ID,
				"total_clients": len(h.clients),
			}).Info("Client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

			h.logger.WithFields(logrus.Fields{
				"client_id":     client.ID,
				"total_clients": len(h.clients),
			}).Info("Client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this change
				if h.isClientSubscribed(client, message) {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// isClientSubscribed checks if a client is subscribed to the change event
func (h *Hub) isClientSubscribed(client *Client, message *models.ServerMessage) bool {
	if message.Change == nil {
		return true // Non-change messages go to all clients
	}

	client.mu.RLock()
	defer client.mu.RUnlock()

	// If no subscriptions, client should not receive change events
	if len(client.subscriptions) == 0 {
		return false
	}

	// Check if any subscription matches the change
	for _, sub := range client.subscriptions {
		if sub.Database == message.Change.Database &&
			(sub.Collection == "" || sub.Collection == message.Change.Collection) {
			return true
		}
	}

	return false
}

// handleWebSocket handles WebSocket upgrade and client management
func (h *Hub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	client := &Client{
		ID:            uuid.New().String(),
		hub:           h,
		conn:          conn,
		send:          make(chan *models.ServerMessage, 1024), // Increased from 256
		subscriptions: make(map[string]*models.Subscription),
	}

	client.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// safeClose safely closes the WebSocket connection, preventing multiple closes
func (c *Client) safeClose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		c.conn.Close()
	}
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.safeClose()
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		c.hub.logger.WithError(err).Error("Failed to set read deadline")
		return
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			c.hub.logger.WithError(err).Error("Failed to set read deadline in pong handler")
		}
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		var clientMessage models.ClientMessage
		if err := json.Unmarshal(messageBytes, &clientMessage); err != nil {
			c.hub.logger.WithError(err).Error("Failed to unmarshal client message")
			continue
		}

		c.handleMessage(&clientMessage)
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.safeClose()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Check if connection is still open before sending close message
				c.mu.RLock()
				isClosed := c.closed
				c.mu.RUnlock()

				if !isClosed {
					if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
						// Only log as debug since connection may have been closed by peer
						c.hub.logger.WithError(err).Debug("Failed to send close message - connection may already be closed")
					}
				}
				return
			}

			// Set longer write deadline for snapshot messages which can be large
			writeDeadline := 10 * time.Second
			if message.Type == models.MessageTypeSnapshot {
				writeDeadline = 30 * time.Second // Longer timeout for snapshot data
			}
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
				c.hub.logger.WithError(err).Error("Failed to set write deadline")
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.hub.logger.WithError(err).WithFields(logrus.Fields{
					"client_id":    c.ID,
					"message_type": message.Type,
				}).Error("Failed to write message to client")
				return
			}

			// Log successful message sends for debugging
			c.hub.logger.WithFields(logrus.Fields{
				"client_id":    c.ID,
				"message_type": message.Type,
			}).Debug("Successfully sent message to client")

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.hub.logger.WithError(err).Error("Failed to set write deadline for ping")
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming client messages
func (c *Client) handleMessage(message *models.ClientMessage) {
	switch message.Type {
	case models.MessageTypeSubscribe:
		c.handleSubscribe(message)
	case models.MessageTypeUnsubscribe:
		c.handleUnsubscribe(message)
	case models.MessageTypePing:
		c.handlePing(message)
	case models.MessageTypeHealth:
		c.handleHealthWS(message)
	default:
		c.hub.logger.WithField("type", message.Type).Warn("Unknown message type")
	}
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(message *models.ClientMessage) {
	// Validate the subscription if a validator is available
	if c.hub.wsServer != nil && c.hub.wsServer.validator != nil {
		if !c.hub.wsServer.validator.IsValidSubscription(message.Database, message.Collection) {
			// Send error response for invalid subscription
			response := &models.ServerMessage{
				Type:      models.MessageTypeError,
				Success:   false,
				Error:     fmt.Sprintf("Invalid subscription: database '%s' collection '%s' is not configured on the server", message.Database, message.Collection),
				RequestID: message.RequestID,
				ErrorCode: 1,
			}

			select {
			case c.send <- response:
			default:
				c.hub.logger.Warn("Failed to send subscription error response")
			}

			c.hub.logger.WithFields(logrus.Fields{
				"client_id":  c.ID,
				"database":   message.Database,
				"collection": message.Collection,
			}).Warn("Client attempted to subscribe to non-configured database/collection")
			return
		}
	}

	// Valid subscription - create it
	subscription := &models.Subscription{
		ID:              uuid.New().String(),
		ClientID:        c.ID,
		Database:        message.Database,
		Collection:      message.Collection,
		CreatedAt:       time.Now(),
		SnapshotOptions: message.SnapshotOptions,
	}

	// Debug: Log what we received
	c.hub.logger.WithFields(logrus.Fields{
		"client_id":         c.ID,
		"raw_snapshot_opts": message.SnapshotOptions,
		"snapshot_opts_nil": message.SnapshotOptions == nil,
	}).Debug("Received subscription with snapshot options")

	if message.SnapshotOptions != nil {
		c.hub.logger.WithFields(logrus.Fields{
			"client_id":        c.ID,
			"include_snapshot": message.SnapshotOptions.IncludeSnapshot,
			"snapshot_limit":   message.SnapshotOptions.SnapshotLimit,
			"batch_size":       message.SnapshotOptions.BatchSize,
		}).Debug("Snapshot options details")
	}

	c.mu.Lock()
	c.subscriptions[subscription.ID] = subscription
	c.mu.Unlock()

	// Send successful subscription response
	response := &models.ServerMessage{
		Type:      models.MessageTypeSubscribe,
		Success:   true,
		RequestID: message.RequestID,
		Data: map[string]interface{}{
			"subscription_id": subscription.ID,
		},
	}

	select {
	case c.send <- response:
	default:
		c.hub.logger.Warn("Failed to send subscription response")
	}

	c.hub.logger.WithFields(logrus.Fields{
		"client_id":    c.ID,
		"subscription": subscription.ID,
		"database":     subscription.Database,
		"collection":   subscription.Collection,
		"snapshot":     subscription.SnapshotOptions != nil && subscription.SnapshotOptions.IncludeSnapshot,
	}).Info("Client subscribed")

	// Handle snapshot if requested
	if subscription.SnapshotOptions != nil && subscription.SnapshotOptions.IncludeSnapshot {
		c.handleSnapshot(subscription)
	}
}

// handleSnapshot handles initial snapshot streaming for a subscription
func (c *Client) handleSnapshot(subscription *models.Subscription) {
	// Check if snapshot streamer is available
	if c.hub.wsServer == nil || c.hub.wsServer.snapshotStreamer == nil {
		c.hub.logger.Warn("Snapshot requested but no snapshot streamer configured")
		return
	}

	// Send snapshot start message
	startMsg := &models.ServerMessage{
		Type: models.MessageTypeSnapshotStart,
	}

	select {
	case c.send <- startMsg:
	default:
		c.hub.logger.Warn("Failed to send snapshot start message")
		return
	}

	// Set up callback for receiving snapshot batches
	callback := func(batch []map[string]interface{}, batchNum int, remaining int, err error) {
		if err != nil {
			// Send error message
			errorMsg := &models.ServerMessage{
				Type:  models.MessageTypeError,
				Error: fmt.Sprintf("Snapshot error: %v", err),
			}

			select {
			case c.send <- errorMsg:
			default:
				c.hub.logger.WithError(err).Warn("Failed to send snapshot error")
			}
			return
		}

		if len(batch) > 0 {
			// Send snapshot batch
			msg := &models.ServerMessage{
				Type:              models.MessageTypeSnapshot,
				SnapshotData:      batch,
				SnapshotBatch:     batchNum,
				SnapshotRemaining: remaining,
			}

			select {
			case c.send <- msg:
			default:
				c.hub.logger.Warn("Failed to send snapshot batch")
				return
			}

			c.hub.logger.WithFields(logrus.Fields{
				"client_id":  c.ID,
				"batch":      batchNum,
				"batch_size": len(batch),
				"remaining":  remaining,
				"database":   subscription.Database,
				"collection": subscription.Collection,
			}).Debug("Sent snapshot batch")
		}

		// Send snapshot end message when complete
		if remaining == 0 {
			endMsg := &models.ServerMessage{
				Type: models.MessageTypeSnapshotEnd,
			}

			select {
			case c.send <- endMsg:
			default:
				c.hub.logger.Warn("Failed to send snapshot end message")
			}

			c.hub.logger.WithFields(logrus.Fields{
				"client_id":  c.ID,
				"database":   subscription.Database,
				"collection": subscription.Collection,
			}).Info("Snapshot streaming completed")
		}
	}

	// Start snapshot streaming in a goroutine to avoid blocking
	go func() {
		c.hub.logger.WithFields(logrus.Fields{
			"client_id":  c.ID,
			"database":   subscription.Database,
			"collection": subscription.Collection,
		}).Info("Starting snapshot streaming")

		c.hub.wsServer.snapshotStreamer.StreamSnapshot(
			subscription.Database,
			subscription.Collection,
			subscription.SnapshotOptions,
			callback,
		)
	}()
}

// handleUnsubscribe handles unsubscription requests
func (c *Client) handleUnsubscribe(message *models.ClientMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var success bool
	var errorMsg string

	if message.SubscriptionID != "" {
		// Remove specific subscription
		if _, exists := c.subscriptions[message.SubscriptionID]; exists {
			delete(c.subscriptions, message.SubscriptionID)
			success = true
			c.hub.logger.WithFields(logrus.Fields{
				"client_id":       c.ID,
				"subscription_id": message.SubscriptionID,
			}).Info("Client unsubscribed from specific subscription")
		} else {
			success = false
			errorMsg = "Subscription not found"
		}
	} else {
		// Remove all subscriptions if no specific ID provided
		c.subscriptions = make(map[string]*models.Subscription)
		success = true
		c.hub.logger.WithField("client_id", c.ID).Info("Client unsubscribed from all subscriptions")
	}

	response := &models.ServerMessage{
		Type:      models.MessageTypeUnsubscribe,
		Success:   success,
		RequestID: message.RequestID,
	}

	if !success && errorMsg != "" {
		response.Error = errorMsg
	}

	select {
	case c.send <- response:
	default:
		c.hub.logger.Warn("Failed to send unsubscribe response")
	}
}

// handlePing handles ping messages
func (c *Client) handlePing(message *models.ClientMessage) {
	response := &models.ServerMessage{
		Type:      models.MessageTypePong,
		RequestID: message.RequestID,
	}

	select {
	case c.send <- response:
	default:
		c.hub.logger.Warn("Failed to send pong response")
	}
}

// handleHealthWS handles WebSocket health check requests
func (c *Client) handleHealthWS(message *models.ClientMessage) {
	response := &models.ServerMessage{
		Type:      models.MessageTypeHealth,
		Success:   true,
		RequestID: message.RequestID,
		Data: map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	select {
	case c.send <- response:
	default:
		c.hub.logger.Warn("Failed to send health response")
	}
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		// Log error but don't change response status since headers are already sent
		log.Printf("Failed to encode health check response: %v", err)
	}
}
