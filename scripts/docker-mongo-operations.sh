#!/bin/bash

# MongoDB Docker Operations Script
# This script demonstrates various MongoDB operations using docker exec commands

set -e  # Exit on any error

CONTAINER_NAME="aktuell-db"
DATABASE="aktuell"
COLLECTION="users"

echo "🚀 MongoDB Docker Operations Script"
echo "====================================="

# Check if container is running
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo "❌ Error: Container '$CONTAINER_NAME' is not running"
    echo "Please start it with: docker-compose up -d"
    exit 1
fi

echo "✅ Container '$CONTAINER_NAME' is running"
echo ""

# Function to execute MongoDB command
execute_mongo_cmd() {
    local cmd="$1"
    local description="$2"
    
    echo "📝 $description"
    echo "Command: $cmd"
    echo "Result:"
    docker exec "$CONTAINER_NAME" mongosh "$DATABASE" --eval "$cmd" --quiet
    echo ""
}

# 1. Show current data count
execute_mongo_cmd "db.$COLLECTION.countDocuments()" "Current document count"

# 2. Insert sample users
echo "📝 Inserting sample users..."
docker exec "$CONTAINER_NAME" mongosh "$DATABASE" --eval "
db.$COLLECTION.insertMany([
  {
    name: 'Alice Johnson',
    email: 'alice@example.com',
    age: 28,
    department: 'Engineering',
    createdAt: new Date(),
    status: 'active'
  },
  {
    name: 'Bob Smith',
    email: 'bob@example.com',
    age: 35,
    department: 'Marketing',
    createdAt: new Date(),
    status: 'active'
  },
  {
    name: 'Carol Davis',
    email: 'carol@example.com',
    age: 42,
    department: 'Sales',
    createdAt: new Date(),
    status: 'inactive'
  },
  {
    name: 'David Wilson',
    email: 'david@example.com',
    age: 31,
    department: 'Engineering',
    createdAt: new Date(),
    status: 'active'
  }
])
" --quiet
echo "✅ Sample users inserted"
echo ""

# 3. Show all users
execute_mongo_cmd "db.$COLLECTION.find().pretty()" "All users in database"

# 4. Update users in Engineering department
execute_mongo_cmd "
db.$COLLECTION.updateMany(
  { department: 'Engineering' },
  { 
    \$set: { 
      salary: 85000,
      lastUpdated: new Date(),
      skills: ['JavaScript', 'Go', 'MongoDB']
    } 
  }
)" "Update Engineering department users"

# 5. Update specific user by email
execute_mongo_cmd "
db.$COLLECTION.updateOne(
  { email: 'bob@example.com' },
  { 
    \$set: { 
      age: 36,
      salary: 75000,
      lastUpdated: new Date()
    } 
  }
)" "Update Bob's information"

# 6. Find users over 30
execute_mongo_cmd "db.$COLLECTION.find({ age: { \$gt: 30 } }).pretty()" "Users over 30 years old"

# 7. Find active Engineering users
execute_mongo_cmd "
db.$COLLECTION.find({ 
  department: 'Engineering',
  status: 'active' 
}).pretty()
" "Active Engineering users"

# 8. Add a new field to all users
execute_mongo_cmd "
db.$COLLECTION.updateMany(
  {},
  { 
    \$set: { 
      lastLogin: new Date(),
      loginCount: Math.floor(Math.random() * 100) + 1
    } 
  }
)" "Add login tracking to all users"

# 9. Delete inactive users
execute_mongo_cmd "
db.$COLLECTION.deleteMany({ status: 'inactive' })
" "Delete inactive users"

# 10. Show final count and remaining users
execute_mongo_cmd "db.$COLLECTION.countDocuments()" "Final document count"
execute_mongo_cmd "db.$COLLECTION.find().pretty()" "Remaining users"

# 11. Create an index for better performance
execute_mongo_cmd "
db.$COLLECTION.createIndex({ email: 1 }, { unique: true })
" "Create unique index on email field"

# 12. Show collection stats
execute_mongo_cmd "db.$COLLECTION.stats()" "Collection statistics"

echo "🎉 MongoDB operations completed successfully!"
echo "You can now check your Aktuell WebSocket stream to see the changes in real-time."