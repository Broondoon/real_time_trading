package database

import (
	"Shared/entities/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
	GetByIDs(IDs []string) (*[]T, map[string]error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, map[string]error)
	GetAll() (*[]T, error)
	Create(ent T) error
	CreateBulk(entities *[]T) map[string]error
	//Update(entity T) error
	//UpdateBulk(entities *[]T) error
	Delete(ID string) error
	DeleteBulk(IDs []string) map[string]error

	//I need a safe updater for numerical values... we can't pass it the updated entity, we have to pass it the values to change the fields by.
	Update([]*entity.EntityUpdateData) map[string]error
	//collect all the values for a string where the fields are the same
}

type columnCacheEntry struct {
	ColumnName string
	FieldType  reflect.Type
}

type EntityData[T entity.EntityInterface] struct {
	PostGresDatabaseInterface
	tableName   string
	columnCache map[string]columnCacheEntry
	// *gorm.DB //note, this allows us to treat this as a gorm.DB WITHIN the EntityData struct. This is not exposed as part of the interface, and thus cannot be used like this with the interface.
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

	// Create an instance with an empty column cache.
	ed := &EntityData[T]{
		PostGresDatabaseInterface: params.Existing,
		columnCache:               make(map[string]columnCacheEntry),
	}

	// Determine the table name for type T.
	tableName, err := ed.getTableName()
	if err != nil {
		panic(fmt.Sprintf("failed to get table name: %v", err))
	}
	ed.tableName = tableName

	// Parse the GORM schema for type T.
	sch, err := ed.getGormSchema()
	if err != nil {
		panic(fmt.Sprintf("failed to get GORM schema: %v", err))
	}

	// Cache the column names for each struct field.
	for _, fieldSchema := range sch.Fields {
		colName := fieldSchema.DBName // e.g. "user_id"
		fieldName := fieldSchema.Name // e.g. "UserID"

		// Instead of indexing by fieldName, index by DBName
		ed.columnCache[colName] = columnCacheEntry{
			ColumnName: colName,
			FieldType:  fieldSchema.FieldType,
		}
		ed.columnCache[fieldName] = columnCacheEntry{
			ColumnName: colName,
			FieldType:  fieldSchema.FieldType,
		}
	}
	log.Println(ed.tableName, " ID Data type: ", ed.columnCache["ID"].FieldType)

	return ed
}

func convertID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse id %s: %v", id, err)
	}
	return uid, nil
}

func convertIDs(ids []string, errors map[string]error) ([]uuid.UUID, map[string]error) {
	uids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		uid, err := convertID(id)
		if err != nil {
			errors[id] = err
			continue
		}
		uids = append(uids, uid)
	}
	return uids, errors
}

func (d *EntityData[T]) PrintOutEntities() {
	entities, err := d.GetAll()
	if err != nil {
		log.Printf("error getting all: %s", err.Error())
		return
	}
	for _, ent := range *entities {
		json, _ := ent.ToJSON()
		log.Println(string(json))
	}
}

func (d *EntityData[T]) GetByID(id string) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("ID is empty")
	}
	var ent T
	uid, err := convertID(id)
	if err != nil {
		log.Printf("error getting: %s", err.Error())
		return zero, err
	}
	result := d.GetDatabaseSession().First(&ent, "id = ?", uid)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("record not found for id: %s", id)
		d.PrintOutEntities()
		return zero, result.Error
	}
	if result.Error != nil {
		log.Printf("error getting: %s", result.Error.Error())
		d.PrintOutEntities()
		return zero, result.Error
	}
	return ent, nil
}

