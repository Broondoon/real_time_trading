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
	//	log.Printf("[Cache Init] Creating Redis client with Addr=%s, TTL=%s", params.RedisAddr, params.DefaultTTL)
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
	//	log.Printf("[Cache Key] Generated key: %s", key)
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
	//	log.Println("[Cache] Connect called")
	c.underlying.Connect()
}

func (c *CachedEntityData[T, TDatabase]) Disconnect() {
	//	log.Println("[Cache] Disconnect called")
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
	ctx := context.Background()
	var zero T
	key := c.redisKey(id)
	//log.Printf("[Cache] GetByID: 🔍 Looking for entity in cache [Key: %s]", key)

	// Step 1: Check cache
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		//log.Printf("[Cache] GetByID: ✅ Cache hit for key [%s]", key)

		var cachedEntity T
		if err = json.Unmarshal([]byte(data), &cachedEntity); err == nil {
			//log.Printf("[Cache] GetByID: 🔄 Successfully unmarshaled cached entity [ID: %s]: %+v", id, cachedEntity)
			return cachedEntity, nil
		}
		log.Printf("[Cache] GetByID: ❌ Error unmarshaling cached data [Key: %s]: %v", key, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetByID: ❌ Redis GET error [Key: %s]: %v", key, err)
	} else {
		//log.Printf("[Cache] GetByID: ❌ Cache miss [Key: %s]", key)
	}

	// Step 2: Fetch from database
	//log.Printf("[Cache] GetByID: 📡 Querying database for ID: %s", id)
	dbEntity, err := c.underlying.GetByID(id)
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Database error for ID [%s]: %v", id, err)
		return zero, err
	}
	//log.Printf("[Cache] GetByID: ✅ Successfully retrieved entity from database [ID: %s]: %+v", id, dbEntity)

	// Step 3: Store in cache
	jsonBytes, err := json.MarshalIndent(dbEntity, "", "  ")
	if err != nil {
		log.Printf("[Cache] GetByID: ❌ Error marshaling entity for cache [ID: %s]: %v", id, err)
	} else {
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] GetByID: ❌ Error caching entity [Key: %s]: %v", key, err)
		} else {
			//log.Printf("[Cache] GetByID: ✅ Cached entity in Redis [Key: %s]:\n%s", key, string(jsonBytes))
		}
	}

	return dbEntity, nil
}

func (c *CachedEntityData[T, TDatabase]) GetByIDs(ids []string) (*[]T, map[string]error) {
	errorList := make(map[string]error)
	ctx := context.Background()
	entityMap := make(map[string]T)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = c.redisKey(id)
	}
	//log.Printf("[Cache] GetByIDs: 🔍 Looking up multiple keys: %v", keys)

	// Step 1: Fetch from cache
	results, err := c.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("[Cache] GetByIDs: ❌ Redis MGET error: %v", err)
	}
	var missingIds []string
	for i, res := range results {
		id := ids[i]
		if res == nil {
			//log.Printf("[Cache] GetByIDs: ❌ Cache miss for ID %s", id)
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
			//log.Printf("[Cache] GetByIDs: ✅ Successfully retrieved from cache [ID: %s]:\n%s", id, str)
			entityMap[id] = cachedEntity
		}
	}

	// Step 2: Fetch missing IDs from database
	if len(missingIds) > 0 {
		//log.Printf("[Cache] GetByIDs: 📡 Fetching missing IDs from database: %v", missingIds)
		dbEntities, errorList := c.underlying.GetByIDs(missingIds)

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
					//	log.Printf("[Cache] GetByIDs: ✅ Cached entity in Redis [ID: %s]:\n%s", id, string(jsonBytes))
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

/* func (c *CachedEntityData[T]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("foreign:%s:%s", foreignIDColumn, foreignID)
	log.Printf("[Cache] GetByForeignID: Looking up key %s", cacheKey)
	var zero []T
	data, err := c.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("[Cache] GetByForeignID: Found data for key %s: %s", cacheKey, data)
		var cachedEntities []T
		if err = json.Unmarshal([]byte(data), &cachedEntities); err == nil {
			log.Printf("[Cache] GetByForeignID: Successfully unmarshaled data for key %s", cacheKey)
			return &cachedEntities, nil
		}
		log.Printf("[Cache] GetByForeignID: Error unmarshaling cache for key %s: %v", cacheKey, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetByForeignID: Redis GET error for key %s: %v", cacheKey, err)
	} else {
		log.Printf("[Cache] GetByForeignID: Cache miss for key %s", cacheKey)
	}

	dbEntities, err := c.underlying.GetByForeignID(foreignIDColumn, foreignID)
	if err != nil {
		log.Printf("[Cache] GetByForeignID: DB error for foreign id %s: %v", foreignID, err)
		return &zero, err
	}

	if len(*dbEntities) > 0 {
		if jsonBytes, err := json.Marshal(dbEntities); err == nil {
			if err := c.redisClient.Set(ctx, cacheKey, jsonBytes, c.defaultTTL).Err(); err == nil {
				log.Printf("[Cache] GetByForeignID: Cached DB result for key %s", cacheKey)
			} else {
				log.Printf("[Cache] GetByForeignID: Error setting cache for key %s: %v", cacheKey, err)
			}
		} else {
			log.Printf("[Cache] GetByForeignID: Error marshaling DB result for key %s: %v", cacheKey, err)
		}
		// not caching the db results if the db result is empty
	} else {
		log.Printf("[Cache] GetByForeignID: DB result is empty; not caching for key %s", cacheKey)
	}
	return dbEntities, nil
} */

func (c *CachedEntityData[T, TDatabase]) GetByForeignID(foreignIDColumn, foreignID string) (*[]T, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("foreign:%s:%s", foreignIDColumn, foreignID)

	// Attempt to fetch from cache
	if data, err := c.redisClient.Get(ctx, cacheKey).Result(); err == nil {
		var cachedEntities []T
		if json.Unmarshal([]byte(data), &cachedEntities) == nil {
			//	log.Printf("[Cache] GetByForeignID: Cache hit for key %s", cacheKey)
			return &cachedEntities, nil
		}
	} else if err != redis.Nil {
		log.Printf("[Cache] Redis error for key %s: %v", cacheKey, err)
	} else {
		//	log.Printf("[Cache] GetByForeignID: Cache miss for key %s", cacheKey)
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
			//	log.Printf("[Cache] GetByForeignID: Cached DB result for key %s", cacheKey)
		}
	}

	return dbEntities, nil
}

