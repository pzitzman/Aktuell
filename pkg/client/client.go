package client

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"aktuell/pkg/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ChangeHandler is a function type for handling change events
type ChangeHandler func(*models.ChangeEvent)

// SnapshotHandler is a function type for handling snapshot events
type SnapshotHandler func(documents []map[string]interface{}, batchNum int, remaining int)

// SnapshotCompleteHandler is a function type for handling snapshot completion
type SnapshotCompleteHandler func()

// ErrorHandler is a function type for handling errors
type ErrorHandler func(error)

// Client represents a Aktuell client that connects to the server
type Client struct {
	serverURL                string
	conn                     *websocket.Conn
	logger                   *logrus.Logger
	mu                       sync.RWMutex
	connected                bool
	reconnecting             bool
	handlers                 map[string]ChangeHandler
	snapshotHandlers         map[string]SnapshotHandler
	snapshotCompleteHandlers map[string]SnapshotCompleteHandler
	errorHandlers            map[string]ErrorHandler
	subscriptions            map[string]*models.Subscription
	doneCh                   chan struct{}
	reconnectCh              chan struct{}
}

// ClientOptions represents configuration options for the client
type ClientOptions struct {
	Logger        *logrus.Logger
	ReconnectWait time.Duration
	PingInterval  time.Duration
}

// NewClient creates a new Aktuell client
func NewClient(serverURL string, opts *ClientOptions) *Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	if opts.Logger == nil {
		opts.Logger = logrus.New()
	}

	if opts.ReconnectWait == 0 {
		opts.ReconnectWait = 5 * time.Second
	}

	if opts.PingInterval == 0 {
		opts.PingInterval = 30 * time.Second
	}

	return &Client{
		serverURL:                serverURL,
		logger:                   opts.Logger,
		handlers:                 make(map[string]ChangeHandler),
		snapshotHandlers:         make(map[string]SnapshotHandler),
		snapshotCompleteHandlers: make(map[string]SnapshotCompleteHandler),
		errorHandlers:            make(map[string]ErrorHandler),
		subscriptions:            make(map[string]*models.Subscription),
		doneCh:                   make(chan struct{}),
		reconnectCh:              make(chan struct{}, 1),
	}
}

// Connect establishes a WebSocket connection to the Aktuell server
func (c *Client) Connect() error {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return err
	}

	c.logger.WithField("server", c.serverURL).Info("Connecting to Aktuell server")

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	// Start message handling
	go c.readMessages()
	go c.pingHandler()

	c.logger.Info("Connected to Aktuell server")
	return nil
}

// Disconnect closes the connection to the server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	close(c.doneCh)

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}

	c.logger.Info("Disconnected from Aktuell server")
	return nil
}

// Subscribe subscribes to changes for a specific database and collection
func (c *Client) Subscribe(database, collection string) error {
	return c.SubscribeWithHandler(database, collection, nil)
}

// SubscribeWithHandler subscribes to changes with a specific handler function
func (c *Client) SubscribeWithHandler(database, collection string, handler ChangeHandler) error {
	return c.SubscribeWithOptions(database, collection, nil, handler, nil, nil, nil)
}

// SubscribeWithSnapshot subscribes to changes with snapshot support
func (c *Client) SubscribeWithSnapshot(database, collection string, snapOpts *models.SnapshotOptions) error {
	return c.SubscribeWithOptions(database, collection, snapOpts, nil, nil, nil, nil)
}

