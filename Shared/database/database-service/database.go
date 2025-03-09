package database

import (
	"Shared/entities/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// BaseDatabase and related types.
type BaseDatabaseInterface interface {
	GetDBUrl() string
	IsConnected() bool
	SetConnected(connected bool)
}

type BaseDatabase struct {
	DatabaseURLEnv string
	Connected      bool
}

type NewBaseDatabaseParams struct {
	DATABASE_URL_ENV string // leave "" for default.
}

func NewBaseDatabase(params *NewBaseDatabaseParams) BaseDatabaseInterface {
	if params.DATABASE_URL_ENV == "" {
		params.DATABASE_URL_ENV = "DATABASE_URL"
	}

	return &BaseDatabase{
		DatabaseURLEnv: params.DATABASE_URL_ENV,
		Connected:      false,
	}
}

func (d *BaseDatabase) GetDBUrl() string {
	dsn := os.Getenv(d.DatabaseURLEnv) // "DATABASE_URL" is an ENV variable set in docker-compose.yml
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
	}
	return dsn
}

func (d *BaseDatabase) IsConnected() bool {
	return d.Connected
}

func (d *BaseDatabase) SetConnected(connected bool) {
	d.Connected = connected
}

// DatabaseInterface and PostGresDatabaseInterface.
type DatabaseInterface interface {
	BaseDatabaseInterface
	Connect()
	Disconnect()
}

type PostGresDatabaseInterface interface {
	DatabaseInterface
	GetDatabaseSession() *gorm.DB
	GetNewDatabaseSession() *gorm.DB
}

type PostGresDatabase struct {
	BaseDatabaseInterface
	database *gorm.DB
}

type NewPostGresDatabaseParams struct {
	*NewBaseDatabaseParams // leave nil for default
}

func NewPostGresDatabase(params *NewPostGresDatabaseParams) PostGresDatabaseInterface {
	if params.NewBaseDatabaseParams == nil {
		params.NewBaseDatabaseParams = &NewBaseDatabaseParams{}
	}
	return &PostGresDatabase{
		BaseDatabaseInterface: NewBaseDatabase(params.NewBaseDatabaseParams),
	}
}

func (d *PostGresDatabase) Connect() {
	if !d.IsConnected() {
		dsn := d.GetDBUrl()
		var db *gorm.DB
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

		for i := 0; i < retries; i++ { // try with converted retries count
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err == nil {
				d.database = db
				d.SetConnected(true)
				return
			}
			log.Printf("Database not ready yet, retrying... (%d/%d)", i+1, retries)
			time.Sleep(time.Duration(interval) * time.Second)
		}
		log.Fatal("Database connection failed after multiple attempts: ", err)
	}
}

func (d *PostGresDatabase) Disconnect() {
	if d.IsConnected() {
		db, err := d.database.DB()
		if err != nil {
			log.Fatal("Failed to disconnect from database: ", err)
		}
		db.Close()
		d.SetConnected(false)
	}
}

func (d *PostGresDatabase) GetDatabaseSession() *gorm.DB {
	if !d.IsConnected() {
		d.Connect()
	}
	return d.database
}

func (d *PostGresDatabase) GetNewDatabaseSession() *gorm.DB {
	return d.GetDatabaseSession().Session(&gorm.Session{NewDB: true})
}

// EntityDataInterface and EntityData implementation.
type EntityDataInterface[T entity.EntityInterface] interface {
	PostGresDatabaseInterface
	GetByID(ID string) (T, error)
	GetByIDs(IDs []string) (*[]T, error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error)
	GetAll() (*[]T, error)
	Create(entity T) error
	CreateBulk(entities *[]T) error
	Update(entity T) error
	Delete(ID string) error
	Exists(ID string) (bool, error)
}

type EntityData[T entity.EntityInterface] struct {
	PostGresDatabaseInterface
}

type NewEntityDataParams struct {
	*NewPostGresDatabaseParams                           // leave nil for default, not used if existing is provided
	Existing                   PostGresDatabaseInterface // leave nil for new database connection
}

func NewEntityData[T entity.EntityInterface](params *NewEntityDataParams) EntityDataInterface[T] {
	if params.NewPostGresDatabaseParams == nil {
		params.NewPostGresDatabaseParams = &NewPostGresDatabaseParams{}
	}

	if params.Existing == nil {
		params.Existing = NewPostGresDatabase(params.NewPostGresDatabaseParams)
	}
	return &EntityData[T]{
		PostGresDatabaseInterface: params.Existing,
	}
}