func (c *CachedEntityData[T, TDatabase]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, map[string]error) {
	return c.underlying.GetByForeignIDBulk(foreignIDColumn, foreignIDs)
}

func (c *CachedEntityData[T, TDatabase]) GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterKey string, filterVal string) (*[]T, map[string]error) {
	return c.underlying.GetByFilteredForeignIDBulk(foreignIDKey, foreignIDs, filterKey, filterVal)
}

func (c *CachedEntityData[T, TDatabase]) Create(entity T) error {
	if err := c.underlying.Create(entity); err != nil {
		return err
	}

	ctx := context.Background()
	entityKey := c.redisKey(entity.GetIdString())

	// Cache entity after DB insertion
	if jsonBytes, err := json.Marshal(entity); err == nil {
		c.redisClient.Set(ctx, entityKey, jsonBytes, c.defaultTTL)
		//	log.Printf("[Cache] Create: Cached entity with key %s", entityKey)
	}

	return nil
}

func (c *CachedEntityData[T, TDatabase]) CreateBulk(entities *[]T) map[string]error {
	//log.Printf("[Cache] CreateBulk: Creating %d entities", len(*entities))
	err := c.underlying.CreateBulk(entities)
	ctx := context.Background()
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
				//	log.Printf("[Cache] CreateBulk: Successfully cached entity with key %s", key)
			}
		}
	}
	return err
}

func (c *CachedEntityData[T, TDatabase]) GetAll() (*[]T, error) {
	ctx := context.Background()
	cacheKey := "all_entities"
	//log.Printf("[Cache] GetAll: Looking for key %s", cacheKey)

	var zero []T
	data, err := c.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		//log.Printf("[Cache] GetAll: Found cached data for key %s: %s", cacheKey, data)
		var cachedEntities []T
		if err = json.Unmarshal([]byte(data), &cachedEntities); err == nil {
			//	log.Printf("[Cache] GetAll: Successfully unmarshaled cached data for key %s", cacheKey)
			return &cachedEntities, nil
		}
		log.Printf("[Cache] GetAll: Error unmarshaling cached data for key %s: %v", cacheKey, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetAll: Redis GET error for key %s: %v", cacheKey, err)
	} else {
		//	log.Printf("[Cache] GetAll: Cache miss for key %s", cacheKey)
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
			//	log.Printf("[Cache] GetAll: Cached DB result for key %s", cacheKey)
		}
	} else {
		log.Printf("[Cache] GetAll: Error marshaling DB result for key %s: %v", cacheKey, err)
	}
	return dbEntities, nil
}

