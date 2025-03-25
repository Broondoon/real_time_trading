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

type PostGresDatabase struct {
	BaseDatabaseInterface
	database *gorm.DB
}

type NewPostGresDatabaseParams struct {
	*NewBaseDatabaseParams // leave nil for default
}

func NewPostGresDatabase(params *NewPostGresDatabaseParams) DatabaseInterface[*gorm.DB] {
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

type columnCacheEntry struct {
	ColumnName string
	FieldType  reflect.Type
}

type PostGresEntityData[T entity.EntityInterface] struct {
	DatabaseInterface[*gorm.DB]
	tableName   string
	columnCache map[string]columnCacheEntry
	// *gorm.DB //note, this allows us to treat this as a gorm.DB WITHIN the EntityData struct. This is not exposed as part of the interface, and thus cannot be used like this with the interface.
}

type NewPostGresEntityDataParams struct {
	*NewPostGresDatabaseParams                             // leave nil for default, not used if existing is provided
	Existing                   DatabaseInterface[*gorm.DB] // leave nil for new database connection
}

func NewPostGresEntityData[T entity.EntityInterface](params *NewPostGresEntityDataParams) EntityDataInterface[T, *gorm.DB] {
	if params.NewPostGresDatabaseParams == nil {
		params.NewPostGresDatabaseParams = &NewPostGresDatabaseParams{}
	}

	if params.Existing == nil {
		params.Existing = NewPostGresDatabase(params.NewPostGresDatabaseParams)
	}

	// Create an instance with an empty column cache.
	ed := &PostGresEntityData[T]{
		DatabaseInterface: params.Existing,
		columnCache:       make(map[string]columnCacheEntry),
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
	//log.Println(ed.tableName, " ID Data type: ", ed.columnCache["ID"].FieldType)

	return ed
}

func convertID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse id %s: %v", id, err)
	}
	return uid, nil
}

func convertIDs(ids []string, errorList map[string]error) ([]uuid.UUID, map[string]error) {
	existingIds := make(map[string]bool)
	uids := make([]uuid.UUID, len(ids))
	i := 0
	for _, id := range ids {
		if _, ok := existingIds[id]; ok {
			continue
		}
		uid, err := convertID(id)
		if err != nil {
			errorList[id] = err
			continue
		}
		uids[i] = uid
		existingIds[id] = true
		i++
	}
	return uids, errorList
}

func (d *PostGresEntityData[T]) PrintOutEntities() {
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

func (d *PostGresEntityData[T]) GetByID(id string) (T, error) {
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
		//d.PrintOutEntities()
		return zero, result.Error
	}
	if result.Error != nil {
		log.Printf("error getting: %s", result.Error.Error())
		//d.PrintOutEntities()
		return zero, result.Error
	}
	return ent, nil
}

func (d *PostGresEntityData[T]) GetByIDs(ids []string) (*[]T, map[string]error) {
	if len(ids) == 0 {
		return nil, map[string]error{"transaction": errors.New("no ids provided")}
	}
	var entities []T
	errorList := make(map[string]error)
	uids, errorList := convertIDs(ids, errorList)
	if len(uids) == 0 {
		return nil, errorList
	}

	results := d.GetDatabaseSession().Find(&entities, "id IN ?", uids)
	if results.Error != nil {
		errorList["transaction"] = results.Error
		log.Printf("error getting by ids: %s", results.Error.Error())
		//d.PrintOutEntities()

		return nil, errorList
	}
	//get all ids in ids that are not in entities
	idsFound := make(map[string]bool)
	for _, ent := range entities {
		idsFound[ent.GetIdString()] = true
	}
	for _, id := range ids {
		if val, ok := idsFound[id]; !ok && !val {
			errorList[id] = gorm.ErrRecordNotFound
			//d.PrintOutEntities()
		}
	}

	return &entities, errorList
}

// This needs the table column names, whihc is a little diffrent
func (d *PostGresEntityData[T]) GetByForeignID(foreignIDKey string, foreignID string) (*[]T, error) {
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
		//log.Println("avalaible columns: ", strings.Join(columns, ", "))
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
		//d.PrintOutEntities()
		return nil, results.Error
	}
	return &entities, nil
}

