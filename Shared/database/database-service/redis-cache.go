package database

import (
	"Shared/entities/entity"
	"Shared/objects"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Caching code below
type CachedEntityData[T entity.EntityInterface, TDatabase any] struct {
	underlying  EntityDataInterface[T, TDatabase]
	redisClient *redis.Client
	defaultTTL  time.Duration
}

type NewCachedEntityDataParams[T entity.EntityInterface, TDatabase any] struct {
	RedisAddr  string
	Password   string
	DefaultTTL time.Duration
	EntityData EntityDataInterface[T, TDatabase]
}

// NewCachedEntityData initializes a new CachedEntityData instance with Redis client and default TTL.
func NewCachedEntityData[T entity.EntityInterface, TDatabase any](params *NewCachedEntityDataParams[T, TDatabase]) *CachedEntityData[T, TDatabase] {
	log.Printf("[Cache Init] Creating Redis client with Addr=%s, TTL=%s", params.RedisAddr, params.DefaultTTL)
	rdb := redis.NewClient(&redis.Options{
		Addr:     params.RedisAddr,
		Password: params.Password,
		DB:       0,
	})
	return &CachedEntityData[T, TDatabase]{
		underlying:  params.EntityData,
		redisClient: rdb,
		defaultTTL:  params.DefaultTTL,
	}
}

// redisKey generates a Redis key for a given entity ID.
func (c *CachedEntityData[T, TDatabase]) redisKey(id string) string {
	key := "entity:" + id
	log.Printf("[Cache Key] Generated key: %s", key)
	return key
}

// GetDBUrl delegates the call to the underlying database to get the database URL.
func (c *CachedEntityData[T, TDatabase]) GetDBUrl() string {
	return c.underlying.GetDBUrl()
}

// IsConnected checks if the underlying database is connected.
func (c *CachedEntityData[T, TDatabase]) IsConnected() bool {
	return c.underlying.IsConnected()
}

// SetConnected sets the connection status of the underlying database.
func (c *CachedEntityData[T, TDatabase]) SetConnected(connected bool) {
	c.underlying.SetConnected(connected)
}

// Connect establishes a connection to the underlying database.
func (c *CachedEntityData[T, TDatabase]) Connect() {
	log.Println("[Cache] Connect called")
	c.underlying.Connect()
}

// Disconnect closes the connection to the underlying database.
func (c *CachedEntityData[T, TDatabase]) Disconnect() {
	log.Println("[Cache] Disconnect called")
	c.underlying.Disconnect()
}

// GetDatabaseSession retrieves the current database session from the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetDatabaseSession() TDatabase {
	return c.underlying.GetDatabaseSession()
}

// GetNewDatabaseSession creates a new database session from the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetNewDatabaseSession() TDatabase {
	return c.underlying.GetNewDatabaseSession()
}

// GetByID retrieves an entity by its ID, first checking the cache and falling back to the database if necessary.
func (c *CachedEntityData[T, TDatabase]) GetByID(id string, shardKey string) (T, error) {
	log.Printf("[Cache] GetByID: Looking for entity with ID %s", id)
	ctx := context.Background()
	var zero T
	key := c.redisKey(id)
	log.Printf("[Cache] GetByID: 🔍 Looking for entity in cache [Key: %s]", key)

	// Step 1: Check cache
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		log.Printf("[Cache] GetByID: ✅ Cache hit for key [%s]", key)

		var cachedEntity T
		if err = json.Unmarshal([]byte(data), &cachedEntity); err == nil {
			log.Printf("[Cache] GetByID: 🔄 Successfully unmarshaled cached entity [ID: %s]: %+v", id, cachedEntity)
			return cachedEntity, nil
		}
		log.Printf("[Cache] GetByID: ❌ Error unmarshaling cached data [Key: %s]: %v", key, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetByID: ❌ Redis GET error [Key: %s]: %v", key, err)
	} else {
		log.Printf("[Cache] GetByID: ❌ Cache miss [Key: %s]", key)
	}

	// Step 2: Fetch from database
	log.Printf("[Cache] GetByID: 📡 Querying database for ID: %s", id)
	dbEntity, err := c.underlying.GetByID(id, shardKey)
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Database error for ID [%s]: %v", id, err)
		return zero, err
	}
	log.Printf("[Cache] GetByID: ✅ Successfully retrieved entity from database [ID: %s]: %+v", id, dbEntity)

	// Step 3: Store in cache
	jsonBytes, err := json.MarshalIndent(dbEntity, "", "  ")
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Error marshaling entity for cache [ID: %s]: %v", id, err)
	} else {
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] GetByID: ❌ Error caching entity [Key: %s]: %v", key, err)
		} else {
			log.Printf("[Cache] GetByID: ✅ Cached entity in Redis [Key: %s]:\n%s", key, string(jsonBytes))
		}
	}

	return dbEntity, nil
}