/* func (c *CachedEntityData[T]) Create(entity T) error {
	log.Printf("[Cache] Create: Attempting to create entity with ID: %s", entity.GetId())

	// Step 1: Insert into the database
	if err := c.underlying.Create(entity); err != nil {
		log.Printf("[Cache] Create: ❌ Failed to create entity in DB [ID: %s]: %v", entity.GetId(), err)
		return err
	}
	log.Printf("[Cache] Create: ✅ Successfully created entity in DB [ID: %s]", entity.GetId())

	// Step 2: Store entity in Redis cache
	ctx := context.Background()
	entityKey := c.redisKey(entity.GetId())

	jsonBytes, err := json.MarshalIndent(entity, "", "  ") // Pretty-print JSON for debugging
	if err != nil {
		log.Printf("[Cache] Create: ❌ Error marshaling entity [ID: %s]: %v", entity.GetId(), err)
		return nil // Not fatal, DB operation was successful
	}

	if err := c.redisClient.Set(ctx, entityKey, jsonBytes, c.defaultTTL).Err(); err != nil {
		log.Printf("[Cache] Create: ❌ Failed to cache entity in Redis [Key: %s]: %v", entityKey, err)
	} else {
		log.Printf("[Cache] Create: ✅ Cached entity in Redis [Key: %s]:\n%s", entityKey, string(jsonBytes))
	}

	// Step 3: Optional - Mark `GetAll()` cache as stale
	cacheKey := "all_entities"
	if err := c.redisClient.Expire(ctx, cacheKey, 10*time.Second).Err(); err != nil {
		log.Printf("[Cache] Create: ⚠️ Failed to mark `all_entities` cache as stale: %v", err)
	} else {
		log.Printf("[Cache] Create: 🔄 Marked `all_entities` cache as stale (TTL: 10s)")
	}

	return nil
} */

func (c *CachedEntityData[T, TDatabase]) Update(updates []*entity.EntityUpdateData) map[string]error {
	//log.Printf("[Cache] Update: Attempting to update entity with ID: %s", entity.GetId())
	err := c.underlying.Update(updates)
	// Step 1: Update entity in the database
	//log.Printf("[Cache] Update: ✅ Successfully updated entity in DB [ID: %s]", entity.GetId())

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
		entities, err2 := c.underlying.GetByIDs(ids)
		for _, ent := range *entities {
			if _, ok := err2[ent.GetIdString()]; !ok {

				// Step 2: Remove outdated cache entry
				ctx := context.Background()
				key := c.redisKey(ent.GetIdString())

				if err := c.redisClient.Del(ctx, key).Err(); err != nil {
					log.Printf("[Cache] Update: ❌ Failed to delete stale cache for [Key: %s]: %v", key, err)
				} else {
					//		log.Printf("[Cache] Update: ✅ Deleted stale cache for [Key: %s]", key)
				}

				// Step 3: Store updated entity in Redis
				jsonBytes, err := json.MarshalIndent(ent, "", "  ")
				if err != nil {
					log.Printf("[Cache] Update: ❌ Error marshaling updated entity [ID: %s]: %v", ent.GetIdString(), err)
					continue
				}

				if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
					log.Printf("[Cache] Update: ❌ Failed to cache updated entity in Redis [Key: %s]: %v", key, err)
				} else {
					//	log.Printf("[Cache] Update: ✅ Cached updated entity in Redis [Key: %s]:\n%s", key, string(jsonBytes))
				}
			}
		}
	}()

	return err
}

func (c *CachedEntityData[T, TDatabase]) Delete(id string) error {
	//	log.Printf("[Cache] Delete: Deleting entity with id %s", id)
	if err := c.underlying.Delete(id); err != nil {
		log.Printf("[Cache] Delete: Underlying DB delete failed for id %s: %v", id, err)
		return err
	}
	ctx := context.Background()
	key := c.redisKey(id)
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[Cache] Delete: Error deleting cache for key %s: %v", key, err)
	} else {
		//	log.Printf("[Cache] Delete: Successfully deleted cache for key %s", key)
	}
	return nil
}

func (c *CachedEntityData[T, TDatabase]) DeleteBulk(ids []string) map[string]error {
	//log.Printf("[Cache] DeleteBulk: Deleting %d entities", len(ids))
	errorList := c.underlying.DeleteBulk(ids)
	ctx := context.Background()
	for _, id := range ids {
		if _, ok := errorList[id]; !ok {
			key := c.redisKey(id)
			if err := c.redisClient.Del(ctx, key).Err(); err != nil {
				log.Printf("[Cache] DeleteBulk: Error deleting cache for key %s: %v", key, err)
			} else {
				//	log.Printf("[Cache] DeleteBulk: Successfully deleted cache for key %s", key)
			}
		}
	}
	return errorList
}
