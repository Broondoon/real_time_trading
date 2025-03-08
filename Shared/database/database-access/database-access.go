package databaseAccess

import (
	"Shared/entities/entity"
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
	GetByID(id string) (TInterface, error)
	GetAll() (*[]TInterface, error)
	GetByIDs(ids []string) (*[]TInterface, error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error)
	Create(entity TInterface) (TInterface, error)
	CreateBulk(entities *[]TInterface) error
	Update(entity TInterface) error
	Delete(id string) error
}
