package database

import (
	"Shared/entities/entity"
	"log"
	"os"
)

// BaseDatabase and related types.
// This should be used as the basis for any type of database, not just postgres.
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

// basic constructor for a database. This should be used as the basis for any type of database, not just postgres.
// This should not be called outside of this package.
// This must be called by anything that implements the DatabaseInterface in it's own constructor.
func NewBaseDatabase(params *NewBaseDatabaseParams) BaseDatabaseInterface {
	if params.DATABASE_URL_ENV == "" {
		params.DATABASE_URL_ENV = "DATABASE_URL"
	}

	return &BaseDatabase{
		DatabaseURLEnv: params.DATABASE_URL_ENV,
		Connected:      false,
	}
}

// get the path to the database container itself.
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

// DatabaseInterface and PostGresDatabaseInterface.s
type DatabaseInterface[TDatabase any] interface {
	BaseDatabaseInterface
	Connect()
	Disconnect()
	GetDatabaseSession() TDatabase
	GetNewDatabaseSession() TDatabase
}

type EntityDataInterface[T entity.EntityInterface, TDatabase any] interface {
	DatabaseInterface[TDatabase]
	GetByID(ID string) (T, error)
	GetByIDs(IDs []string) (*[]T, map[string]error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]T, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]T, map[string]error)
	GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterCol string, filterVal string) (*[]T, map[string]error)
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
