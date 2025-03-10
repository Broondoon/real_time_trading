package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"log"

	"github.com/google/uuid"
)

type EntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	EntityDataServiceTemp database.EntityDataInterface[TEntity]
}

type NewEntityDataAccessParams[TEntity entity.EntityInterface] struct {
	*NewDatabaseAccessParams //Leave blank for defaults. (Usually fine)
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	//Cheap ignore of sepearation between access and database. Later on, we'd actually likely have a cache between here and the database, but for now, we'll just connect directly.
	//This would actually go in the proper main of the database. Since however we're currently just testing the database, we'll put it here.
	EntityDataServiceTemp database.EntityDataInterface[TEntity]
}

func NewEntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface](params *NewEntityDataAccessParams[TEntity]) EntityDataAccessInterface[TEntity, TInterface] {
	if params.NewDatabaseAccessParams == nil {
		params.NewDatabaseAccessParams = &NewDatabaseAccessParams{}
	}

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

func (d *EntityDataAccess[TEntity, TInterface]) GetByID(id *uuid.UUID) (TInterface, error) {
	entity, err := d.EntityDataServiceTemp.GetByID(id.String())
	if err != nil {
		log.Fatal("Failed to get entity by ID: ", err)
	}
	return interface{}(entity).(TInterface), nil
}

func (d *EntityDataAccess[TEntity, TInterface]) GetAll() (*[]TInterface, error) {
	entities, err := d.EntityDataServiceTemp.GetAll()
	if err != nil {
		log.Fatal("Failed to get all entities: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}
func (d *EntityDataAccess[TEntity, TInterface]) GetByIDs(ids []*uuid.UUID) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
	entities, err := d.EntityDataServiceTemp.GetByForeignID(foreignIDColumn, foreignID)
	if err != nil {
		log.Fatal("Failed to get entities by ForeignKey: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface]) CreateBulk(entities *[]TInterface) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	return entity, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) Update(entity TInterface) error {
	err := d.EntityDataServiceTemp.Update(*entity.GetUpdates())
	if len(err) > 0 {
		for _, e := range err {
			log.Fatal("Failed to update entity: ", e)
		}
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface]) UpdateBulk(entities *[]TInterface) (map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface]) Delete(id *uuid.UUID) error {
	err := d.EntityDataServiceTemp.Delete(id.String())
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface]) DeleteBulk(ids []*uuid.UUID) (map[string]int, error) {
	panic("implement me") // TODO: Implement
}
