package sync

import (
	"context"
	"fmt"
	"time"

	"aktuell/pkg/models"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database wraps MongoDB operations and change stream functionality
type Database struct {
	client        *mongo.Client
	db            *mongo.Database
	connectionURI string
	logger        *logrus.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	changesCh     chan *models.ChangeEvent
}

// NewDatabase creates a new Database instance
func NewDatabase(mongoURI, dbName string, logger *logrus.Logger) (*Database, error) {
	ctx, cancel := context.WithCancel(context.Background())

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := &Database{
		client:        client,
		db:            client.Database(dbName),
		connectionURI: mongoURI,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		changesCh:     make(chan *models.ChangeEvent, 100), // Buffered channel
	}

	logger.WithFields(logrus.Fields{
		"uri":      mongoURI,
		"database": dbName,
	}).Info("Connected to MongoDB")

	return db, nil
}

// StartChangeStream starts monitoring MongoDB change streams
func (d *Database) StartChangeStream(collections []string) error {
	// Pipeline to filter for specific collections if provided
	var pipeline mongo.Pipeline
	if len(collections) > 0 {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "ns.coll", Value: bson.D{
					{Key: "$in", Value: collections},
				}},
			}},
		})
	}

	// Options for change stream
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	// Watch the entire database
	stream, err := d.db.Watch(d.ctx, pipeline, opts)
	if err != nil {
		return fmt.Errorf("failed to create change stream: %w", err)
	}

	go d.processChangeStream(stream)

	d.logger.WithFields(logrus.Fields{
		"database":    d.db.Name(),
		"collections": collections,
	}).Info("Started MongoDB change stream")

	return nil
}

// processChangeStream processes change stream events
func (d *Database) processChangeStream(stream *mongo.ChangeStream) {
	defer stream.Close(d.ctx)

	for stream.Next(d.ctx) {
		var changeDoc bson.M
		if err := stream.Decode(&changeDoc); err != nil {
			d.logger.WithError(err).Error("Failed to decode change stream document")
			continue
		}

		changeEvent := d.parseChangeEvent(changeDoc)
		if changeEvent != nil {
			select {
			case d.changesCh <- changeEvent:
				// Successfully sent to channel
			default:
				d.logger.Warn("Change event channel is full, dropping event")
			}
		}
	}

	if err := stream.Err(); err != nil {
		d.logger.WithError(err).Error("Change stream error")
	}
}

// parseChangeEvent converts MongoDB change document to our ChangeEvent
func (d *Database) parseChangeEvent(changeDoc bson.M) *models.ChangeEvent {
	event := &models.ChangeEvent{
		ClientTimestamp: time.Now(),
	}

	// Extract operation type
	if opType, ok := changeDoc["operationType"].(string); ok {
		event.OperationType = opType
	}

	// Extract namespace (database and collection)
	if ns, ok := changeDoc["ns"].(bson.M); ok {
		if db, ok := ns["db"].(string); ok {
			event.Database = db
		}
		if coll, ok := ns["coll"].(string); ok {
			event.Collection = coll
		}
	}

	// Extract document key
	if docKey, ok := changeDoc["documentKey"].(bson.M); ok {
		event.DocumentKey = docKey
	}

	// Extract full document (for insert/replace operations)
	if fullDoc, ok := changeDoc["fullDocument"].(bson.M); ok {
		event.FullDocument = fullDoc
	}

	// Extract update description
	if updateDesc, ok := changeDoc["updateDescription"].(bson.M); ok {
		if updatedFields, ok := updateDesc["updatedFields"].(bson.M); ok {
			event.UpdatedFields = updatedFields
		}
		if removedFields, ok := updateDesc["removedFields"].(bson.A); ok {
			for _, field := range removedFields {
				if fieldStr, ok := field.(string); ok {
					event.RemovedFields = append(event.RemovedFields, fieldStr)
				}
			}
		}
	}

	// Extract cluster time
	if clusterTime, ok := changeDoc["clusterTime"]; ok {
		if ts, ok := clusterTime.(primitive.Timestamp); ok {
			event.Timestamp = ts
		}
	}

	// Generate ID for the event
	if id, ok := changeDoc["_id"]; ok {
		event.ID = fmt.Sprintf("%v", id)
	}

	return event
}

