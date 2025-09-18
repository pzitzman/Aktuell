# Aktuell - Real-Time MongoDB Change Streams

[![CI](https://github.com/pzitzman/aktuell/actions/workflows/ci.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/ci.yml)
[![Release](https://github.com/pzitzman/aktuell/actions/workflows/release.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/release.yml)
[![Docker](https://github.com/pzitzman/aktuell/actions/workflows/docker.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pzitzman/aktuell)](https://goreportcard.com/report/github.com/pzitzman/aktuell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Aktuell** (German for "current/up-to-date") is a high-performance, real-time MongoDB change streams monitoring system with WebSocket-based live updates and a modern React dashboard.

## 🚀 Features

- **Real-time Change Streams**: Monitor MongoDB collections with instant updates
- **WebSocket Integration**: Live data synchronization with sub-second latency
- **Example Dashboard**: React TypeScript interface with real-time visualization
- **Multi-platform Support**: Runs on Linux, macOS, and Windows
- **Docker Ready**: Containerized deployment with multi-architecture support

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │    │                 │
│  React Client   │◄───┤  WebSocket API  │◄───┤    MongoDB      │
│   (Dashboard)   │    │   (Go Server)   │    │ Change Streams  │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- MongoDB 4.0+ (for change streams support)
- Docker (optional)

### Installation

```bash
git clone <repository-url>
cd Aktuell
go mod download
```

### Troubleshooting Go Version Issues

If you encounter Go version mismatch errors (e.g., `compile: version "go1.22.1" does not match go tool version "go1.25.1"`):

The project includes universal environment detection that works with any Go installation:

```bash
# Option 1: Use the Makefile (automatically detects and fixes environment)
make build

# Option 2: Run the environment fix script manually
./scripts/fix-go-env.sh

# Option 3: Source the Go environment fix
source .env.go
go build -o aktuell ./cmd/server
```

### Running the Server

```bash
go run cmd/server/main.go
```

### Using the Client SDK

The Aktuell client SDK supports subscribing to multiple databases and collections:

```go
package main

import (
    "log"
    "time"
    "aktuell/pkg/client"
    "aktuell/pkg/models"
)

func main() {
    // Create client
    c := client.NewClient("ws://localhost:8080/ws", &client.ClientOptions{})

    // Set up change handler
    c.OnChange(func(change *models.ChangeEvent) {
        log.Printf("📦 Change in %s.%s: %s", 
            change.Database, change.Collection, change.OperationType)
    })
    
    // Connect to server
    if err := c.Connect(); err != nil {
        log.Fatal(err)
    }
    defer c.Disconnect()
    
    // Subscribe to multiple databases and collections
    c.Subscribe("InventoryDB", "Products", nil)
    c.Subscribe("InventoryDB", "Orders", nil)
    c.Subscribe("LogsDB", "SystemLogs", nil)
    
    // Enable auto-reconnect
    c.EnableAutoReconnect(5 * time.Second)
    
    select {} // Keep running
}
```

## Configuration

Aktuell supports both single-database and multi-database configurations:

### Multi-Database Configuration (Recommended)

```yaml
mongodb:
  uri: "mongodb://localhost:27017"
  # Configure multiple databases with their collections
  databases:
    - name: "InventoryDB"
      collections:
        - "Products"
        - "Orders"
        - "Customers"
    - name: "LogsDB"
      collections:
        - "SystemLogs"
        - "ErrorLogs"

server:
  port: 8080
  host: "localhost"
  
logging:
  level: "info"
```

### Legacy Single-Database Configuration (Still Supported)

```yaml
mongodb:
  uri: "mongodb://localhost:27017"
  database: "myapp"
  collections:
    - "users"
    - "posts"
  
server:
  port: 8080
  host: "localhost"
  
logging:
  level: "info"
```

### Environment Variables

You can also configure using environment variables with the `AKTUELL_` prefix:

```bash
export AKTUELL_MONGODB_URI="mongodb://localhost:27017"
export AKTUELL_SERVER_HOST="0.0.0.0"
export AKTUELL_SERVER_PORT="8080"
export AKTUELL_LOGGING_LEVEL="debug"
```

## Project Structure

```
Aktuell/
├── .github/             # GitHub Actions workflows and templates
│   └── workflows/       # CI/CD pipeline configurations
├── cmd/
│   └── server/          # Server application entry point
├── pkg/
│   ├── client/          # Client SDK
│   ├── server/          # WebSocket server implementation
│   ├── sync/            # Synchronization manager
│   └── models/          # Data models and types
├── react-client/        # React TypeScript dashboard
│   ├── src/             # React source code
│   ├── public/          # Static assets
│   └── package.json     # React dependencies
├── tests/               # Integration and end-to-end tests
├── examples/            # Example applications
├── docker/              # Docker configuration
└── docs/                # Documentation
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to our repository.

## License

MIT License - see LICENSE file for details.