// SubscribeWithOptions subscribes to changes with full options and handlers
func (c *Client) SubscribeWithOptions(
	database, collection string,
	snapOpts *models.SnapshotOptions,
	changeHandler ChangeHandler,
	snapshotHandler SnapshotHandler,
	snapshotCompleteHandler SnapshotCompleteHandler,
	errorHandler ErrorHandler,
) error {
	subscriptionID := uuid.New().String()
	requestID := uuid.New().String()

	subscription := &models.Subscription{
		ID:              subscriptionID,
		Database:        database,
		Collection:      collection,
		CreatedAt:       time.Now(),
		SnapshotOptions: snapOpts,
	}

	message := &models.ClientMessage{
		Type:            models.MessageTypeSubscribe,
		Database:        database,
		Collection:      collection,
		RequestID:       requestID,
		SnapshotOptions: snapOpts,
	}

	c.mu.Lock()
	c.subscriptions[subscriptionID] = subscription
	if changeHandler != nil {
		c.handlers[subscriptionID] = changeHandler
	}
	if snapshotHandler != nil {
		c.snapshotHandlers[subscriptionID] = snapshotHandler
	}
	if snapshotCompleteHandler != nil {
		c.snapshotCompleteHandlers[subscriptionID] = snapshotCompleteHandler
	}
	if errorHandler != nil {
		c.errorHandlers[subscriptionID] = errorHandler
	}
	c.mu.Unlock()

	return c.sendMessage(message)
}

// Unsubscribe removes all subscriptions
func (c *Client) Unsubscribe() error {
	message := &models.ClientMessage{
		Type:      models.MessageTypeUnsubscribe,
		RequestID: uuid.New().String(),
	}

	c.mu.Lock()
	c.subscriptions = make(map[string]*models.Subscription)
	c.handlers = make(map[string]ChangeHandler)
	c.snapshotHandlers = make(map[string]SnapshotHandler)
	c.snapshotCompleteHandlers = make(map[string]SnapshotCompleteHandler)
	c.errorHandlers = make(map[string]ErrorHandler)
	c.mu.Unlock()

	return c.sendMessage(message)
}

// OnChange sets a global change handler for all subscriptions
func (c *Client) OnChange(handler ChangeHandler) {
	c.mu.Lock()
	c.handlers["global"] = handler
	c.mu.Unlock()
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// sendMessage sends a message to the server
func (c *Client) sendMessage(message *models.ClientMessage) error {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return ErrNotConnected
	}

	return conn.WriteJSON(message)
}

// readMessages handles incoming messages from the server
func (c *Client) readMessages() {
	defer func() {
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.connected = false
		c.mu.Unlock()

		// Trigger reconnection if not intentionally disconnected
		if !c.reconnecting {
			select {
			case c.reconnectCh <- struct{}{}:
			default:
			}
		}
	}()

	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		var message models.ServerMessage
		if err := conn.ReadJSON(&message); err != nil {
			c.logger.WithError(err).Error("Failed to read message from server")
			return
		}

		c.handleMessage(&message)
	}
}

// handleMessage processes incoming server messages
func (c *Client) handleMessage(message *models.ServerMessage) {
	switch message.Type {
	case models.MessageTypeChange:
		c.handleChangeEvent(message.Change)
	case models.MessageTypeSnapshot:
		c.handleSnapshotBatch(message)
	case models.MessageTypeSnapshotStart:
		c.handleSnapshotStart(message)
	case models.MessageTypeSnapshotEnd:
		c.handleSnapshotEnd(message)
	case models.MessageTypeError:
		c.handleError(message)
	case models.MessageTypePong:
		c.logger.Debug("Received pong from server")
	default:
		c.logger.WithField("type", message.Type).Debug("Received message from server")
	}
}

// handleChangeEvent handles change events from the server
func (c *Client) handleChangeEvent(change *models.ChangeEvent) {
	if change == nil {
		return
	}

	c.logger.WithFields(logrus.Fields{
		"operation":  change.OperationType,
		"database":   change.Database,
		"collection": change.Collection,
	}).Debug("Received change event")

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Call global handler if it exists
	if handler, exists := c.handlers["global"]; exists {
		go handler(change)
	}

	// Call specific subscription handlers
	for subscriptionID, subscription := range c.subscriptions {
		if c.matchesSubscription(change, subscription) {
			if handler, exists := c.handlers[subscriptionID]; exists {
				go handler(change)
			}
		}
	}
}