func (d *EntityData[T]) GetByIDs(ids []string) (*[]T, map[string]error) {
	if len(ids) == 0 {
		return nil, map[string]error{"transaction": errors.New("no ids provided")}
	}
	var entities []T
	errors := make(map[string]error)
	uids, errors := convertIDs(ids, errors)
	if len(uids) == 0 {
		return nil, errors
	}

	results := d.GetDatabaseSession().Find(&entities, "id IN ?", uids)
	if results.Error != nil {
		errors["transaction"] = results.Error
		log.Printf("error getting by ids: %s", results.Error.Error())
		d.PrintOutEntities()

		return nil, errors
	}
	//get all ids in ids that are not in entities
	idsFound := make(map[string]bool)
	for _, ent := range entities {
		idsFound[ent.GetIdString()] = true
	}
	for _, id := range ids {
		if val, ok := idsFound[id]; !ok && !val {
			errors[id] = gorm.ErrRecordNotFound
			d.PrintOutEntities()
		}
	}

	return &entities, errors
}

// This needs the table column names, whihc is a little diffrent
func (d *EntityData[T]) GetByForeignID(foreignIDKey string, foreignID string) (*[]T, error) {
	if foreignIDKey == "" {
		err := fmt.Errorf("foreign key column is empty")
		log.Printf("error getting by foreignKey: %s", err.Error())
		return nil, err
	}
	if foreignID == "" {
		err := fmt.Errorf("foreign key is empty")
		log.Printf("error getting by foreignKey: %s", err.Error())
		return nil, err
	}

	var entities []T
	foreignIDColumn, ok := d.columnCache[foreignIDKey]
	if !ok {
		err := fmt.Errorf("foreign key column %s not found", foreignIDKey)
		log.Printf("error getting by foreignKey: %s", err.Error())
		columns := make([]string, 0, len(d.columnCache))
		for _, d := range d.columnCache {
			columns = append(columns, d.ColumnName)
		}
		log.Println("avalaible columns: ", strings.Join(columns, ", "))
		return nil, err
	}
	var results *gorm.DB

	if strings.Contains(foreignIDColumn.ColumnName, "_id") || foreignIDColumn.ColumnName == "id" {
		uid, err := convertID(foreignID)
		if err != nil {
			log.Printf("error getting by foreignKey: %s", err.Error())
			return nil, err
		}
		results = d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" = ?", uid)
	} else {
		results = d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" = ?", foreignID)
	}
	if results.Error != nil {
		log.Printf("error getting by foreignKey: %s", results.Error.Error())
		d.PrintOutEntities()
		return nil, results.Error
	}
	return &entities, nil
}

func (d *EntityData[T]) GetByForeignIDBulk(foreignIDKey string, foreignIDs []string) (*[]T, map[string]error) {
	if foreignIDKey == "" {
		err := fmt.Errorf("foreign key column is empty")
		log.Printf("error getting by foreignKey: %s", err.Error())
		return nil, map[string]error{"transaction": err}
	}
	if len(foreignIDs) == 0 {
		err := fmt.Errorf("foreign key is empty")
		log.Printf("error getting by foreignKey: %s", err.Error())
		return nil, map[string]error{"transaction": err}
	}

	var entities []T
	errors := make(map[string]error)
	foreignIDColumn, ok := d.columnCache[foreignIDKey]
	if !ok {
		errors["transaction"] = fmt.Errorf("foreign key column %s not found", foreignIDKey)
		log.Printf("error getting by foreignKey: %s", errors["transaction"].Error())
		columns := make([]string, 0, len(d.columnCache))
		for _, d := range d.columnCache {
			columns = append(columns, d.ColumnName)
		}
		log.Println("avalaible columns: ", strings.Join(columns, ", "))
		return nil, errors
	}
	var results *gorm.DB
	println("key: ", foreignIDKey, "Foreign ID Column: ", foreignIDColumn.ColumnName)
	if strings.Contains(foreignIDColumn.ColumnName, "_id") || foreignIDColumn.ColumnName == "id" {
		uids, errors := convertIDs(foreignIDs, errors)
		if len(uids) == 0 {
			return nil, errors
		}
		results = d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" IN ?", uids)
	} else {
		results = d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" IN ?", foreignIDs)
	}
	if results.Error != nil {
		errors["transaction"] = results.Error
		log.Printf("error getting by foreignKey: %s", results.Error.Error())
		d.PrintOutEntities()
		return nil, errors
	}

	//get all ids in ids that are not in entities
	idsFound := make(map[string]bool)
	for _, ent := range entities {
		val := reflect.ValueOf(ent)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		fieldVal := val.FieldByName(foreignIDKey)
		switch actual := fieldVal.Interface().(type) {
		case uuid.UUID:
			// If the field is a value type
			foreignID := actual.String()
			idsFound[foreignID] = true
		case *uuid.UUID:
			// If the field is a pointer type
			if actual != nil {
				foreignID := actual.String()
				idsFound[foreignID] = true
				log.Println("Foreign ID Found: ", actual.String())
			} else {
				// Possibly store an empty string or skip
				continue
			}
		default:
			foreignID := fieldVal.String()
			log.Println("Foreign ID Found: ", foreignID)
			idsFound[foreignID] = true
		}
	}
	for _, id := range foreignIDs {
		log.Println("Checking for foreign ID: ", id)
		if val, ok := idsFound[id]; !ok || !val {
			d.PrintOutEntities()
			errors[id] = gorm.ErrRecordNotFound
		}
	}
	return &entities, errors
}

