package databaseAccess

import (
	"Shared/entities/entity"

	"github.com/google/uuid"
)

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

type EntityDataAccessInterface[TEntity entity.EntityInterface, TInterface entity.EntityInterface] interface {
	DatabaseAccessInterface
	GetByID(id *uuid.UUID) (TInterface, error)
	GetAll() (*[]TInterface, error)
	GetByIDs(ids []*uuid.UUID) (*[]TInterface, map[string]int, error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error)
	GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, map[string]int, error)
	Create(entity TInterface) (TInterface, error)
	CreateBulk(entities *[]TInterface) (*[]TInterface, map[string]int, error)
	Update(entity TInterface) error
	UpdateBulk(entities *[]TInterface) (map[string]int, error)
	Delete(id *uuid.UUID) error
	DeleteBulk(ids []*uuid.UUID) (map[string]int, error)
}
