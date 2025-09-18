package server

import (
	"net/http"
	"os"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		allowedOrigins string
		requestOrigin  string
		expectedResult bool
		description    string
	}{
		// Development mode tests
		{
			name:           "dev_localhost_http",
			env:            "development",
			allowedOrigins: "",
			requestOrigin:  "http://localhost:3000",
			expectedResult: true,
			description:    "Development mode should allow localhost HTTP",
		},
		{
			name:           "dev_localhost_https",
			env:            "development",
			allowedOrigins: "",
			requestOrigin:  "https://localhost:3000",
			expectedResult: true,
			description:    "Development mode should allow localhost HTTPS",
		},
		{
			name:           "dev_127_0_0_1",
			env:            "development",
			allowedOrigins: "",
			requestOrigin:  "http://127.0.0.1:8080",
			expectedResult: true,
			description:    "Development mode should allow 127.0.0.1",
		},
		{
			name:           "dev_empty_origin",
			env:            "development",
			allowedOrigins: "",
			requestOrigin:  "",
			expectedResult: true,
			description:    "Development mode should allow empty origin (same-origin)",
		},
		{
			name:           "dev_external_domain",
			env:            "development",
			allowedOrigins: "",
			requestOrigin:  "https://evil.com",
			expectedResult: false,
			description:    "Development mode should reject external domains",
		},
		{
			name:           "dev_with_custom_origins",
			env:            "development",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "https://myapp.com",
			expectedResult: true,
			description:    "Development mode should allow custom origins",
		},

		// Production mode tests - no AKTUELL_ALLOWED_ORIGINS
		{
			name:           "prod_no_config_localhost",
			env:            "production",
			allowedOrigins: "",
			requestOrigin:  "http://localhost:3000",
			expectedResult: false,
			description:    "Production mode without config should reject localhost",
		},
		{
			name:           "prod_no_config_external",
			env:            "production",
			allowedOrigins: "",
			requestOrigin:  "https://myapp.com",
			expectedResult: false,
			description:    "Production mode without config should reject all origins",
		},
		{
			name:           "prod_no_config_empty",
			env:            "production",
			allowedOrigins: "",
			requestOrigin:  "",
			expectedResult: false,
			description:    "Production mode without config should reject empty origin",
		},

		// Production mode tests - with AKTUELL_ALLOWED_ORIGINS
		{
			name:           "prod_with_config_allowed",
			env:            "production",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "https://myapp.com",
			expectedResult: true,
			description:    "Production mode should allow configured origins",
		},
		{
			name:           "prod_with_config_allowed_second",
			env:            "production",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "https://dashboard.myapp.com",
			expectedResult: true,
			description:    "Production mode should allow second configured origin",
		},
		{
			name:           "prod_with_config_rejected",
			env:            "production",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "https://evil.com",
			expectedResult: false,
			description:    "Production mode should reject non-configured origins",
		},
		{
			name:           "prod_with_config_localhost_rejected",
			env:            "production",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "http://localhost:3000",
			expectedResult: false,
			description:    "Production mode should reject localhost even with config",
		},
		{
			name:           "prod_with_config_empty_rejected",
			env:            "production",
			allowedOrigins: "https://myapp.com,https://dashboard.myapp.com",
			requestOrigin:  "",
			expectedResult: false,
			description:    "Production mode should reject empty origin even with config",
		},

		// Edge cases
		{
			name:           "prod_whitespace_in_origins",
			env:            "production",
			allowedOrigins: " https://myapp.com , https://dashboard.myapp.com ",
			requestOrigin:  "https://myapp.com",
			expectedResult: true,
			description:    "Should handle whitespace in AKTUELL_ALLOWED_ORIGINS",
		},
		{
			name:           "unset_env_localhost",
			env:            "",
			allowedOrigins: "",
			requestOrigin:  "http://localhost:3000",
			expectedResult: true,
			description:    "Unset environment should behave like development mode",
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

			// Create test request
			req := &http.Request{
				Header: make(http.Header),
			}
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			// Test the function
			result := checkOrigin(req)

			// Verify result
			if result != tt.expectedResult {
				t.Errorf("%s: expected %v, got %v - %s",
					tt.name, tt.expectedResult, result, tt.description)
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

func TestGetDefaultAllowedOrigins(t *testing.T) {
	tests := []struct {
		name        string
		env         string
		expectedLen int
		description string
	}{
		{
			name:        "development_mode",
			env:         "development",
			expectedLen: 3,
			description: "Development mode should return default localhost origins",
		},
		{
			name:        "production_mode",
			env:         "production",
			expectedLen: 0,
			description: "Production mode should return empty slice",
		},
		{
			name:        "unset_env",
			env:         "",
			expectedLen: 3,
			description: "Unset environment should behave like development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original environment value
			originalEnv := os.Getenv("AKTUELL_ENV")

			// Set test environment
			if tt.env != "" {
				os.Setenv("AKTUELL_ENV", tt.env)
			} else {
				os.Unsetenv("AKTUELL_ENV")
			}

			// Test the function
			origins := getDefaultAllowedOrigins()

			// Verify result
			if len(origins) != tt.expectedLen {
				t.Errorf("%s: expected %d origins, got %d - %s",
					tt.name, tt.expectedLen, len(origins), tt.description)
			}

			// In development mode, verify we have localhost origins
			if tt.env != "production" && tt.expectedLen > 0 {
				foundLocalhost := false
				for _, origin := range origins {
					if origin == "http://localhost:3000" {
						foundLocalhost = true
						break
					}
				}
				if !foundLocalhost {
					t.Errorf("%s: expected to find localhost:3000 in default origins", tt.name)
				}
			}

			// Restore original environment
			if originalEnv != "" {
				os.Setenv("AKTUELL_ENV", originalEnv)
			} else {
				os.Unsetenv("AKTUELL_ENV")
			}
		})
	}
}

// Benchmark the origin checking function
func BenchmarkCheckOrigin(b *testing.B) {
	// Set up test environment
	os.Setenv("AKTUELL_ENV", "production")
	os.Setenv("AKTUELL_ALLOWED_ORIGINS", "https://app1.com,https://app2.com,https://app3.com")
	defer func() {
		os.Unsetenv("AKTUELL_ENV")
		os.Unsetenv("AKTUELL_ALLOWED_ORIGINS")
	}()

	req := &http.Request{
		Header: make(http.Header),
	}
	req.Header.Set("Origin", "https://app2.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkOrigin(req)
	}
}
