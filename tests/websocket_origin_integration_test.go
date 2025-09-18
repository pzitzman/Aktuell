//go:build integration
// +build integration

package tests

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"aktuell/pkg/server"
)

func TestWebSocketOriginSecurity(t *testing.T) {
	// Create a WebSocket server for testing
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	wsServer := server.NewWebSocketServer("localhost:0", logger)

	// Start server in background
	go func() {
		if err := wsServer.Start(); err != nil {
			logger.WithError(err).Debug("Server stopped")
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	serverAddr := wsServer.GetAddr()
	require.NotEmpty(t, serverAddr, "Server should have started with valid address")

	defer wsServer.Stop()

	tests := []struct {
		name             string
		env              string
		allowedOrigins   string
		requestOrigin    string
		expectConnection bool
		description      string
	}{
		// Development mode tests
		{
			name:             "dev_localhost_allowed",
			env:              "development",
			allowedOrigins:   "",
			requestOrigin:    "http://localhost:3000",
			expectConnection: true,
			description:      "Development mode should allow localhost",
		},
		{
			name:             "dev_external_rejected",
			env:              "development",
			allowedOrigins:   "",
			requestOrigin:    "https://evil.com",
			expectConnection: false,
			description:      "Development mode should reject external domains",
		},

		// Production mode tests
		{
			name:             "prod_no_config_all_rejected",
			env:              "production",
			allowedOrigins:   "",
			requestOrigin:    "https://myapp.com",
			expectConnection: false,
			description:      "Production without config should reject all",
		},
		{
			name:             "prod_with_config_allowed",
			env:              "production",
			allowedOrigins:   "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:    "https://myapp.com",
			expectConnection: true,
			description:      "Production should allow configured origins",
		},
		{
			name:             "prod_with_config_rejected",
			env:              "production",
			allowedOrigins:   "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:    "https://evil.com",
			expectConnection: false,
			description:      "Production should reject non-configured origins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original environment values
			originalEnv := os.Getenv("AKTUELL_ENV")
			originalOrigins := os.Getenv("AKTUELL_ALLOWED_ORIGINS")

			// Set test environment
			if tt.env != "" {
				os.Setenv("AKTUELL_ENV", tt.env)
			} else {
				os.Unsetenv("AKTUELL_ENV")
			}

			if tt.allowedOrigins != "" {
				os.Setenv("AKTUELL_ALLOWED_ORIGINS", tt.allowedOrigins)
			} else {
				os.Unsetenv("AKTUELL_ALLOWED_ORIGINS")
			}

			// Create WebSocket dialer with custom origin
			dialer := websocket.DefaultDialer
			headers := make(map[string][]string)
			if tt.requestOrigin != "" {
				headers["Origin"] = []string{tt.requestOrigin}
			}

			// Attempt connection
			wsURL := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}
			conn, resp, err := dialer.Dial(wsURL.String(), headers)

			if tt.expectConnection {
				// Should succeed
				assert.NoError(t, err, "Connection should succeed for %s", tt.description)
				assert.NotNil(t, conn, "Connection should be established")
				if conn != nil {
					conn.Close()
				}
			} else {
				// Should fail with 403 Forbidden (origin rejected)
				assert.Error(t, err, "Connection should fail for %s", tt.description)
				assert.Nil(t, conn, "Connection should not be established")
				if resp != nil {
					assert.Equal(t, 403, resp.StatusCode, "Should get 403 Forbidden for rejected origin")
				}
			}

			if resp != nil {
				resp.Body.Close()
			}

			// Restore original environment
			if originalEnv != "" {
				os.Setenv("AKTUELL_ENV", originalEnv)
			} else {
				os.Unsetenv("AKTUELL_ENV")
			}

			if originalOrigins != "" {
				os.Setenv("AKTUELL_ALLOWED_ORIGINS", originalOrigins)
			} else {
				os.Unsetenv("AKTUELL_ALLOWED_ORIGINS")
			}
		})
	}
}

// Test concurrent connections with different origins
func TestWebSocketOriginSecurityConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent origin test in short mode")
	}

	// Create a WebSocket server for testing
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	wsServer := server.NewWebSocketServer("localhost:0", logger)

	// Start server in background
	go func() {
		if err := wsServer.Start(); err != nil {
			logger.WithError(err).Debug("Server stopped")
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	serverAddr := wsServer.GetAddr()
	require.NotEmpty(t, serverAddr, "Server should have started with valid address")

	defer wsServer.Stop()

	// Set production mode with specific allowed origins
	originalEnv := os.Getenv("AKTUELL_ENV")
	originalOrigins := os.Getenv("AKTUELL_ALLOWED_ORIGINS")
	defer func() {
		if originalEnv != "" {
			os.Setenv("AKTUELL_ENV", originalEnv)
		} else {
			os.Unsetenv("AKTUELL_ENV")
		}
		if originalOrigins != "" {
			os.Setenv("AKTUELL_ALLOWED_ORIGINS", originalOrigins)
		} else {
			os.Unsetenv("AKTUELL_ALLOWED_ORIGINS")
		}
	}()

	os.Setenv("AKTUELL_ENV", "production")
	os.Setenv("AKTUELL_ALLOWED_ORIGINS", "https://app1.com,https://app2.com")

	// Test multiple concurrent connections
	origins := []struct {
		origin   string
		expected bool
	}{
		{"https://app1.com", true},
		{"https://app2.com", true},
		{"https://evil1.com", false},
		{"https://evil2.com", false},
		{"http://localhost:3000", false},
	}

	results := make(chan bool, len(origins))

	// Launch concurrent connection attempts
	for _, test := range origins {
		go func(origin string, expected bool) {
			dialer := websocket.DefaultDialer
			headers := make(map[string][]string)
			headers["Origin"] = []string{origin}

			wsURL := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}
			conn, resp, err := dialer.Dial(wsURL.String(), headers)

			success := err == nil && conn != nil
			if conn != nil {
				conn.Close()
			}
			if resp != nil {
				resp.Body.Close()
			}

			results <- success == expected
		}(test.origin, test.expected)
	}

	// Collect results
	successCount := 0
	for i := 0; i < len(origins); i++ {
		if <-results {
			successCount++
		}
	}

	assert.Equal(t, len(origins), successCount, "All origin security tests should pass")
}
