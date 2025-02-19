package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"log"
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
	GetByID(id string) TInterface
	GetAll() *[]TInterface
	GetByIDs(ids []string) *[]TInterface
	Create(entity TInterface)
	Update(entity TInterface)
	Delete(id string)
}

type EntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	EntityDataServiceTemp database.EntityDataInterface[TEntity]
}

type NewEntityDataAccessParams[TEntity entity.EntityInterface] struct {
	*NewDatabaseAccessParams
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	//Cheap ignore of sepearation between access and database. Later on, we'd actually likely have a cache between here and the database, but for now, we'll just connect directly.
	//This would actually go in the proper main of the database. Since however we're currently just testing the database, we'll put it here.
	EntityDataServiceTemp database.EntityDataInterface[TEntity]
}

func NewEntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface](params *NewEntityDataAccessParams[TEntity]) EntityDataAccessInterface[TEntity, TInterface] {
	return &EntityDataAccess[TEntity, TInterface]{
		BaseDatabaseAccessInterface: NewBaseDatabaseAccess(params.NewDatabaseAccessParams),
		EntityDataServiceTemp:       params.EntityDataServiceTemp,
	}
}

func (d *EntityDataAccess[TEntity, TInterface]) Connect() {
	d.EntityDataServiceTemp.Connect()
}

func (d *EntityDataAccess[TEntity, TInterface]) Disconnect() {
	d.EntityDataServiceTemp.Disconnect()
}

func (d *EntityDataAccess[TEntity, TInterface]) GetByID(id string) TInterface {
	entity, err := d.EntityDataServiceTemp.GetByID(id)
	if err != nil {
		log.Fatal("Failed to get entity by ID: ", err)
	}
	return interface{}(entity).(TInterface)
}

func (d *EntityDataAccess[TEntity, TInterface]) GetAll() *[]TInterface {
	entities, err := d.EntityDataServiceTemp.GetAll()
	if err != nil {
		log.Fatal("Failed to get all entities: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted
}

func (d *EntityDataAccess[TEntity, TInterface]) GetByIDs(ids []string) *[]TInterface {
	entities, err := d.EntityDataServiceTemp.GetByIDs(ids)
	if err != nil {
		log.Fatal("Failed to get entities by IDs: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted
}

func (d *EntityDataAccess[TEntity, TInterface]) Create(entity TInterface) {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
}

func (d *EntityDataAccess[TEntity, TInterface]) Update(entity TInterface) {
	err := d.EntityDataServiceTemp.Update(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to update entity: ", err)
	}
}

func (d *EntityDataAccess[TEntity, TInterface]) Delete(id string) {
	err := d.EntityDataServiceTemp.Delete(id)
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
}
