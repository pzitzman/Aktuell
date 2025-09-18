// Initialize MongoDB replica set for change streams
rs.initiate({
    _id: "rs0",
    version: 1,
    members: [
        { _id: 0, host: "localhost:27017" }
    ]
});

// Wait for replica set to be ready
while (rs.status().ok !== 1) {
    sleep(1000);
}

print("MongoDB replica set initialized successfully!");