func (d *EntityData[T]) Exists(ID string) (bool, error) {
	var ent T

	result := d.GetNewDatabaseSession().First(&ent, "id = ?", ID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if result.Error != nil {
		fmt.Printf("error checking if entity exists: %s", result.Error.Error())
		return false, result.Error
	}
	return true, nil
}

func (d *EntityData[T]) GetByID(id string) (T, error) {
	var ent T
	result := d.GetDatabaseSession().First(&ent, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		var zero T
		fmt.Printf("record not found for id: %s", id)
		return zero, result.Error
	}
	if result.Error != nil {
		var zero T
		fmt.Printf("error getting: %s", result.Error.Error())
		return zero, result.Error
	}
	return ent, nil
}

func (d *EntityData[T]) GetByIDs(ids []string) (*[]T, error) {
	var entities []T
	results := d.GetDatabaseSession().Find(&entities, "id IN ?", ids)
	if results.Error != nil {
		fmt.Printf("error getting by ids: %s", results.Error.Error())
		return nil, results.Error
	}
	return &entities, nil
}

func (d *EntityData[T]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error) {
	var entities []T
	results := d.GetDatabaseSession().Find(&entities, foreignIDColumn+" = ?", foreignID)
	if results.Error != nil {
		fmt.Printf("error getting by foreignKey: %s", results.Error.Error())
		return nil, results.Error
	}
	println("Printing Entities")
	for _, entity := range entities {
		jso, _ := entity.ToJSON()
		println("Entity: ", string(jso))
	}
	return &entities, nil
}

func (d *EntityData[T]) GetAll() (*[]T, error) {
	var entities []T
	d.GetDatabaseSession().Find(&entities)
	return &entities, nil
}

func (d *EntityData[T]) CreateBulk(entities *[]T) error {
	maxInsertCount, err := strconv.Atoi(os.Getenv("MAX_DB_INSERT_COUNT"))
	if err != nil {
		fmt.Printf("error getting max insert count: %s", err.Error())
		return err
	}

	result := d.GetNewDatabaseSession().CreateInBatches(&entities, maxInsertCount)
	if result.Error != nil {
		fmt.Printf("error creating entities in bulk: %s", result.Error.Error())
		return result.Error
	}
	return nil
}

func (d *EntityData[T]) Create(entity T) error {
	// json, _ := entity.ToJSON()
	// print("Creating entity: ", string(json))
	result := d.GetNewDatabaseSession().Create(&entity)
	//if we have a conflicting ID
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			candidateID := generateRandomID()
			for {
				newEnt := entity
				newEnt.SetId(candidateID)
				result := d.GetNewDatabaseSession().Create(&entity)
				//result := d.GetNewDatabaseSession().FirstOrCreate(&newEnt, "id = ?", candidateID)
				if result.Error != nil {
					if errors.Is(result.Error, gorm.ErrRecordNotFound) {
						candidateID = generateRandomID()
						//continue
					}
					fmt.Printf("error checking if entity exists: %s", result.Error.Error())
					return result.Error
				} else {
					entity.SetId(candidateID)
					entity.SetDateCreated(newEnt.GetDateCreated())
					entity.SetDateModified(newEnt.GetDateModified())
					break
				}

				// result, err := d.Exists()
				// if err != nil {
				// 	fmt.Printf("error checking existing: %s", err.Error())
				// 	return err
				// }

				// if !result {
				// 	break
				// }
			}
		} else {
			fmt.Printf("error creating %s: %s", entity.GetId(), result.Error.Error())
			return result.Error
		}
	}

	// entity.SetId(candidateID)
	// createResult := d.GetDatabaseSession().Create(entity)

	// if createResult.Error != nil {
	// 	fmt.Printf("error creating %s: %s", entity.GetId(), createResult.Error.Error())
	// 	return createResult.Error
	// }
	return nil
}

func generateRandomID() string {
	// Generate a new UUID as the stock ID
	return uuid.New().String()
}

func (d *EntityData[T]) Update(entity T) error {
	updateResult := d.GetDatabaseSession().Save(entity)
	if updateResult.Error != nil {
		fmt.Printf("error updating %s: %s", entity.GetId(), updateResult.Error.Error())
		return updateResult.Error
	}
	return nil
}

func (d *EntityData[T]) Delete(id string) error {
	// _, err := d.GetByID(id)
	// if err != nil {
	// 	return err
	// }
	var zero T
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id = ?", id)
	if deleteResult.Error != nil {
		fmt.Printf("error deleting %s: %s", id, deleteResult.Error.Error())
		return deleteResult.Error
	}
	return nil
}

// Caching code below
type CachedEntityData[T entity.EntityInterface] struct {
	underlying  EntityDataInterface[T]
	redisClient *redis.Client
	defaultTTL  time.Duration
}