// handleSnapshotBatch handles snapshot batch messages from the server
func (c *Client) handleSnapshotBatch(message *models.ServerMessage) {
	if message.SnapshotData == nil {
		return
	}

	c.logger.WithFields(logrus.Fields{
		"batch":     message.SnapshotBatch,
		"remaining": message.SnapshotRemaining,
		"documents": len(message.SnapshotData),
	}).Debug("Received snapshot batch")

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Call snapshot handlers for all subscriptions
	// In a more sophisticated implementation, we'd match handlers to specific subscriptions
	for subscriptionID := range c.snapshotHandlers {
		if handler, exists := c.snapshotHandlers[subscriptionID]; exists {
			go handler(message.SnapshotData, message.SnapshotBatch, message.SnapshotRemaining)
		}
	}
}

// handleSnapshotStart handles snapshot start messages from the server
func (c *Client) handleSnapshotStart(message *models.ServerMessage) {
	c.logger.Info("Snapshot streaming started")
}

// handleSnapshotEnd handles snapshot end messages from the server
func (c *Client) handleSnapshotEnd(message *models.ServerMessage) {
	c.logger.Info("Snapshot streaming completed")

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Call snapshot complete handlers for all subscriptions
	for subscriptionID := range c.snapshotCompleteHandlers {
		if handler, exists := c.snapshotCompleteHandlers[subscriptionID]; exists {
			go handler()
		}
	}
}

// handleError handles error messages from the server
func (c *Client) handleError(message *models.ServerMessage) {
	c.logger.WithField("error", message.Error).Error("Server error")

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Call error handlers for all subscriptions
	for subscriptionID := range c.errorHandlers {
		if handler, exists := c.errorHandlers[subscriptionID]; exists {
			go handler(fmt.Errorf("server error: %s", message.Error))
		}
	}
}

// matchesSubscription checks if a change event matches a subscription
func (c *Client) matchesSubscription(change *models.ChangeEvent, subscription *models.Subscription) bool {
	if subscription.Database != "" && subscription.Database != change.Database {
		return false
	}

	if subscription.Collection != "" && subscription.Collection != change.Collection {
		return false
	}

	// TODO: Implement filter matching logic
	// For now, we'll match all changes that pass database and collection filters

	return true
}

// pingHandler sends periodic ping messages to keep the connection alive
func (c *Client) pingHandler() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.doneCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.connected
			c.mu.RUnlock()

			if connected {
				message := &models.ClientMessage{
					Type:      models.MessageTypePing,
					RequestID: uuid.New().String(),
				}
				if err := c.sendMessage(message); err != nil {
					c.logger.WithError(err).Error("Failed to send ping")
				}
			}
		}
	}
}

// EnableAutoReconnect enables automatic reconnection when the connection is lost
func (c *Client) EnableAutoReconnect(wait time.Duration) {
	go c.reconnectHandler(wait)
}

// reconnectHandler handles automatic reconnection
func (c *Client) reconnectHandler(wait time.Duration) {
	for {
		select {
		case <-c.doneCh:
			return
		case <-c.reconnectCh:
			c.mu.Lock()
			c.reconnecting = true
			c.mu.Unlock()

			c.logger.WithField("wait", wait).Info("Connection lost, attempting to reconnect")
			time.Sleep(wait)

			if err := c.Connect(); err != nil {
				c.logger.WithError(err).Error("Failed to reconnect, will retry")
				// Trigger another reconnection attempt
				select {
				case c.reconnectCh <- struct{}{}:
				default:
				}
			} else {
				c.mu.Lock()
				c.reconnecting = false
				c.mu.Unlock()

				// Re-establish subscriptions
				c.resubscribe()
			}
		}
	}
}

// resubscribe re-establishes all subscriptions after reconnection
func (c *Client) resubscribe() {
	c.mu.RLock()
	subscriptions := make([]*models.Subscription, 0, len(c.subscriptions))
	for _, sub := range c.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	c.mu.RUnlock()

	for _, sub := range subscriptions {
		// Re-subscribe without filter since server delivers all events
		if err := c.Subscribe(sub.Database, sub.Collection); err != nil {
			c.logger.WithError(err).Error("Failed to re-establish subscription")
		}
	}

	c.logger.Info("Re-established subscriptions after reconnection")
}

// Custom errors
var (
	ErrNotConnected = fmt.Errorf("not connected to server")
)
