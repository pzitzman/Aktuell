#!/bin/bash

# Simple MongoDB Docker Operations
# Quick commands to modify users table

CONTAINER="aktuell-db"
DB="aktuell"

echo "Simple MongoDB Operations for Docker"
echo "===================================="

case "$1" in
    "insert")
        echo "Inserting a new user..."
        docker exec $CONTAINER mongosh $DB --eval "
        db.users.insertOne({
          name: 'User $(date +%s)',
          email: 'user$(date +%s)@example.com',
          age: $((RANDOM % 50 + 20)),
          salary: Math.floor(Math.random() * 80000) + 40000,
          loginCount: Math.floor(Math.random() * 100) + 1,
          status: ['active', 'inactive', 'pending'][Math.floor(Math.random() * 3)],
          skills: [['JavaScript', 'React'], ['Python', 'Django'], ['Go', 'MongoDB'], ['Java', 'Spring'], ['Node.js', 'Express']][Math.floor(Math.random() * 5)],
          department: ['Engineering', 'Marketing', 'Sales', 'HR', 'Finance'][Math.floor(Math.random() * 5)],
          createdAt: new Date()
        })
        "
        ;;
    "update")
        echo "Updating random user..."
        docker exec $CONTAINER mongosh $DB --eval "
        db.users.updateOne(
          {},
          { \$set: { lastUpdated: new Date(), status: 'updated', salary: Math.floor(Math.random() * 50000) + 50000 } }
        )
        "
        ;;
    "delete")
        echo "Deleting one user..."
        docker exec $CONTAINER mongosh $DB --eval "
        db.users.deleteOne({})
        "
        ;;
    "count")
        echo "Counting users..."
        docker exec $CONTAINER mongosh $DB --eval "db.users.countDocuments()" --quiet
        ;;
    "list")
        echo "Listing all users..."
        docker exec $CONTAINER mongosh $DB --eval "db.users.find().pretty()" --quiet
        ;;
    "clear")
        echo "Clearing all users..."
        docker exec $CONTAINER mongosh $DB --eval "db.users.deleteMany({})"
        ;;
    *)
        echo "Usage: $0 {insert|update|delete|count|list|clear}"
        echo ""
        echo "Commands:"
        echo "  insert  - Insert a new random user"
        echo "  update  - Update a random user"
        echo "  delete  - Delete one user"
        echo "  count   - Count total users"
        echo "  list    - List all users"
        echo "  clear   - Delete all users"
        exit 1
        ;;
esac