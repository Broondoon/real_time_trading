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
	GetByIDs(IDs []string) (*[]T, error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, error)
	GetAll() (*[]T, error)
	Create(entity T) error
	CreateBulk(entities *[]T) error
	//Update(entity T) error
	//UpdateBulk(entities *[]T) error
	Delete(ID string) error
	DeleteBulk(IDs []string) error
	Exists(ID string) (bool, error)

	//I need a safe updater for numerical values... we can't pass it the updated entity, we have to pass it the values to change the fields by.
	Update([]*entity.EntityUpdateData) error
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
	for fieldName, fieldSchema := range sch.FieldsByName {
		ed.columnCache[fieldName] = columnCacheEntry{
			ColumnName: fieldSchema.DBName,
			FieldType:  fieldSchema.FieldType,
		}
	}

	return ed
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

// This needs the table column names, whihc is a little diffrent
func (d *EntityData[T]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error) {
	var entities []T
	results := d.GetDatabaseSession().Find(&entities, foreignIDColumn+" = ?", foreignID)
	if results.Error != nil {
		fmt.Printf("error getting by foreignKey: %s", results.Error.Error())
		return nil, results.Error
	}
	println("Printing ENtities")
	for _, entity := range entities {
		jso, _ := entity.ToJSON()
		println("Entity: ", string(jso))
	}
	return &entities, nil
}

func (d *EntityData[T]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, error) {
	var entities []T
	results := d.GetDatabaseSession().Find(&entities, foreignIDColumn+" IN ?", foreignIDs)
	if results.Error != nil {
		fmt.Printf("error getting by foreignKey: %s", results.Error.Error())
		return nil, results.Error
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

// func (d *EntityData[T]) Update(entity T) error {
// 	updateResult := d.GetDatabaseSession().Save(entity)
// 	if updateResult.Error != nil {
// 		fmt.Printf("error updating %s: %s", entity.GetId(), updateResult.Error.Error())
// 		return updateResult.Error
// 	}
// 	return nil
// }

//	func (d *EntityData[T]) UpdateBulk(entities *[]T) error {
//		updateResult := d.GetDatabaseSession().Save(entities)
//		if updateResult.Error != nil {
//			fmt.Printf("error updating entities: %s", updateResult.Error.Error())
//			return updateResult.Error
//		}
//		return nil
//	}
//
// Generated with assistance of Chat GPT 03-mini-high: https://chatgpt.com/share/67cb6dc5-7cf4-8006-a7cc-b33fa7765051
func (d *EntityData[T]) Update(updates []*entity.EntityUpdateData) error {
	// Maps to hold the aggregated updates.
	// newUpdates: key is field name, value is map from ID to the new value.
	newUpdates := make(map[string]map[string]string)
	// alterUpdates: key is field name, value is map from ID to the cumulative delta.
	alterUpdates := make(map[string]map[string]float64)

	// Process each update.
	for _, upd := range updates {
		if upd.NewValue != nil {
			if _, ok := newUpdates[upd.Field]; !ok {
				newUpdates[upd.Field] = make(map[string]string)
			}
			newUpdates[upd.Field][upd.ID] = *upd.NewValue
		} else if upd.AlterValue != nil {
			parsed, err := strconv.ParseFloat(*upd.AlterValue, 64)
			if err != nil {
				return fmt.Errorf("failed to parse alter value '%s' for ID %s and field %s: %v", *upd.AlterValue, upd.ID, upd.Field, err)
			}
			if _, ok := alterUpdates[upd.Field]; !ok {
				alterUpdates[upd.Field] = make(map[string]float64)
			}
			alterUpdates[upd.Field][upd.ID] += parsed
		}
	}

	// Wrap all updates in a transaction.
	return d.GetNewDatabaseSession().Transaction(func(tx *gorm.DB) error {
		// First, handle new value updates.
		for field, idToNewVal := range newUpdates {
			cacheEntry, ok := d.columnCache[field]
			if !ok {
				return fmt.Errorf("unknown field %s in column cache", field)
			}

			valueTuples := make([]string, 0, len(idToNewVal))
			args := make([]interface{}, 0, len(idToNewVal)*2)

			// For each update, convert the new value based on the field's type.
			for id, newVal := range idToNewVal {
				var converted interface{}
				var castType string

				switch cacheEntry.FieldType.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					i, err := strconv.ParseInt(newVal, 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse new value '%s' as integer for field %s: %v", newVal, field, err)
					}
					converted = i
					castType = "bigint"
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					u, err := strconv.ParseUint(newVal, 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse new value '%s' as unsigned integer for field %s: %v", newVal, field, err)
					}
					converted = u
					castType = "bigint"
				case reflect.Float32, reflect.Float64:
					f, err := strconv.ParseFloat(newVal, 64)
					if err != nil {
						return fmt.Errorf("failed to parse new value '%s' as float for field %s: %v", newVal, field, err)
					}
					converted = f
					castType = "double precision"
				default:
					// Treat as text if not a recognized numeric type.
					converted = newVal
					castType = "text"
				}
				// Build each tuple, casting the ID to UUID and the new value to the proper type.
				valueTuples = append(valueTuples, fmt.Sprintf("(CAST(? AS uuid), CAST(? AS %s))", castType))
				args = append(args, id, converted)
			}

			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = u.new_val
				FROM (VALUES %s) AS u(id, new_val)
				WHERE t.id = u.id
			`, d.tableName, cacheEntry.ColumnName, strings.Join(valueTuples, ", "))

			if err := tx.Exec(query, args...).Error; err != nil {
				return fmt.Errorf("failed new value bulk update for field '%s': %v", field, err)
			}
		}

		// Next, handle alter value updates.
		for field, idToDelta := range alterUpdates {
			cacheEntry, ok := d.columnCache[field]
			if !ok {
				return fmt.Errorf("unknown field %s in column cache", field)
			}

			var castType string
			valueTuples := make([]string, 0, len(idToDelta))
			args := make([]interface{}, 0, len(idToDelta)*2)

			for id, delta := range idToDelta {
				var deltaValue interface{}
				switch cacheEntry.FieldType.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					castType = "bigint"
					deltaValue = int64(delta)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					castType = "bigint"
					deltaValue = uint64(delta)
				case reflect.Float32, reflect.Float64:
					castType = "double precision"
					deltaValue = delta
				default:
					return fmt.Errorf("unsupported numeric field type %s for field %s", cacheEntry.FieldType.Kind(), field)
				}
				// Build each tuple, casting the ID to UUID and the delta to the proper type.
				valueTuples = append(valueTuples, fmt.Sprintf("(CAST(? AS uuid), CAST(? AS %s))", castType))
				args = append(args, id, deltaValue)
			}

			query := fmt.Sprintf(`
				UPDATE %s AS t
				SET %s = t.%s + u.delta
				FROM (VALUES %s) AS u(id, delta)
				WHERE t.id = u.id
			`, d.tableName, cacheEntry.ColumnName, cacheEntry.ColumnName, strings.Join(valueTuples, ", "))

			if err := tx.Exec(query, args...).Error; err != nil {
				return fmt.Errorf("failed alter value bulk update for field '%s': %v", field, err)
			}
		}

		return nil
	})
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

func (d *EntityData[T]) DeleteBulk(ids []string) error {
	// _, err := d.GetByIDs(ids)
	// if err != nil {
	// 	return err
	// }
	var zero T
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id IN ?", ids)
	if deleteResult.Error != nil {
		fmt.Printf("error deleting entities: %s", deleteResult.Error.Error())
		return deleteResult.Error
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
