// Initialize MongoDB replica set for change streams
// Use 127.0.0.1 for initialization, but the external apps will connect via 'mongodb' hostname
rs.initiate({
    _id: "rs0",
    version: 1,
    members: [
        { _id: 0, host: "127.0.0.1:27017" }
    ]
});

// Wait for replica set to be ready
while (rs.status().ok !== 1) {
    sleep(1000);
}

print("MongoDB replica set initialized successfully!");

// After initialization, reconfigure to use the external hostname
// This allows both internal initialization and external connections
var config = rs.conf();
config.members[0].host = "mongodb:27017";
config.version++;

rs.reconfig(config, {force: true});

print("MongoDB replica set reconfigured for Docker network!");