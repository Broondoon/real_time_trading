package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDatabase encapsulates the MongoDB client and database.
type MongoDatabase struct {
	client    *mongo.Client
	database  *mongo.Database
	connected bool
}

// Connect establishes a connection to MongoDB using the provided URI and database name.
// It uses a retry mechanism similar to the Postgres template.
func (d *MongoDatabase) Connect(uri string, dbName string) {
	retriesStr := os.Getenv("HEALTHCHECK_RETRIES")
	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		retries = 10
	}
	intervalStr := os.Getenv("HEALTHCHECK_INTERVAL")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 1
	}

	var client *mongo.Client
	var connectErr error
	ctx := context.Background()

	for i := 0; i < retries; i++ {
		clientOptions := options.Client().ApplyURI(uri)
		client, connectErr = mongo.Connect(ctx, clientOptions)
		if connectErr == nil {
			// Check if the connection is alive.
			if err := client.Ping(ctx, nil); err == nil {
				d.client = client
				d.database = client.Database(dbName)
				d.connected = true
				return
			}
		}
		log.Printf("MongoDB not ready yet, retrying... (%d/%d)", i+1, retries)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	log.Fatal("MongoDB connection failed after multiple attempts: ", connectErr)
}

// Disconnect cleanly disconnects the MongoDB client.
func (d *MongoDatabase) Disconnect(ctx context.Context) {
	if d.connected {
		if err := d.client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect from MongoDB: ", err)
		}
		d.connected = false
	}
}

// GetDatabase returns the connected database instance.
func (d *MongoDatabase) GetDatabase() *mongo.Database {
	if !d.connected {
		log.Fatal("MongoDB is not connected")
	}
	return d.database
}

// MongoEntityData provides CRUD operations for entities of type T.
// It mimics the PostgresEntityData interface but is adapted to MongoDB operations.
type MongoEntityData[T any] struct {
	db             *MongoDatabase
	collectionName string
}

// NewMongoEntityData creates a new instance of MongoEntityData for a given collection.
func NewMongoEntityData[T any](db *MongoDatabase, collectionName string) *MongoEntityData[T] {
	return &MongoEntityData[T]{
		db:             db,
		collectionName: collectionName,
	}
}

// GetCollection returns the MongoDB collection for the entity.
func (m *MongoEntityData[T]) GetCollection() *mongo.Collection {
	return m.db.GetDatabase().Collection(m.collectionName)
}

// Create inserts a new entity into the collection.
func (m *MongoEntityData[T]) Create(ctx context.Context, entity T) error {
	collection := m.GetCollection()
	_, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("error inserting entity: %w", err)
	}
	return nil
}

// GetByID retrieves an entity by its ID.
// It assumes that the entity has a field named "_id" and that id is of type string.
func (m *MongoEntityData[T]) GetByID(ctx context.Context, id string) (T, error) {
	var result T
	if id == "" {
		return result, fmt.Errorf("ID is empty")
	}
	collection := m.GetCollection()
	filter := bson.M{"_id": id}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("error finding entity by ID %s: %w", id, err)
	}
	return result, nil
}

// GetAll retrieves all entities from the collection.
func (m *MongoEntityData[T]) GetAll(ctx context.Context) ([]T, error) {
	var results []T
	collection := m.GetCollection()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error retrieving all entities: %w", err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("error decoding entity: %w", err)
		}
		results = append(results, item)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	return results, nil
}

// Update performs an update on an entity identified by its ID.
// The update parameter should be a valid BSON document (e.g., bson.M{"$set": bson.M{"field": newValue}}).
func (m *MongoEntityData[T]) Update(ctx context.Context, id string, update interface{}) error {
	if id == "" {
		return fmt.Errorf("ID is empty")
	}
	collection := m.GetCollection()
	filter := bson.M{"_id": id}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating entity with ID %s: %w", id, err)
	}
	return nil
}

// Delete removes an entity by its ID.
func (m *MongoEntityData[T]) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID is empty")
	}
	collection := m.GetCollection()
	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting entity with ID %s: %w", id, err)
	}
	return nil
}