type NewCachedEntityDataParams struct {
	*NewEntityDataParams
	RedisAddr  string
	Password   string
	DefaultTTL time.Duration
}

func NewCachedEntityData[T entity.EntityInterface](params *NewCachedEntityDataParams) *CachedEntityData[T] {
	log.Printf("[Cache Init] Creating Redis client with Addr=%s, TTL=%s", params.RedisAddr, params.DefaultTTL)
	rdb := redis.NewClient(&redis.Options{
		Addr:     params.RedisAddr,
		Password: params.Password,
		DB:       0,
	})
	return &CachedEntityData[T]{
		underlying:  NewEntityData[T](params.NewEntityDataParams),
		redisClient: rdb,
		defaultTTL:  params.DefaultTTL,
	}
}

func (c *CachedEntityData[T]) redisKey(id string) string {
	key := "entity:" + id
	log.Printf("[Cache Key] Generated key: %s", key)
	return key
}

// Delegate BaseDatabaseInterface methods.
func (c *CachedEntityData[T]) GetDBUrl() string {
	return c.underlying.GetDBUrl()
}

func (c *CachedEntityData[T]) IsConnected() bool {
	return c.underlying.IsConnected()
}

func (c *CachedEntityData[T]) SetConnected(connected bool) {
	c.underlying.SetConnected(connected)
}

// Delegate DatabaseInterface methods.
func (c *CachedEntityData[T]) Connect() {
	log.Println("[Cache] Connect called")
	c.underlying.Connect()
}

func (c *CachedEntityData[T]) Disconnect() {
	log.Println("[Cache] Disconnect called")
	c.underlying.Disconnect()
}

// Delegate PostGresDatabaseInterface methods.
func (c *CachedEntityData[T]) GetDatabaseSession() *gorm.DB {
	return c.underlying.GetDatabaseSession()
}

func (c *CachedEntityData[T]) GetNewDatabaseSession() *gorm.DB {
	return c.underlying.GetNewDatabaseSession()
}

func (c *CachedEntityData[T]) Exists(ID string) (bool, error) {
	return c.underlying.Exists(ID)
}

func (c *CachedEntityData[T]) GetByID(id string) (T, error) {
	ctx := context.Background()
	var zero T
	key := c.redisKey(id)
	log.Printf("[Cache] GetByID: Looking for key: %s", key)
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		log.Printf("[Cache] GetByID: Found data in cache for key %s: %s", key, data)
		var cachedEntity T
		if err = json.Unmarshal([]byte(data), &cachedEntity); err == nil {
			log.Printf("[Cache] GetByID: Successfully unmarshaled cached data for key %s", key)
			return cachedEntity, nil
		}
		log.Printf("[Cache] GetByID: Error unmarshaling cache for key %s: %v", key, err)
	} else if err != redis.Nil {
		log.Printf("[Cache] GetByID: Redis GET error for key %s: %v", key, err)
	} else {
		log.Printf("[Cache] GetByID: Cache miss for key %s", key)
	}

	// Fallback to DB
	log.Printf("[Cache] GetByID: Falling back to underlying DB for id: %s", id)
	dbEntity, err := c.underlying.GetByID(id)
	if err != nil {
		log.Printf("[Cache] GetByID: Underlying DB error for id %s: %v", id, err)
		return zero, err
	}

	// Cache the result
	jsonBytes, err := json.Marshal(dbEntity)
	if err != nil {
		log.Printf("[Cache] GetByID: Error marshaling entity for id %s: %v", id, err)
	} else {
		log.Printf("[Cache] GetByID: Caching entity for id %s with key %s", id, key)
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] GetByID: Error setting cache for key %s: %v", key, err)
		} else {
			log.Printf("[Cache] GetByID: Successfully cached key %s with TTL %s", key, c.defaultTTL)
		}
	}
	return dbEntity, nil
}

