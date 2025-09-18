# Aktuell Usage Guide

## Quick Start

### 1. Start MongoDB with Replica Set (Required for Change Streams)

Using Docker:
```bash
make dev-setup
```

Or manually:
```bash
# Start MongoDB with replica set
docker run --name mongodb -p 27017:27017 -d mongo:7.0 --replSet rs0

# Initialize replica set
docker exec -it mongodb mongosh --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'localhost:27017'}]})"
```

### 2. Start Aktuell Server

```bash
# Build and run locally
make run

# Or using Docker
make docker-run
```

### 3. Run Example Client

In a separate terminal:
```bash
make run-client
```

### 4. Generate Test Data

In another terminal:
```bash
make run-generator
```

You should now see real-time change events in the client terminal!

## Configuration

### Environment Variables

```bash
export AKTUELL_MONGODB_URI="mongodb://localhost:27017/aktuell?replicaSet=rs0"
export AKTUELL_MONGODB_DATABASE="aktuell"
export AKTUELL_SERVER_HOST="localhost"
export AKTUELL_SERVER_PORT="8080"
export AKTUELL_LOGGING_LEVEL="info"
```

### Configuration File (config.yaml)

```yaml
mongodb:
  uri: "mongodb://localhost:27017/aktuell?replicaSet=rs0"
  database: "aktuell"
  collections:
    - "users"
    - "posts"
    - "comments"

server:
  host: "localhost"
  port: 8080

logging:
  level: "info"
```

## Client SDK Usage

### Basic Usage

```go
package main

import (
    "aktuell/pkg/client"
    "aktuell/pkg/models"
)

func main() {
    // Create client
    c := client.NewClient("ws://localhost:8080/ws", nil)
    
    // Set change handler
    c.OnChange(func(change *models.ChangeEvent) {
        // Handle change event
        fmt.Printf("Change: %s in %s.%s\n", 
            change.OperationType, 
            change.Database, 
            change.Collection)
    })
    
    // Connect and subscribe
    c.Connect()
    c.Subscribe("mydb", "mycollection", nil)
    
    // Keep running
    select {}
}
```


### Auto-reconnection

```go
// Enable automatic reconnection
c.EnableAutoReconnect(5 * time.Second)
```

## WebSocket API

### Connect
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

### Subscribe to Changes
```javascript
ws.send(JSON.stringify({
    "type": "subscribe",
    "database": "aktuell",
    "collection": "users",
    "requestId": "unique-id"
}));
```

### Receive Changes
```javascript
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    if (message.type === "change") {
        console.log("Change received:", message.change);
    }
};
```

## Production Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  mongodb:
    image: mongo:7.0
    command: ["--replSet", "rs0"]
    
  aktuell:
    build: .
    ports:
      - "8080:8080"
    environment:
     AKTUELL_MONGODB_URI: mongodb://mongodb:27017/myapp?replicaSet=rs0
     AKTUELL_MONGODB_DATABASE: myapp
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aktuell
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aktuell
  template:
    metadata:
      labels:
        app: aktuell
    spec:
      containers:
      - name: aktuell
        image: aktuell:latest
        ports:
        - containerPort: 8080
        env:
        - name:AKTUELL_MONGODB_URI
          value: "mongodb://mongodb:27017/myapp?replicaSet=rs0"
```

## Performance Tuning

### MongoDB Configuration

- Ensure replica set is properly configured
- Use appropriate read preferences
- Configure oplog size appropriately
- Monitor change stream cursor lag

### Aktuell Configuration

- Adjust WebSocket buffer sizes for high throughput
- Configure appropriate timeouts
- Use connection pooling for multiple databases
- Monitor memory usage for large change event volumes

## Troubleshooting

### Common Issues

1. **Change streams not working**
   - Ensure MongoDB is running as a replica set
   - Check MongoDB version (4.0+ required)

2. **WebSocket connection drops**
   - Check network connectivity
   - Verify firewall settings
   - Enable auto-reconnection in client

3. **High memory usage**
   - Reduce change event buffer sizes
   - Implement proper client-side filtering
   - Monitor connection counts

4. **Go version mismatch errors**
   - The Makefile automatically handles Go environment issues
   - If building manually, source the environment fix: `source .env.go`
   - Or use the fix script directly: `./scripts/fix-go-env.sh go build ...`

### Debug Mode

Enable debug logging:
```bash
export AKTUELL_LOGGING_LEVEL=debug
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/client/` - Basic Go client
- `examples/generator/` - Data generator for testing

## API Reference

### Change Event Structure

```json
{
  "id": "change-event-id",
  "operationType": "insert|update|delete|replace",
  "database": "database-name",
  "collection": "collection-name",
  "documentKey": {"_id": "document-id"},
  "fullDocument": {...},
  "updatedFields": {...},
  "removedFields": ["field1", "field2"],
  "timestamp": "2023-01-01T00:00:00Z",
  "clientTimestamp": "2023-01-01T00:00:00Z"
}
```

### Client Message Types

- `subscribe` - Subscribe to changes
- `unsubscribe` - Unsubscribe from changes  
- `ping` - Keep connection alive

### Server Message Types

- `change` - Change event notification
- `error` - Error message
- `pong` - Ping response