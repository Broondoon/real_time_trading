package databaseAccess

import (
	"Shared/entities/entity"
	"Shared/objects"

	"github.com/google/uuid"
)

// Basic stuff for database access not specific to entities. Currently blank as we don't have any generic database access stuff, but might need it in the future.
type BaseDatabaseAccessInterface interface {
}

type BaseDatabaseAccess struct {
}

type DatabaseAccessInterface interface {
	Connect()
	Disconnect()
}

type NewDatabaseAccessParams struct {
}

func NewBaseDatabaseAccess(params *NewDatabaseAccessParams) BaseDatabaseAccessInterface {
	return &BaseDatabaseAccess{}
}

// If you create a new database access, you must implement this interface.
// For the record, this is essentially access to a specific table on the DB, which is currently mapped to a specific entity.
// This is a generic interface for accessing the database. It's used to interact with the database on a generic level.
// It's expected to pass through the Network/client to the Network/network
// Implmentaitons should handle communication with the service the DB is hosted on, and parsing the returned results
// The type constirctions above are [base entity type, such as User, and the interface for that entity, such as UserInterface]
type EntityDataAccessInterface[TEntity entity.EntityInterface, TInterface entity.EntityInterface] interface {
	DatabaseAccessInterface
	GetByID(id *uuid.UUID, shardKey string) (TInterface, error) //Get an entity by its ID. Mapped to
	GetAll() (*[]TInterface, error)
	GetByIDs(ids []*uuid.UUID, shardKeys *[]string) (*[]TInterface, map[string]int, error)
	GetByPairedID(idColumn1 string, idColumn2 string, ids objects.Pair) (TInterface, error)
	GetByPairedIDBulk(idColumn1 string, idColumn2 string, ids *[]objects.Pair) (*[]TInterface, map[string]int, error)
	GetByForeignID(foreignIDColumn string, foreignID string, shardKey string) (*[]TInterface, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string, shardKey *[]string) (*[]TInterface, map[string]int, error)
	GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterKey string, filterVal string, shardKey *[]string) (*[]TInterface, map[string]int, error)
	Create(entity TInterface) (TInterface, error)
	CreateBulk(entities *[]TInterface) (*[]TInterface, map[string]int, error)
	Update(entity TInterface) error
	UpdateBulk(entities *[]TInterface) (map[string]int, error)
	Delete(id *uuid.UUID, shardKey string) error
	DeleteBulk(ids []*uuid.UUID, shardKey *[]string) (map[string]int, error)
}