func (d *PostGresEntityData[T]) GetByForeignIDBulk(foreignIDKey string, foreignIDs []string) (*[]T, map[string]error) {
	return d.GetByFilteredForeignIDBulk(foreignIDKey, foreignIDs, "", "")
}

func (d *PostGresEntityData[T]) GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterCol string, filterVal string) (*[]T, map[string]error) {
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
	errorList := make(map[string]error)
	foreignIDColumn, ok := d.columnCache[foreignIDKey]
	if !ok {
		errorList["transaction"] = fmt.Errorf("foreign key column %s not found", foreignIDKey)
		log.Printf("error getting by foreignKey: %s", errorList["transaction"].Error())
		return nil, errorList
	}
	var (
		results   *gorm.DB
		condition string
		args      []interface{}
		uids      []uuid.UUID
	)
	condition = foreignIDColumn.ColumnName + " IN ?"
	if strings.Contains(foreignIDColumn.ColumnName, "_id") || foreignIDColumn.ColumnName == "id" {
		uids, errorList = convertIDs(foreignIDs, errorList)
		if len(uids) == 0 {
			return nil, errorList
		}
		args = append(args, uids)
	} else {
		args = append(args, foreignIDs)
	}

	if filterCol != "" {
		condition += " AND " + filterCol + " = ?"
		if strings.Contains(filterCol, "_id") {
			filterUid, err := uuid.Parse(filterVal)
			if err != nil {
				errorList["transaction"] = err
				log.Printf("error getting by foreignKey: %s", err.Error())
				return nil, errorList
			}
			args = append(args, filterUid)
		} else {
			args = append(args, filterVal)
		}
	}
	results = d.GetDatabaseSession().Where(condition, args...).Find(&entities)

	if results.Error != nil {
		errorList["transaction"] = results.Error
		log.Printf("error getting by foreignKey: %s", results.Error.Error())
		//	d.PrintOutEntities()
		return nil, errorList
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
				//log.Println("Foreign ID Found: ", actual.String())
			} else {
				// Possibly store an empty string or skip
				continue
			}
		default:
			foreignID := fieldVal.String()
			//log.Println("Foreign ID Found: ", foreignID)
			idsFound[foreignID] = true
		}
	}
	for _, id := range foreignIDs {
		//log.Println("Checking for foreign ID: ", id)
		if val, ok := idsFound[id]; !ok || !val {
			//d.PrintOutEntities()
			errorList[id] = gorm.ErrRecordNotFound
		}
	}
	return &entities, errorList
}

func (d *PostGresEntityData[T]) GetAll() (*[]T, error) {
	var entities []T
	result := d.GetDatabaseSession().Find(&entities)
	if result.Error != nil {
		log.Printf("error getting all: %s", result.Error.Error())
		return nil, result.Error
	}
	return &entities, nil
}

func (d *PostGresEntityData[T]) CreateBulk(entities *[]T) map[string]error {
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

func (d *PostGresEntityData[T]) Create(ent T) error {
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

func (d *PostGresEntityData[T]) Update(updates []*entity.EntityUpdateData) map[string]error {
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
			uid, err := uuid.Parse(strings.TrimSpace(newVal))
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
			//	println("testing for ID column. Field type is ", fieldType.Kind().String())
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
				uid, err := uuid.Parse(strings.TrimSpace(id))
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
				uid, err := uuid.Parse(strings.TrimSpace(id))
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
			uid, err := uuid.Parse(strings.TrimSpace(id))
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
			uid, err := uuid.Parse(strings.TrimSpace(id))
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

func (d *PostGresEntityData[T]) Delete(id string) error {
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

func (d *PostGresEntityData[T]) DeleteBulk(ids []string) map[string]error {
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
func (d *PostGresEntityData[T]) getGormSchema() (*schema.Schema, error) {
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
func (d *PostGresEntityData[T]) getTableName() (string, error) {
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
func (d *PostGresEntityData[T]) getColumnName(fieldName string) (string, error) {
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