func (d *EntityData[T]) GetAll() (*[]T, error) {
	var entities []T
	result := d.GetDatabaseSession().Find(&entities)
	if result.Error != nil {
		log.Printf("error getting all: %s", result.Error.Error())
		return nil, result.Error
	}
	return &entities, nil
}

func (d *EntityData[T]) CreateBulk(entities *[]T) map[string]error {
	if len(*entities) == 0 {
		return map[string]error{"transaction": errors.New("CREATE: no entities provided")}
	}

	// errorMap accumulates errors keyed by the entity's ID.
	errorMap := make(map[string]error)

	maxInsertCount, err := strconv.Atoi(os.Getenv("MAX_DB_INSERT_COUNT"))
	if err != nil {
		log.Printf("error getting max insert count: %s", err.Error())
		errorMap["transaction"] = err
		return errorMap
	}

	result := d.GetNewDatabaseSession().CreateInBatches(&entities, maxInsertCount)
	if result.Error != nil {

		// Get a new database session and begin a transaction.
		db := d.GetNewDatabaseSession()
		tx := db.Begin()
		if tx.Error != nil {
			errorMap["transaction"] = tx.Error
			return errorMap
		}

		// Use a counter to generate unique savepoint names.
		spCounter := 0

		// Process each entity individually.
		for i := range *entities {
			ent := (*entities)[i]
			spCounter++
			spName := fmt.Sprintf("sp_%d", spCounter)
			tx.SavePoint(spName)

			// Try inserting the entity.
			if err := tx.Create(&ent).Error; err != nil {
				// If an error occurs, rollback to the savepoint so that this insert is undone.
				val := reflect.ValueOf(ent)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				tx.RollbackTo(spName)
				errorMap[ent.GetUniquePairing().String()] = fmt.Errorf("error creating entity: %v", err)
			}
			// Continue to the next entity.
			continue
		}
		// Optionally, you can log successful insertions if needed.

		// Commit the transaction.
		if err := tx.Commit().Error; err != nil {
			// If the commit itself fails, record a transaction-level error.
			errorMap["transaction"] = fmt.Errorf("failed to commit transaction: %v", err)
		}
	}
	return errorMap
}

func (d *EntityData[T]) Create(ent T) error {
	// json, _ := entity.ToJSON()
	// print("Creating entity: ", string(json))
	result := d.GetNewDatabaseSession().Create(&ent)
	//if we have a conflicting ID
	if result.Error != nil {
		ent.SetId(nil)
		result = d.GetNewDatabaseSession().Create(&ent)
		if result.Error != nil {
			log.Printf("error creating %s: %s", ent.GetId(), result.Error.Error())
			return result.Error
		}
	}

	// entity.SetId(candidateID)
	// createResult := d.GetDatabaseSession().Create(entity)

	// if createResult.Error != nil {
	// 	log.Printf("error creating %s: %s", entity.GetId(), createResult.Error.Error())
	// 	return createResult.Error
	// }
	return nil
}