// GetChanges returns the channel for receiving change events
func (d *Database) GetChanges() <-chan *models.ChangeEvent {
	return d.changesCh
}

// Insert inserts a document into a collection
func (d *Database) Insert(collection string, document interface{}) (*mongo.InsertOneResult, error) {
	coll := d.db.Collection(collection)
	return coll.InsertOne(d.ctx, document)
}

// Update updates documents in a collection
func (d *Database) Update(collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	coll := d.db.Collection(collection)
	return coll.UpdateMany(d.ctx, filter, update)
}

// Delete deletes documents from a collection
func (d *Database) Delete(collection string, filter interface{}) (*mongo.DeleteResult, error) {
	coll := d.db.Collection(collection)
	return coll.DeleteMany(d.ctx, filter)
}

// Find finds documents in a collection
func (d *Database) Find(collection string, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	coll := d.db.Collection(collection)
	return coll.Find(d.ctx, filter, opts...)
}

// Close closes the database connection
func (d *Database) Close() error {
	d.cancel()
	close(d.changesCh)
	return d.client.Disconnect(d.ctx)
}

// GetConnectionURI returns the MongoDB connection URI
func (d *Database) GetConnectionURI() string {
	return d.connectionURI
}

// StreamSnapshot streams existing documents from a collection in batches
func (d *Database) StreamSnapshot(collection string, snapOpts *models.SnapshotOptions, callback func([]map[string]interface{}, int, int, error)) {
	if snapOpts == nil {
		callback(nil, 0, 0, fmt.Errorf("snapshot options are required"))
		return
	}

	// Set default values
	batchSize := snapOpts.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	limit := snapOpts.SnapshotLimit
	if limit <= 0 {
		limit = 10000 // Default max limit for safety
	}

	// Build the filter
	filter := bson.M{}
	if snapOpts.SnapshotFilter != nil {
		filter = snapOpts.SnapshotFilter
	}

	// Build find options
	findOpts := options.Find().SetLimit(int64(limit))
	if snapOpts.SnapshotSort != nil {
		findOpts.SetSort(snapOpts.SnapshotSort)
	}

	// Get total count (with filter applied)
	coll := d.db.Collection(collection)
	totalCount, err := coll.CountDocuments(d.ctx, filter)
	if err != nil {
		callback(nil, 0, 0, fmt.Errorf("failed to count documents: %w", err))
		return
	}

	// Don't exceed the limit
	actualTotal := int(totalCount)
	if actualTotal > limit {
		actualTotal = limit
	}

	d.logger.WithFields(logrus.Fields{
		"collection":    collection,
		"total":         actualTotal,
		"batch_size":    batchSize,
		"filter_fields": len(filter),
	}).Info("Starting snapshot stream")

	// Stream documents in batches
	skip := 0
	batchNum := 1
	remaining := actualTotal

	for remaining > 0 {
		currentBatchSize := batchSize
		if remaining < batchSize {
			currentBatchSize = remaining
		}

		// Fetch batch
		batchOpts := options.Find().
			SetSkip(int64(skip)).
			SetLimit(int64(currentBatchSize))

		if snapOpts.SnapshotSort != nil {
			batchOpts.SetSort(snapOpts.SnapshotSort)
		}

		cursor, err := coll.Find(d.ctx, filter, batchOpts)
		if err != nil {
			callback(nil, batchNum, remaining, fmt.Errorf("failed to find documents: %w", err))
			return
		}

		// Decode batch
		var batch []map[string]interface{}
		for cursor.Next(d.ctx) {
			var doc map[string]interface{}
			if err := cursor.Decode(&doc); err != nil {
				cursor.Close(d.ctx)
				callback(nil, batchNum, remaining, fmt.Errorf("failed to decode document: %w", err))
				return
			}
			batch = append(batch, doc)
		}
		cursor.Close(d.ctx)

		// Send batch to callback
		callback(batch, batchNum, remaining-len(batch), nil)

		// Update counters
		skip += len(batch)
		remaining -= len(batch)
		batchNum++

		// Break if we got fewer documents than expected (end of collection)
		if len(batch) < currentBatchSize {
			break
		}
	}

	d.logger.WithFields(logrus.Fields{
		"collection": collection,
		"batches":    batchNum - 1,
		"documents":  actualTotal - remaining,
	}).Info("Snapshot stream completed")
}