func (c *CachedEntityData[T]) GetByIDs(ids []string) (*[]T, error) {
	ctx := context.Background()
	entityMap := make(map[string]T)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = c.redisKey(id)
	}
	log.Printf("[Cache] GetByIDs: Looking up keys: %v", keys)
	results, err := c.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("[Cache] GetByIDs: Redis MGet error: %v", err)
	}
	var missingIds []string
	for i, res := range results {
		id := ids[i]
		if res == nil {
			log.Printf("[Cache] GetByIDs: Key not found for id %s", id)
			missingIds = append(missingIds, id)
			continue
		}
		str, ok := res.(string)
		if !ok {
			log.Printf("[Cache] GetByIDs: Value for id %s is not a string", id)
			missingIds = append(missingIds, id)
			continue
		}
		var cachedEntity T
		if err := json.Unmarshal([]byte(str), &cachedEntity); err != nil {
			log.Printf("[Cache] GetByIDs: Error unmarshaling cached entity for id %s: %v", id, err)
			missingIds = append(missingIds, id)
		} else {
			log.Printf("[Cache] GetByIDs: Successfully retrieved cached entity for id %s", id)
			entityMap[id] = cachedEntity
		}
	}
	if len(missingIds) > 0 {
		log.Printf("[Cache] GetByIDs: Missing ids from cache: %v. Fetching from DB...", missingIds)
		dbEntities, err := c.underlying.GetByIDs(missingIds)
		if err != nil {
			log.Printf("[Cache] GetByIDs: DB error for missing ids %v: %v", missingIds, err)
			return nil, err
		}
		for _, entity := range *dbEntities {
			id := entity.GetId()
			entityMap[id] = entity
			if jsonBytes, err := json.Marshal(entity); err == nil {
				_ = c.redisClient.Set(ctx, c.redisKey(id), jsonBytes, c.defaultTTL).Err()
				log.Printf("[Cache] GetByIDs: Cached entity for id %s", id)
			} else {
				log.Printf("[Cache] GetByIDs: Error marshaling entity for id %s: %v", id, err)
			}
		}
	}
	finalEntities := make([]T, 0, len(ids))
	for _, id := range ids {
		if entity, exists := entityMap[id]; exists {
			finalEntities = append(finalEntities, entity)
		} else {
			log.Printf("[Cache] GetByIDs: No entity found for id %s in final aggregation", id)
		}
	}
	return &finalEntities, nil
}

func (c *CachedEntityData[T]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error) {
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
}

func (c *CachedEntityData[T]) GetAll() (*[]T, error) {
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

func (c *CachedEntityData[T]) Create(entity T) error {
	log.Printf("[Cache] Create: Creating entity with id %s", entity.GetId())
	if err := c.underlying.Create(entity); err != nil {
		log.Printf("[Cache] Create: Underlying DB create failed for id %s: %v", entity.GetId(), err)
		return err
	}
	ctx := context.Background()
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		log.Printf("[Cache] Create: Error marshaling entity with id %s: %v", entity.GetId(), err)
	} else {
		key := c.redisKey(entity.GetId())
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] Create: Error caching entity with key %s: %v", key, err)
		} else {
			log.Printf("[Cache] Create: Successfully cached entity with key %s", key)
		}
	}
	return nil
}

func (c *CachedEntityData[T]) Update(entity T) error {
	log.Printf("[Cache] Update: Updating entity with id %s", entity.GetId())
	if err := c.underlying.Update(entity); err != nil {
		log.Printf("[Cache] Update: Underlying DB update failed for id %s: %v", entity.GetId(), err)
		return err
	}
	ctx := context.Background()
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		log.Printf("[Cache] Update: Error marshaling entity with id %s: %v", entity.GetId(), err)
	} else {
		key := c.redisKey(entity.GetId())
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] Update: Error caching updated entity with key %s: %v", key, err)
		} else {
			log.Printf("[Cache] Update: Successfully cached updated entity with key %s", key)
		}
	}
	return nil
}

func (c *CachedEntityData[T]) Delete(id string) error {
	log.Printf("[Cache] Delete: Deleting entity with id %s", id)
	if err := c.underlying.Delete(id); err != nil {
		log.Printf("[Cache] Delete: Underlying DB delete failed for id %s: %v", id, err)
		return err
	}
	ctx := context.Background()
	key := c.redisKey(id)
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[Cache] Delete: Error deleting cache for key %s: %v", key, err)
	} else {
		log.Printf("[Cache] Delete: Successfully deleted cache for key %s", key)
	}
	return nil
}

func (c *CachedEntityData[T]) CreateBulk(entities *[]T) error {
	log.Printf("[Cache] CreateBulk: Creating %d entities", len(*entities))
	if err := c.underlying.CreateBulk(entities); err != nil {
		log.Printf("[Cache] CreateBulk: Underlying DB bulk create failed: %v", err)
		return err
	}
	ctx := context.Background()
	for _, entity := range *entities {
		jsonBytes, err := json.Marshal(entity)
		if err != nil {
			log.Printf("[Cache] CreateBulk: Error marshaling entity with id %s: %v", entity.GetId(), err)
			continue
		}
		key := c.redisKey(entity.GetId())
		if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
			log.Printf("[Cache] CreateBulk: Error caching entity with key %s: %v", key, err)
		} else {
			log.Printf("[Cache] CreateBulk: Successfully cached entity with key %s", key)
		}
	}
	return nil
}