// Generated with assistance of Chat GPT 03-mini-high: https://chatgpt.com/share/67cb6dc5-7cf4-8006-a7cc-b33fa7765051

func (d *EntityData[T]) Update(updates []*entity.EntityUpdateData) map[string]error {
	if len(updates) == 0 {
		return map[string]error{"transaction": errors.New("UPDATE: no updates provided")}
	}

	// errorMap will accumulate errors keyed by row ID.
	errorMap := make(map[string]error)
	// Aggregate new and alter updates.
	newUpdates := make(map[string]map[string]string)    // field -> (row ID -> new value)
	alterUpdates := make(map[string]map[string]float64) // field -> (row ID -> cumulative delta)

	for _, upd := range updates {
		if upd.NewValue != nil {
			if newUpdates[upd.Field] == nil {
				newUpdates[upd.Field] = make(map[string]string)
			}
			newUpdates[upd.Field][upd.ID.String()] = *upd.NewValue
		} else if upd.AlterValue != nil {
			parsed, err := strconv.ParseFloat(*upd.AlterValue, 64)
			if err != nil {
				errorMap[upd.ID.String()] = fmt.Errorf("failed to parse alter value '%s' for field %s: %v", *upd.AlterValue, upd.Field, err)
				continue
			}
			if alterUpdates[upd.Field] == nil {
				alterUpdates[upd.Field] = make(map[string]float64)
			}
			alterUpdates[upd.Field][upd.ID.String()] += parsed
		}
	}

	// Helper to convert new value to proper type and return SQL cast type.
	convertNewValue := func(newVal string, fieldType reflect.Type) (interface{}, string, error) {
		// Check if the field type is uuid.UUID
		if fieldType == reflect.TypeOf(uuid.UUID{}) || fieldType == reflect.TypeOf(&uuid.UUID{}) {
			uid, err := uuid.Parse(newVal)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse '%s' as UUID: %v", newVal, err)
			}
			return uid, "uuid", nil
		}
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(newVal, 10, 64)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse '%s' as integer: %v", newVal, err)
			}
			return i, "bigint", nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(newVal, 10, 64)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse '%s' as unsigned integer: %v", newVal, err)
			}
			return u, "bigint", nil
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(newVal, 64)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse '%s' as float: %v", newVal, err)
			}
			return f, "double precision", nil
		default:
			println("testing for ID column. Field type is ", fieldType.Kind().String())
			return newVal, "text", nil
		}
	}

	// Helper to convert alter delta to proper type and return SQL cast type.
	convertDelta := func(delta float64, fieldType reflect.Type) (interface{}, string, error) {
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(delta), "bigint", nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return uint64(delta), "bigint", nil
		case reflect.Float32, reflect.Float64:
			return delta, "double precision", nil
		default:
			return nil, "", fmt.Errorf("unsupported numeric field type %s", fieldType.Kind())
		}
	}

	// Bulk update transaction.
	err := d.GetNewDatabaseSession().Transaction(func(tx *gorm.DB) error {
		// Process new value updates in bulk.
		for field, idToNewVal := range newUpdates {
			cacheEntry, ok := d.columnCache[field]
			if !ok {
				return fmt.Errorf("unknown field %s in column cache", field)
			}
			var valueTuples []string
			var args []interface{}
			for id, newVal := range idToNewVal {
				converted, castType, err := convertNewValue(newVal, cacheEntry.FieldType)
				if err != nil {
					return fmt.Errorf("field %s for id %s: %v", field, id, err)
				}
				valueTuples = append(valueTuples, fmt.Sprintf("(?::uuid, ?::%s)", castType))
				uid, err := uuid.Parse(id)
				if err != nil {
					return fmt.Errorf("failed to parse id %s: %v", id, err)
				}
				args = append(args, uid, converted)
			}
			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = u.delta
				FROM (VALUES %s) AS u(id, delta)
    			WHERE t.id = u.id
			`, d.tableName, cacheEntry.ColumnName, strings.Join(valueTuples, ", "))
			if err := tx.Exec(query, args...).Error; err != nil {
				return fmt.Errorf("failed bulk new value update for field '%s': %v", field, err)
			}
		}

		// Process alter value updates in bulk.
		for field, idToDelta := range alterUpdates {
			cacheEntry, ok := d.columnCache[field]
			if !ok {
				return fmt.Errorf("unknown field %s in column cache", field)
			}
			var valueTuples []string
			var args []interface{}
			for id, delta := range idToDelta {
				deltaValue, castType, err := convertDelta(delta, cacheEntry.FieldType)
				if err != nil {
					return fmt.Errorf("field %s for id %s: %v", field, id, err)
				}
				valueTuples = append(valueTuples, fmt.Sprintf("(?::uuid, ?::%s)", castType))
				uid, err := uuid.Parse(id)
				if err != nil {
					return fmt.Errorf("failed to parse id %s: %v", id, err)
				}
				args = append(args, uid, deltaValue)
			}
			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = t.%s + u.delta
				FROM (VALUES %s) AS u(id, delta)
    			WHERE t.id = u.id
			`, d.tableName, cacheEntry.ColumnName, cacheEntry.ColumnName, strings.Join(valueTuples, ", "))
			if err := tx.Exec(query, args...).Error; err != nil {
				return fmt.Errorf("failed bulk alter value update for field '%s': %v", field, err)
			}
		}
		return nil
	})
	if err == nil {
		return nil
	}

	// Fallback: update row-by-row if bulk update fails.
	tx := d.GetNewDatabaseSession().Begin()
	if tx.Error != nil {
		errorMap["transaction"] = tx.Error
		return errorMap
	}
	spCounter := 0

	// Process new value updates row-by-row.
	for field, idToNewVal := range newUpdates {
		cacheEntry, ok := d.columnCache[field]
		if !ok {
			for id := range idToNewVal {
				errorMap[id] = fmt.Errorf("unknown field %s in column cache", field)
			}
			continue
		}
		var castType string
		switch cacheEntry.FieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			castType = "bigint"
		case reflect.Float32, reflect.Float64:
			castType = "double precision"
		default:
			castType = "text"
		}
		for id, newVal := range idToNewVal {
			spCounter++
			spName := fmt.Sprintf("sp_new_%d", spCounter)
			tx.SavePoint(spName)
			converted, _, err := convertNewValue(newVal, cacheEntry.FieldType)
			if err != nil {
				errorMap[id] = fmt.Errorf("failed to convert new value for field '%s': %v", field, err)
				tx.RollbackTo(spName)
				continue
			}
			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = CAST(? AS %s)
    			WHERE t.id = ?
			`, d.tableName, cacheEntry.ColumnName, castType)
			uid, err := uuid.Parse(id)
			if err != nil {
				errorMap[id] = fmt.Errorf("failed to parse id %s: %v", id, err)
				tx.RollbackTo(spName)
				continue
			}
			if err := tx.Exec(query, converted, uid).Error; err != nil {
				tx.RollbackTo(spName)
				errorMap[id] = fmt.Errorf("failed new value update for field '%s': %v", field, err)
			}
		}
	}

	// Process alter value updates row-by-row.
	for field, idToDelta := range alterUpdates {
		cacheEntry, ok := d.columnCache[field]
		if !ok {
			for id := range idToDelta {
				errorMap[id] = fmt.Errorf("unknown field %s in column cache", field)
			}
			continue
		}
		var castType string
		switch cacheEntry.FieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			castType = "bigint"
		case reflect.Float32, reflect.Float64:
			castType = "double precision"
		default:
			errorMap["general"] = fmt.Errorf("unsupported numeric field type %s for field %s", cacheEntry.FieldType.Kind(), field)
			continue
		}
		for id, delta := range idToDelta {
			spCounter++
			spName := fmt.Sprintf("sp_alter_%d", spCounter)
			tx.SavePoint(spName)
			deltaValue, _, err := convertDelta(delta, cacheEntry.FieldType)
			if err != nil {
				errorMap[id] = fmt.Errorf("failed to convert delta for field '%s': %v", field, err)
				tx.RollbackTo(spName)
				continue
			}
			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = t.%s + CAST(? AS %s)
    			WHERE t.id = ?
			`, d.tableName, cacheEntry.ColumnName, cacheEntry.ColumnName, castType)
			uid, err := uuid.Parse(id)
			if err != nil {
				errorMap[id] = fmt.Errorf("failed to parse id %s: %v", id, err)
				tx.RollbackTo(spName)
				continue
			}
			if err := tx.Exec(query, deltaValue, uid).Error; err != nil {
				tx.RollbackTo(spName)
				errorMap[id] = fmt.Errorf("failed alter value update for field '%s': %v", field, err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		errorMap["transaction"] = fmt.Errorf("failed to commit transaction: %v", err)
	}
	for id := range errorMap {
		if _, ok := errorMap[id]; !ok {
			log.Printf("error updating entity %s: %v", id, errorMap[id])
		}
	}
	return errorMap
}

