package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"log"
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

func (d *EntityDataAccess[TEntity, TInterface]) GetByID(id string) (TInterface, error) {
	entity, err := d.EntityDataServiceTemp.GetByID(id)
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
func (d *EntityDataAccess[TEntity, TInterface]) GetByIDs(ids []string) (*[]TInterface, error) {
	entities, err := d.EntityDataServiceTemp.GetByIDs(ids)
	if err != nil {
		log.Fatal("Failed to get entities by Ids: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
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

func (d *EntityDataAccess[TEntity, TInterface]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, error) {
	entities, err := d.EntityDataServiceTemp.GetByForeignIDBulk(foreignIDColumn, foreignIDs)
	if err != nil {
		log.Fatal("Failed to get entities by ForeignKey: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) CreateBulk(entities *[]TInterface) (*[]TInterface, error) {
	entitiesTemp := make([]TEntity, len(*entities))
	for i, e := range *entities {
		entitiesTemp[i] = interface{}(e).(TEntity)
	}
	err := d.EntityDataServiceTemp.CreateBulk(&entitiesTemp)
	if err != nil {
		log.Fatal("Failed to create entities: ", err)
	}
	return entities, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	return entity, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) Update(entity TInterface) error {
	err := d.EntityDataServiceTemp.Update(entity.GetUpdates())
	if err != nil {
		log.Fatal("Failed to update entity: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface]) UpdateBulk(entities *[]TInterface) error {
	updates := make([]*entity.EntityUpdateData, 0)
	for _, v := range *entities {
		for _, u := range v.GetUpdates() {
			updates = append(updates, u)
		}
	}
	err := d.EntityDataServiceTemp.Update(updates)
	if err != nil {
		log.Fatal("Failed to update entities: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface]) Delete(id string) error {
	err := d.EntityDataServiceTemp.Delete(id)
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface]) DeleteBulk(ids []string) error {
	err := d.EntityDataServiceTemp.DeleteBulk(ids)
	if err != nil {
		log.Fatal("Failed to delete entities: ", err)
	}
	return nil
}