// GetByIDs retrieves multiple entities by their IDs, using the cache and database as needed.
func (c *CachedEntityData[T, TDatabase]) GetByIDs(ids []string, shardKeys *[]string) (*[]T, map[string]error) {
	log.Printf("[Cache] GetByIDs: Looking up multiple IDs: %v", ids)
	errorList := make(map[string]error)
	ctx := context.Background()
	entityMap := make(map[string]T)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = c.redisKey(id)
	}
	log.Printf("[Cache] GetByIDs: 🔍 Looking up multiple keys: %v", keys)

	// Step 1: Fetch from cache
	results, err := c.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("[Cache] GetByIDs: ❌ Redis MGET error: %v", err)
	}
	var missingIds []string
	for i, res := range results {
		id := ids[i]
		if res == nil {
			log.Printf("[Cache] GetByIDs: ❌ Cache miss for ID %s", id)
			missingIds = append(missingIds, id)
			continue
		}
		str, ok := res.(string)
		if !ok {
			log.Printf("[Cache] GetByIDs: ❌ Invalid data type in cache for ID %s", id)
			missingIds = append(missingIds, id)
			continue
		}
		var cachedEntity T
		if err := json.Unmarshal([]byte(str), &cachedEntity); err != nil {
			log.Printf("[Cache] GetByIDs: ❌ Error unmarshaling cached entity for ID %s: %v", id, err)
			missingIds = append(missingIds, id)
		} else {
			log.Printf("[Cache] GetByIDs: ✅ Successfully retrieved from cache [ID: %s]:\n%s", id, str)
			entityMap[id] = cachedEntity
		}
	}

	// Step 2: Fetch missing IDs from database
	if len(missingIds) > 0 {
		log.Printf("[Cache] GetByIDs: 📡 Fetching missing IDs from database: %v", missingIds)
		dbEntities, errorList := c.underlying.GetByIDs(missingIds, shardKeys)

		// Step 3: Cache newly retrieved entities
		for _, entity := range *dbEntities {
			if _, ok := errorList[entity.GetIdString()]; !ok {
				id := entity.GetIdString()
				entityMap[id] = entity
				jsonBytes, err := json.MarshalIndent(entity, "", "  ")
				if err != nil {
					log.Printf("[Cache] GetByIDs: ❌ Error marshaling entity for cache [ID: %s]: %v", id, err)
					continue
				}
				if err := c.redisClient.Set(ctx, c.redisKey(id), jsonBytes, c.defaultTTL).Err(); err != nil {
					log.Printf("[Cache] GetByIDs: ❌ Error caching entity [ID: %s]: %v", id, err)
				} else {
					log.Printf("[Cache] GetByIDs: ✅ Cached entity in Redis [ID: %s]:\n%s", id, string(jsonBytes))
				}
			}
		}
	}

	// Step 4: Aggregate results
	finalEntities := make([]T, 0, len(ids))
	for _, id := range ids {
		if entity, exists := entityMap[id]; exists {
			finalEntities = append(finalEntities, entity)
		} else {
			log.Printf("[Cache] GetByIDs: ❌ No entity found for ID %s", id)
		}
	}
	return &finalEntities, errorList
}

// GetByForeignID retrieves entities by a foreign key, using the cache and database as needed.
func (c *CachedEntityData[T, TDatabase]) GetByForeignID(foreignIDColumn, foreignID string, shardKey string) (*[]T, error) {
	return c.underlying.GetByForeignID(foreignIDColumn, foreignID, shardKey)
}

// GetByForeignIDBulk retrieves entities by multiple foreign keys, delegating to the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string, shardKeys *[]string) (*[]T, map[string]error) {
	return c.underlying.GetByForeignIDBulk(foreignIDColumn, foreignIDs, shardKeys)
}

// GetByFilteredForeignIDBulk retrieves entities by filtered foreign keys, delegating to the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterKey string, filterVal string, shardKeys *[]string) (*[]T, map[string]error) {
	return c.underlying.GetByFilteredForeignIDBulk(foreignIDKey, foreignIDs, filterKey, filterVal, shardKeys)
}

