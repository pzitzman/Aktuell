# Aktuell - Real-Time MongoDB Change Streams

[![CI](https://github.com/pzitzman/aktuell/actions/workflows/ci.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/ci.yml)
[![Release](https://github.com/pzitzman/aktuell/actions/workflows/release.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/release.yml)
[![Docker](https://github.com/pzitzman/aktuell/actions/workflows/docker.yml/badge.svg)](https://github.com/pzitzman/aktuell/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pzitzman/aktuell)](https://goreportcard.com/report/github.com/pzitzman/aktuell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![Aktuell Logo](./logo.svg)

**Aktuell** (German for "current/up-to-date") is a high-performance, real-time MongoDB change streams monitoring system with WebSocket-based live updates and a modern React dashboard.

Project is in development phase. Feel free to try it out!

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
         │                       │                       │
         │ Subscribe Request     │                       │
         │ w/ Snapshot Options   │                       │
         ├──────────────────────►│                       │
         │                       │                       │
         │                   ┌───▼────┐                  │
         │                   │Snapshot│                  │
         │                   │Manager │                  │
         │                   └───┬────┘                  │
         │                       │                       │
         │                       │ Query Existing Data   │
         │                       ├──────────────────────►│
         │ Snapshot Batch 1/N    │                       │
         │◄──────────────────────┤                       │
         │ Snapshot Batch 2/N    │                       │
         │◄──────────────────────┤                       │
         │ ... (batched data)    │                       │
         │ Snapshot Complete     │                       │
         │◄──────────────────────┤                       │
         │                       │                       │
         │ Real-time Changes     │                       │
         │◄──────────────────────┤◄──────────────────────│
```

### Message Flow

1. **Initial Connection**: Client establishes WebSocket connection
2. **Subscribe with Snapshot**: Client sends subscription request with `SnapshotOptions`
3. **Snapshot Streaming**: Server queries existing documents and streams them in batches
4. **Real-time Updates**: After snapshot completion, live change events are streamed

### WebSocket Protocol

**Client → Server Messages:**
```json
{
  "type": "subscribe",
  "database": "inventory", 
  "collection": "products",
  "snapshot_options": {
    "include_snapshot": true,
    "batch_size": 100,
    "snapshot_limit": 1000,
    "snapshot_filter": {"status": "active"},
    "snapshot_sort": {"created_at": -1}
  }
}
```
If include_snapshot is set to false, no initial dataload is retrieved.
And only the changes are propagated.

**Server → Client Messages:**
```json
// Snapshot batch
{
  "type": "snapshot",
  "snapshot_data": [{...}, {...}],
  "snapshot_batch": 1,
  "snapshot_total": 1000,
  "snapshot_remaining": 900
}

// Change event
{
  "type": "change", 
  "change": {
    "operationType": "insert",
    "database": "inventory",
    "collection": "products", 
    "documentKey": {"_id": "..."},
    "fullDocument": {...}
  }
}
```

## ⚡ Quick Setup Guide

Get Aktuell running:

### 1. Start the Backend Services
```bash
# Clone the repository
git clone https://github.com/pzitzman/aktuell.git
cd aktuell

# Start MongoDB and Aktuell server with Docker Compose
docker-compose up -d

# Verify services are running
docker ps
```

### 2. Start the React Frontend
```bash
# In a new terminal, start the React development server
cd react-client
npm install
npm start
```

The React dashboard will open at [http://localhost:3000](http://localhost:3000)

### 3. Generate Test Data

Use any of these commands to create and modify data:

```bash
# Quick data operations
make mongo-insert    # Insert a user with realistic data
make mongo-update    # Update a user with new salary/status
make mongo-count     # Check user count
make mongo-list      # View all users

# Or use the comprehensive demo
make mongo-demo      # Run full demonstration script

# Alternative: Use bash scripts directly
./scripts/mongo-quick.sh insert
./scripts/mongo-quick.sh update
./scripts/mongo-quick.sh list
```

### 4. Watch Real-time Updates

- Open the React dashboard at [http://localhost:3000](http://localhost:3000)
- Run `make mongo-insert` or `make mongo-update` in your terminal
- See the changes appear instantly in the dashboard! ✨

### Services Overview

| Service | URL | Description |
|---------|-----|-------------|
| **React Dashboard** | http://localhost:3000 | Real-time data visualization |
| **WebSocket API** | ws://localhost:8080/ws | Live data stream endpoint |
| **MongoDB** | localhost:27017 | Database (aktuell-db container) |

**That's it!** You now have a complete real-time data streaming system running locally.

## Quick Start without docker

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

## Security Configuration

Aktuell includes built-in security measures for WebSocket connections to prevent CSRF attacks and unauthorized access.

### WebSocket Origin Validation

By default, Aktuell validates the `Origin` header of WebSocket connections:

**Development Mode** (when `AKTUELL_ENV` ≠ "production"):
- Allows all `localhost` and `127.0.0.1` origins
- Permits same-origin requests

**Production Mode** (when `AKTUELL_ENV` = "production"):
- Strictly validates against allowed origins list
- Rejects and logs unauthorized connection attempts

### Configuring Allowed Origins

#### Method 1: Environment Variable (Required for Production)
```bash
# Production: REQUIRED - Comma-separated list of allowed origins
export AKTUELL_ENV="production"
export AKTUELL_ALLOWED_ORIGINS="https://app.example.com,https://dashboard.example.com"

# Development: Optional - uses localhost defaults if not set
export AKTUELL_ENV="development"  # or leave unset
export AKTUELL_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:8080"  # optional
```

**Important**: In production mode (`AKTUELL_ENV=production`):
- You **MUST** set `AKTUELL_ALLOWED_ORIGINS` 
- If `AKTUELL_ALLOWED_ORIGINS` is not set, **all** WebSocket connections will be rejected
- Only origins in `AKTUELL_ALLOWED_ORIGINS` will be allowed

**Development mode** (default when `AKTUELL_ENV` is not "production"):
- Automatically allows `localhost` and `127.0.0.1` origins
- `AKTUELL_ALLOWED_ORIGINS` is optional and adds to the allowed list

### Security Best Practices

1. **Always set `AKTUELL_ENV=production`** in production environments
2. **Always set `AKTUELL_ALLOWED_ORIGINS`** in production environments
3. **Use HTTPS origins** in production (`https://` not `http://`)
4. **Be specific with origins** - avoid wildcards or overly broad patterns
5. **Monitor logs** for rejected connection attempts
6. **Use TLS/SSL** for WebSocket connections in production (WSS)

### Authentication and Authorization

⚠️ **Important Security Notice**: Aktuell itself does not provide built-in authentication or authorization mechanisms. The WebSocket server is designed to be deployed behind authenticated infrastructure components.

**Recommended Production Deployment:**

```
Internet → Load Balancer → Auth Proxy → Aktuell Server
              ↓              ↓            ↓
         TLS Termination  Authentication  WebSocket API
         Rate Limiting    Authorization   Change Streams
```

**Common Authentication Patterns:**

1. **Reverse Proxy with Authentication** (Recommended)
   ```nginx
   # Example Nginx configuration
   server {
       listen 443 ssl;
       server_name api.mycompany.com;
       
       # Authentication middleware
       auth_request /auth;
       
       location /ws {
           proxy_pass http://aktuell:8080;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection "upgrade";
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

2. **API Gateway with JWT Validation**
   - Use cloud API gateways (AWS API Gateway, Google Cloud Endpoints, etc.)
   - Validate JWT tokens before forwarding to Aktuell
   - Implement rate limiting and request throttling

3. **Service Mesh Authentication**
   - Deploy within Kubernetes with service mesh (Istio, Linkerd)
   - Use mutual TLS (mTLS) between services
   - Implement policy-based access control

**Security Considerations:**
- Never expose Aktuell directly to the internet without authentication
- Implement proper session management in your auth layer
- Use database-level access controls for MongoDB
- Consider implementing audit logging for WebSocket connections
- Validate user permissions for database/collection access in your auth layer

### Example Production Configuration

```bash
# Production environment - REQUIRED settings
export AKTUELL_ENV="production"
export AKTUELL_ALLOWED_ORIGINS="https://dashboard.mycompany.com,https://app.mycompany.com"

# Secure MongoDB connection
export AKTUELL_MONGODB_URI="mongodb://username:password@mongo.mycompany.com:27017/myapp?ssl=true"

# Bind to all interfaces securely (behind reverse proxy)
export AKTUELL_SERVER_HOST="0.0.0.0"
export AKTUELL_SERVER_PORT="8080"
```

### Example Development Configuration

```bash
# Development environment - optional settings
export AKTUELL_ENV="development"  # or leave unset
# AKTUELL_ALLOWED_ORIGINS is optional - localhost origins are automatically allowed

# Local MongoDB connection
export AKTUELL_MONGODB_URI="mongodb://localhost:27017"
export AKTUELL_SERVER_HOST="localhost"
export AKTUELL_SERVER_PORT="8080"
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to our repository.

## License

MIT License - see LICENSE file for details.