func (d *EntityData[T]) Delete(id string) error {
	if id == "" {
		return errors.New("DELETE: id is required")
	}
	var zero T
	uuid, err := convertID(id)
	if err != nil {
		return err
	}
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id = ?", uuid)
	if deleteResult.Error != nil {
		log.Printf("error deleting %s: %s", id, deleteResult.Error.Error())
		return deleteResult.Error
	}
	return nil
}

func (d *EntityData[T]) DeleteBulk(ids []string) map[string]error {
	if len(ids) == 0 {
		return map[string]error{"transaction": errors.New("DELETE: no IDs provided")}
	}

	errorMap := make(map[string]error)
	var zero T
	uids, errorMap := convertIDs(ids, errorMap)
	if len(uids) == 0 {
		return errorMap
	}
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id IN ?", uids)
	if deleteResult.Error != nil {
		db := d.GetNewDatabaseSession()
		tx := db.Begin()
		if tx.Error != nil {
			errorMap["transaction"] = tx.Error
			return errorMap
		}

		// Use a counter to generate unique savepoint names.
		spCounter := 0

		// Process each entity individually.
		for _, id := range ids {
			spCounter++
			spName := fmt.Sprintf("sp_%d", spCounter)
			tx.SavePoint(spName)
			uid, err := convertID(id)
			if err != nil {
				errorMap[id] = fmt.Errorf("failed to parse id %s: %v", id, err)
				tx.RollbackTo(spName)
				continue
			}

			// Try inserting the entity.
			if err := tx.Delete(&zero, "id = ?", uid).Error; err != nil {
				// If an error occurs, rollback to the savepoint so that this insert is undone.
				tx.RollbackTo(spName)
				// Record the error keyed by the entity's ID.
				errorMap[id] = fmt.Errorf("error deleting entity: %v", err)
				// Continue to the next entity.
				continue
			}
			// Optionally, you can log successful insertions if needed.
		}

		// Commit the transaction.
		if err := tx.Commit().Error; err != nil {
			// If the commit itself fails, record a transaction-level error.
			errorMap["transaction"] = fmt.Errorf("failed to commit transaction: %v", err)
		}
	}
	return nil
}