// Create adds a new entity to the database and updates the cache.
func (c *CachedEntityData[T, TDatabase]) Create(entity T) error {
	log.Printf("[Cache] Create: Creating entity with ID %s", entity.GetIdString())

	if err := c.underlying.Create(entity); err != nil {
		return err
	}

	ctx := context.Background()
	entityKey := c.redisKey(entity.GetIdString())

	// Step 1: Cache the individual entity
	if jsonBytes, err := json.Marshal(entity); err == nil {
		c.redisClient.Set(ctx, entityKey, jsonBytes, c.defaultTTL)
		log.Printf("[Cache] Create: Cached entity with key %s", entityKey)
	}

	// Step 2: Update the all_entities cache
	allKey := "all_entities"
	var currentEntities []T

	data, err := c.redisClient.Get(ctx, allKey).Result()
	if err == nil {
		// Cache hit — try to unmarshal
		if err := json.Unmarshal([]byte(data), &currentEntities); err != nil {
			log.Printf("[Cache] Create: ❌ Failed to unmarshal existing all_entities cache: %v", err)
			currentEntities = []T{} // fallback to empty
		}
	} else if err != redis.Nil {
		log.Printf("[Cache] Create: ❌ Redis error on all_entities: %v", err)
	}

	// Append the new entity
	currentEntities = append(currentEntities, entity)

	// Re-cache the updated list
	if jsonBytes, err := json.Marshal(currentEntities); err == nil {
		if err := c.redisClient.Set(ctx, allKey, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] Create: ❌ Failed to update all_entities cache: %v", err)
		} else {
			log.Printf("[Cache] Create: ✅ Updated all_entities cache")
		}
	}

	return nil
}

// CreateBulk adds multiple entities to the database and updates the cache.
func (c *CachedEntityData[T, TDatabase]) CreateBulk(entities *[]T) map[string]error {
	log.Printf("[Cache] CreateBulk: Creating %d entities", len(*entities))
	err := c.underlying.CreateBulk(entities)
	ctx := context.Background()

	// Cache each successful entity
	for _, entity := range *entities {
		if _, ok := err[entity.GetUniquePairing().String()]; !ok {
			jsonBytes, err := json.Marshal(entity)
			if err != nil {
				log.Printf("[Cache] CreateBulk: Error marshaling entity with id %s: %v", entity.GetId(), err)
				continue
			}
			key := c.redisKey(entity.GetIdString())
			if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
				log.Printf("[Cache] CreateBulk: Error caching entity with key %s: %v", key, err)
			} else {
				log.Printf("[Cache] CreateBulk: Successfully cached entity with key %s", key)
			}
		}
	}

	// Step 2: Update all_entities
	allKey := "all_entities"
	var currentEntities []T

	data, getErr := c.redisClient.Get(ctx, allKey).Result()
	if getErr == nil {
		if err := json.Unmarshal([]byte(data), &currentEntities); err != nil {
			log.Printf("[Cache] CreateBulk: ❌ Failed to unmarshal existing all_entities cache: %v", err)
			currentEntities = []T{} // fallback to empty
		}
	} else if getErr != redis.Nil {
		log.Printf("[Cache] CreateBulk: ❌ Redis error on all_entities: %v", getErr)
	}

	// Append new entities to current list
	currentEntities = append(currentEntities, *entities...)

	// Save updated list
	if jsonBytes, err := json.Marshal(currentEntities); err == nil {
		if err := c.redisClient.Set(ctx, allKey, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] CreateBulk: ❌ Failed to update all_entities cache: %v", err)
		} else {
			log.Printf("[Cache] CreateBulk: ✅ Updated all_entities cache")
		}
	}

	return err
}

// GetAll retrieves all entities, using the cache and falling back to the database if necessary.
func (c *CachedEntityData[T, TDatabase]) GetAll() (*[]T, error) {
	return c.underlying.GetAll()
}

// Update modifies existing entities in the database and updates the cache.
func (c *CachedEntityData[T, TDatabase]) Update(updates []*entity.EntityUpdateData) map[string]error {
	log.Printf("[Cache] Update: Updating entities with IDs: %v", updates)
	err := c.underlying.Update(updates)

	uniqueIds := make(map[string]struct{})
	ids := make([]string, 0)
	for _, update := range updates {
		if _, ok := uniqueIds[update.ID.String()]; ok {
			continue
		}
		uniqueIds[update.ID.String()] = struct{}{}
		if _, ok := err[update.ID.String()]; ok {
			continue
		}
		ids = append(ids, update.ID.String())
	}

	go func() {
		//iterate through ids
		ctx := context.Background()
		//entities, err2 := c.underlying.GetByIDs(ids)
		//for loop that iterates through all entities
		//add each entity to a hashmap keyed to its id string
		for _, id := range ids { //replace entities with id's
			/* if _, ok := err2[ent.GetIdString()]; !ok { //anywhere it says ent.IdString it needs to be any value were iterating through
			key := c.redisKey(ent.GetIdString())
			_ = c.redisClient.Del(ctx, key)

			jsonBytes, err := json.MarshalIndent(ent, "", "  ")
			if err == nil {
				_ = c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL)
			} */
			key := c.redisKey(id)
			_ = c.redisClient.Del(ctx, key)
		}

		// ❗ Invalidate the all_entities cache

	}()

	return err
}

