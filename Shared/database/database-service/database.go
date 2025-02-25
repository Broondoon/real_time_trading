package database

import (
	"Shared/entities/entity"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
		var err error
		for i := 0; i < 10; i++ { // try 10 times
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err == nil {
				d.database = db
				d.SetConnected(true)
				return
			}
			log.Printf("Database not ready yet, retrying... (%d/10)", i+1)
			time.Sleep(2 * time.Second)
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
	GetAll() (*[]T, error)
	Create(entity T) error
	Update(entity T) error
	Delete(ID string) error
	Exists(ID string) (bool, error)
}

type EntityData[T entity.EntityInterface] struct {
	PostGresDatabaseInterface
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

func (d *EntityData[T]) GetAll() (*[]T, error) {
	var entities []T
	d.GetDatabaseSession().Find(&entities)
	return &entities, nil
}

func (d *EntityData[T]) Create(entity T) error {
	json, _ := entity.ToJSON()
	print("Creating entity: ", string(json))
	candidateID := entity.GetId()
	if candidateID == "" {
		candidateID = generateRandomID()
	}
	for {
		result, err := d.Exists(candidateID)
		if err != nil {
			fmt.Printf("error checking existing: %s", err.Error())
			return err
		}

		if !result {
			break
		}

		candidateID = generateRandomID()
	}

	entity.SetId(candidateID)
	createResult := d.GetDatabaseSession().Create(entity)

	if createResult.Error != nil {
		fmt.Printf("error creating %s: %s", entity.GetId(), createResult.Error.Error())
		return createResult.Error
	}
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
	_, err := d.GetByID(id)
	if err != nil {
		return err
	}
	var zero T
	deleteResult := d.GetDatabaseSession().Delete(&zero, "id = ?", id)
	if deleteResult.Error != nil {
		fmt.Printf("error deleting %s: %s", id, deleteResult.Error.Error())
		return deleteResult.Error
	}
	return nil
}
