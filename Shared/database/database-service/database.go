package database

import (
	"Shared/entities/entity"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
	dsn := os.Getenv(d.DatabaseURLEnv) // "DATABASE_URL" is an ENV variable that
	// is set in docker-compose.yml
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

type EntityDataInterface[T entity.EntityInterface] interface {
	PostGresDatabaseInterface
	GetByID(ID string) (T, error)
	GetByIDs(IDs []string) (*[]T, map[string]error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, map[string]error)
	GetAll() (*[]T, error)
	Create(entity T) error
	CreateBulk(entities *[]T) map[string]error
	//Update(entity T) error
	//UpdateBulk(entities *[]T) error
	Delete(ID string) error
	DeleteBulk(IDs []string) map[string]error
	Exists(ID string) (bool, error)

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
	*NewPostGresDatabaseParams                           // leave nil for default, Not used if existing is provided
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

func (d *EntityData[T]) Exists(ID string) (bool, error) {
	if ID == "" {
		return false, fmt.Errorf("ID is empty")
	}
	var ent T

	result := d.GetNewDatabaseSession().First(&ent, "id = ?", ID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if result.Error != nil {
		log.Printf("error checking if entity exists: %s", result.Error.Error())
		return false, result.Error
	}
	return true, nil
}

func (d *EntityData[T]) GetByID(id string) (T, error) {
	if id == "" {
		var zero T
		return zero, fmt.Errorf("ID is empty")
	}
	var ent T
	result := d.GetDatabaseSession().First(&ent, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		var zero T
		log.Printf("record not found for id: %s", id)
		return zero, result.Error
	}
	if result.Error != nil {
		var zero T
		log.Printf("error getting: %s", result.Error.Error())
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
	results := d.GetDatabaseSession().Find(&entities, "id IN ?", ids)
	if results.Error != nil {
		errors["transaction"] = results.Error
		log.Printf("error getting by ids: %s", results.Error.Error())
		return nil, errors
	}
	//get all ids in ids that are not in entities
	idsFound := make(map[string]bool)
	for _, entity := range entities {
		idsFound[entity.GetId()] = true
	}
	for _, id := range ids {
		if val, ok := idsFound[id]; !ok && !val {
			errors[id] = gorm.ErrRecordNotFound
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
	results := d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" = ?", foreignID)
	if results.Error != nil {
		log.Printf("error getting by foreignKey: %s", results.Error.Error())
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

	results := d.GetDatabaseSession().Find(&entities, foreignIDColumn.ColumnName+" IN ?", foreignIDs)
	if results.Error != nil {
		errors["transaction"] = results.Error
		log.Printf("error getting by foreignKey: %s", results.Error.Error())
		return nil, errors
	}

	//get all ids in ids that are not in entities
	idsFound := make(map[string]bool)
	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		fieldVal := val.FieldByName(foreignIDKey)
		if !fieldVal.IsValid() {
			errors[entity.GetId()] = fmt.Errorf("foreign key column %s not found", foreignIDKey)
			continue
		}
		foreignID := fieldVal.String()
		idsFound[foreignID] = true
	}
	for _, id := range foreignIDs {
		if val, ok := idsFound[id]; !ok && !val {
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
			entity := (*entities)[i]
			spCounter++
			spName := fmt.Sprintf("sp_%d", spCounter)
			tx.SavePoint(spName)

			// Try inserting the entity.
			if err := tx.Create(&entity).Error; err != nil {
				// If an error occurs, rollback to the savepoint so that this insert is undone.
				tx.RollbackTo(spName)
				// Record the error keyed by the entity's ID.
				if timestampColumn, ok := d.columnCache["timestamp"]; ok {
					//get the timestamp
					timestamp := reflect.ValueOf(entity).FieldByName(timestampColumn.ColumnName).String()
					errorMap[timestamp] = fmt.Errorf("error creating entity: %v", err)
				} else if userColumn, ok := d.columnCache["user_id"]; ok {
					//get the user_id
					userID := reflect.ValueOf(entity).FieldByName(userColumn.ColumnName).String()
					errorMap[userID] = fmt.Errorf("error creating entity: %v", err)
				} else if nameColumn, ok := d.columnCache["Name"]; ok {
					//get the name
					name := reflect.ValueOf(entity).FieldByName(nameColumn.ColumnName).String()
					errorMap[name] = fmt.Errorf("error creating entity: %v", err)
				} else {
					errorMap[entity.GetId()] = fmt.Errorf("error creating entity: %v", err)
				}
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
	return errorMap
}

func (d *EntityData[T]) Create(entity T) error {
	// json, _ := entity.ToJSON()
	// print("Creating entity: ", string(json))
	result := d.GetNewDatabaseSession().Create(&entity)
	//if we have a conflicting ID
	if result.Error != nil {
		entity.SetId("")
		result = d.GetNewDatabaseSession().Create(&entity)
		if result.Error != nil {
			log.Printf("error creating %s: %s", entity.GetId(), result.Error.Error())
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
			newUpdates[upd.Field][upd.ID] = *upd.NewValue
		} else if upd.AlterValue != nil {
			parsed, err := strconv.ParseFloat(*upd.AlterValue, 64)
			if err != nil {
				errorMap[upd.ID] = fmt.Errorf("failed to parse alter value '%s' for field %s: %v", *upd.AlterValue, upd.Field, err)
				continue
			}
			if alterUpdates[upd.Field] == nil {
				alterUpdates[upd.Field] = make(map[string]float64)
			}
			alterUpdates[upd.Field][upd.ID] += parsed
		}
	}

	// Helper to convert new value to proper type and return SQL cast type.
	convertNewValue := func(newVal string, fieldType reflect.Type) (interface{}, string, error) {
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
				valueTuples = append(valueTuples, fmt.Sprintf("(?::text, ?::%s)", castType))
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
				valueTuples = append(valueTuples, fmt.Sprintf("(?::text, ?::%s)", castType))
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
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id = ?", id)
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
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id IN ?", ids)
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

			// Try inserting the entity.
			if err := tx.Delete(&zero, "id = ?", id).Error; err != nil {
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