// Delete removes an entity by its ID from the database and invalidates the cache.
func (c *CachedEntityData[T, TDatabase]) Delete(id string, shardKey string) error {
	log.Printf("[Cache] Delete: Deleting entity with ID %s", id)
	if err := c.underlying.Delete(id, shardKey); err != nil {
		log.Printf("[Cache] Delete: Underlying DB delete failed for id %s: %v", id, err)
		return err
	}

	ctx := context.Background()
	key := c.redisKey(id)
	_ = c.redisClient.Del(ctx, key)

	// ❗ Invalidate the all_entities cache
	_ = c.redisClient.Del(ctx, "all_entities")

	return nil
}

// DeleteBulk removes multiple entities by their IDs from the database and invalidates the cache.
func (c *CachedEntityData[T, TDatabase]) DeleteBulk(ids []string, shardKeys *[]string) map[string]error {
	log.Printf("[Cache] DeleteBulk: Deleting entities with IDs: %v", ids)
	errorList := c.underlying.DeleteBulk(ids, shardKeys)

	ctx := context.Background()
	for _, id := range ids {
		if _, ok := errorList[id]; !ok {
			key := c.redisKey(id)
			_ = c.redisClient.Del(ctx, key)
		}
	}

	// ❗ Invalidate the all_entities cache
	_ = c.redisClient.Del(ctx, "all_entities")

	return errorList
}

/* func getCacheKey(ids objects.Pair) string {
	return ids.String()
} */

// GetByPairedID retrieves an entity by a pair of IDs, delegating to the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetByPairedID(idColumn1 string, idColumn2 string, ids objects.Pair) (T, error) {
	return c.underlying.GetByPairedID(idColumn1, idColumn2, ids)
	/* id := getCacheKey(ids)
	log.Printf("[Cache] GetByID: Looking for entity with ID %s", id)
	ctx := context.Background()
	var zero T
	key := c.redisKey(id)
	log.Printf("[Cache] GetByID: 🔍 Looking for entity in cache [Key: %s]", key)

	// Step 1: Check cache
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		log.Printf("[Cache] GetByID: ✅ Cache hit for key [%s]", key)

		var cachedEntity T
		if err = json.Unmarshal([]byte(data), &cachedEntity); err == nil {
			log.Printf("[Cache] GetByID: 🔄 Successfully unmarshaled cached entity [ID: %s]: %+v", id, cachedEntity)
			return cachedEntity, nil
		}
		log.Printf("[Cache] GetByID: ❌ Error unmarshaling cached data [Key: %s]: %v", key, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetByID: ❌ Redis GET error [Key: %s]: %v", key, err)
	} else {
		log.Printf("[Cache] GetByID: ❌ Cache miss [Key: %s]", key)
	}

	// Step 2: Fetch from database
	log.Printf("[Cache] GetByID: 📡 Querying database for ID: %s", id)
	dbEntity, err := c.underlying.GetByID(id)
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Database error for ID [%s]: %v", id, err)
		return zero, err
	}
	log.Printf("[Cache] GetByID: ✅ Successfully retrieved entity from database [ID: %s]: %+v", id, dbEntity)

	// Step 3: Store in cache
	jsonBytes, err := json.MarshalIndent(dbEntity, "", "  ")
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Error marshaling entity for cache [ID: %s]: %v", id, err)
	} else {
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] GetByID: ❌ Error caching entity [Key: %s]: %v", key, err)
		} else {
			log.Printf("[Cache] GetByID: ✅ Cached entity in Redis [Key: %s]:\n%s", key, string(jsonBytes))
		}
	}

	return dbEntity, nil */
}

// GetByPairedIDBulk retrieves multiple entities by pairs of IDs, delegating to the underlying database.
func (c *CachedEntityData[T, TDatabase]) GetByPairedIDBulk(idColumn1 string, idColumn2 string, ids *[]objects.Pair) (*[]T, map[string]error) {
	return c.underlying.GetByPairedIDBulk(idColumn1, idColumn2, ids)
}
