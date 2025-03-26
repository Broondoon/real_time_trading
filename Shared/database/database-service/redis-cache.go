package database

import (
	"Shared/entities/entity"
	"context"
	"encoding/json"
	"fmt"
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

func (c *CachedEntityData[T, TDatabase]) redisKey(id string) string {
	key := "entity:" + id
	log.Printf("[Cache Key] Generated key: %s", key)
	return key
}

// Delegate BaseDatabaseInterface methods.
func (c *CachedEntityData[T, TDatabase]) GetDBUrl() string {
	return c.underlying.GetDBUrl()
}

func (c *CachedEntityData[T, TDatabase]) IsConnected() bool {
	return c.underlying.IsConnected()
}

func (c *CachedEntityData[T, TDatabase]) SetConnected(connected bool) {
	c.underlying.SetConnected(connected)
}

// Delegate DatabaseInterface methods.
func (c *CachedEntityData[T, TDatabase]) Connect() {
	log.Println("[Cache] Connect called")
	c.underlying.Connect()
}

func (c *CachedEntityData[T, TDatabase]) Disconnect() {
	log.Println("[Cache] Disconnect called")
	c.underlying.Disconnect()
}

// // Delegate PostGresDatabaseInterface methods.
func (c *CachedEntityData[T, TDatabase]) GetDatabaseSession() TDatabase {
	return c.underlying.GetDatabaseSession()
}

func (c *CachedEntityData[T, TDatabase]) GetNewDatabaseSession() TDatabase {
	return c.underlying.GetNewDatabaseSession()
}

func (c *CachedEntityData[T, TDatabase]) GetByID(id string) (T, error) {
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

	return dbEntity, nil
}

func (c *CachedEntityData[T, TDatabase]) GetByIDs(ids []string) (*[]T, map[string]error) {
	log.Printf("[Cache] GetByIDs: Looking up multiple IDs: %v", ids)
	errors := make(map[string]error)
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
		dbEntities, errors := c.underlying.GetByIDs(missingIds)

		// Step 3: Cache newly retrieved entities
		for _, entity := range *dbEntities {
			if _, ok := errors[entity.GetIdString()]; !ok {
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
	return &finalEntities, errors
}

func (c *CachedEntityData[T, TDatabase]) GetByForeignID(foreignIDColumn, foreignID string) (*[]T, error) {
	log.Printf("[Cache] GetByForeignID: Looking for entities with foreign ID %s in column %s", foreignID, foreignIDColumn)
	ctx := context.Background()
	cacheKey := fmt.Sprintf("foreign:%s:%s", foreignIDColumn, foreignID)

	// Attempt to fetch from cache
	if data, err := c.redisClient.Get(ctx, cacheKey).Result(); err == nil {
		var cachedEntities []T
		if json.Unmarshal([]byte(data), &cachedEntities) == nil {
			log.Printf("[Cache] GetByForeignID: Cache hit for key %s", cacheKey)
			return &cachedEntities, nil
		}
	} else if err != redis.Nil {
		log.Printf("[Cache] Redis error for key %s: %v", cacheKey, err)
	} else {
		log.Printf("[Cache] GetByForeignID: Cache miss for key %s", cacheKey)
	}

	// Fetch from DB if cache miss
	dbEntities, err := c.underlying.GetByForeignID(foreignIDColumn, foreignID)
	if err != nil {
		return nil, err
	}

	// Cache non-empty results
	if len(*dbEntities) > 0 {
		if jsonBytes, err := json.Marshal(dbEntities); err == nil {
			c.redisClient.Set(ctx, cacheKey, jsonBytes, c.defaultTTL)
			log.Printf("[Cache] GetByForeignID: Cached DB result for key %s", cacheKey)
		}
	}

	return dbEntities, nil
}

func (c *CachedEntityData[T, TDatabase]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, map[string]error) {
	return c.underlying.GetByForeignIDBulk(foreignIDColumn, foreignIDs)
}

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

func (c *CachedEntityData[T, TDatabase]) GetAll() (*[]T, error) {
	log.Println("[Cache] GetAll: Retrieving all entities")
	ctx := context.Background()
	cacheKey := "all_entities"
	log.Printf("[Cache] GetAll: Looking for key %s", cacheKey)

	var zero []T
	data, err := c.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("[Cache] GetAll: Found cached data for key %s: %s", cacheKey, data)
		var cachedEntities []T
		if err = json.Unmarshal([]byte(data), &cachedEntities); err == nil {
			log.Printf("[Cache] GetAll: Successfully unmarshaled cached data for key %s", cacheKey)
			return &cachedEntities, nil
		}
		log.Printf("[Cache] GetAll: Error unmarshaling cached data for key %s: %v", cacheKey, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetAll: Redis GET error for key %s: %v", cacheKey, err)
	} else {
		log.Printf("[Cache] GetAll: Cache miss for key %s", cacheKey)
	}

	// Fallback to the underlying database if cache miss or error.
	dbEntities, err := c.underlying.GetAll()
	if err != nil {
		log.Printf("[Cache] GetAll: DB error: %v", err)
		return &zero, err
	}

	// Cache the result from the database.
	if jsonBytes, err := json.Marshal(dbEntities); err == nil {
		if err := c.redisClient.Set(ctx, cacheKey, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] GetAll: Error caching DB result for key %s: %v", cacheKey, err)
		} else {
			log.Printf("[Cache] GetAll: Cached DB result for key %s", cacheKey)
		}
	} else {
		log.Printf("[Cache] GetAll: Error marshaling DB result for key %s: %v", cacheKey, err)
	}
	return dbEntities, nil
}

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
		ctx := context.Background()
		entities, err2 := c.underlying.GetByIDs(ids)
		for _, ent := range *entities {
			if _, ok := err2[ent.GetIdString()]; !ok {
				key := c.redisKey(ent.GetIdString())
				_ = c.redisClient.Del(ctx, key)

				jsonBytes, err := json.MarshalIndent(ent, "", "  ")
				if err == nil {
					_ = c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL)
				}
			}
		}

		// ❗ Invalidate the all_entities cache
		_ = c.redisClient.Del(ctx, "all_entities")
	}()

	return err
}

func (c *CachedEntityData[T, TDatabase]) Delete(id string) error {
	log.Printf("[Cache] Delete: Deleting entity with ID %s", id)
	if err := c.underlying.Delete(id); err != nil {
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

func (c *CachedEntityData[T, TDatabase]) DeleteBulk(ids []string) map[string]error {
	log.Printf("[Cache] DeleteBulk: Deleting entities with IDs: %v", ids)
	errors := c.underlying.DeleteBulk(ids)

	ctx := context.Background()
	for _, id := range ids {
		if _, ok := errors[id]; !ok {
			key := c.redisKey(id)
			_ = c.redisClient.Del(ctx, key)
		}
	}

	// ❗ Invalidate the all_entities cache
	_ = c.redisClient.Del(ctx, "all_entities")

	return errors
}