// Generated with assistance of Chat GPT 03-mini-high: https://chatgpt.com/share/67cb6dc5-7cf4-8006-a7cc-b33fa7765051
// getGormSchema parses and returns the GORM schema for type T.
// It uses a sync.Map as a cache placeholder (you might want to cache the schema for performance).
func (d *EntityData[T]) getGormSchema() (*schema.Schema, error) {
	var t T
	typ := reflect.TypeOf(t)
	// If T is a pointer, use its underlying type.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	// Create a new instance to pass to schema.Parse.
	return schema.Parse(reflect.New(typ).Interface(), &sync.Map{}, d.GetDatabaseSession().NamingStrategy)
}

// Generated with assistance of Chat GPT 03-mini-high: https://chatgpt.com/share/67cb6dc5-7cf4-8006-a7cc-b33fa7765051
// getTableName returns the table name for type T.
// It first checks if T implements a TableName() string method, and if not, uses the GORM naming strategy.
func (d *EntityData[T]) getTableName() (string, error) {
	var t T
	if tn, ok := any(t).(interface{ TableName() string }); ok {
		return tn.TableName(), nil
	}
	if tn, ok := any(&t).(interface{ TableName() string }); ok {
		return tn.TableName(), nil
	}
	sch, err := d.getGormSchema()
	if err != nil {
		return "", err
	}
	return sch.Table, nil
}

// Generated with assistance of Chat GPT 03-mini-high: https://chatgpt.com/share/67cb6dc5-7cf4-8006-a7cc-b33fa7765051
// getColumnName returns the column name in the database for the given struct field name.
func (d *EntityData[T]) getColumnName(fieldName string) (string, error) {
	sch, err := d.getGormSchema()
	if err != nil {
		return "", err
	}
	fieldSchema, ok := sch.FieldsByName[fieldName]
	if !ok {
		return "", fmt.Errorf("field %s not found in schema", fieldName)
	}
	return fieldSchema.DBName, nil
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

func (c *CachedEntityData[T]) GetByIDs(ids []string) (*[]T, error) {
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
		dbEntities, err := c.underlying.GetByIDs(missingIds)
		if err != nil {
			log.Printf("[Cache] GetByIDs: ❌ Database error for missing IDs %v: %v", missingIds, err)
			return nil, err
		}

		// Step 3: Cache newly retrieved entities
		for _, entity := range *dbEntities {
			id := entity.GetId()
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

	// Step 4: Aggregate results
	finalEntities := make([]T, 0, len(ids))
	for _, id := range ids {
		if entity, exists := entityMap[id]; exists {
			finalEntities = append(finalEntities, entity)
		} else {
			log.Printf("[Cache] GetByIDs: ❌ No entity found for ID %s", id)
		}
	}
	return &finalEntities, nil
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

func (c *CachedEntityData[T]) GetByForeignID(foreignIDColumn, foreignID string) (*[]T, error) {
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

func (c *CachedEntityData[T]) Create(entity T) error {
	if err := c.underlying.Create(entity); err != nil {
		return err
	}

	ctx := context.Background()
	entityKey := c.redisKey(entity.GetId())

	// Cache entity after DB insertion
	if jsonBytes, err := json.Marshal(entity); err == nil {
		c.redisClient.Set(ctx, entityKey, jsonBytes, c.defaultTTL)
		log.Printf("[Cache] Create: Cached entity with key %s", entityKey)
	}

	return nil
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

func (c *CachedEntityData[T]) Update(entity T) error {
	log.Printf("[Cache] Update: Attempting to update entity with ID: %s", entity.GetId())

	// Step 1: Update entity in the database
	if err := c.underlying.Update(entity); err != nil {
		log.Printf("[Cache] Update: ❌ Failed to update entity in DB [ID: %s]: %v", entity.GetId(), err)
		return err
	}
	log.Printf("[Cache] Update: ✅ Successfully updated entity in DB [ID: %s]", entity.GetId())

	// Step 2: Remove outdated cache entry
	ctx := context.Background()
	key := c.redisKey(entity.GetId())

	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[Cache] Update: ❌ Failed to delete stale cache for [Key: %s]: %v", key, err)
	} else {
		log.Printf("[Cache] Update: ✅ Deleted stale cache for [Key: %s]", key)
	}

	// Step 3: Store updated entity in Redis
	jsonBytes, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		log.Printf("[Cache] Update: ❌ Error marshaling updated entity [ID: %s]: %v", entity.GetId(), err)
		return nil
	}

	if err := c.redisClient.Set(ctx, key, jsonBytes, c.defaultTTL).Err(); err != nil {
		log.Printf("[Cache] Update: ❌ Failed to cache updated entity in Redis [Key: %s]: %v", key, err)
	} else {
		log.Printf("[Cache] Update: ✅ Cached updated entity in Redis [Key: %s]:\n%s", key, string(jsonBytes))
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
