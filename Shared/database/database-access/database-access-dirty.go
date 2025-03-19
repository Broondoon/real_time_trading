package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"log"

	"github.com/google/uuid"
)

// This is a dirty implementation that combines the Database access and the database service.
// Usually the access and service would be on seperate serivces and the access would be a client to the service.
// Here we actually have the service in the access. It's used only for the matching engine.
type EntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface, TDatabase any] struct {
	BaseDatabaseAccessInterface
	EntityDataServiceTemp database.EntityDataInterface[TEntity, TDatabase]
}

type NewEntityDataAccessParams[TEntity entity.EntityInterface, TDatabase any] struct {
	*NewDatabaseAccessParams //Leave blank for defaults. (Usually fine)
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	//Cheap ignore of sepearation between access and database. Later on, we'd actually likely have a cache between here and the database, but for now, we'll just connect directly.
	//This would actually go in the proper main of the database. Since however we're currently just testing the database, we'll put it here.
	EntityDataServiceTemp database.EntityDataInterface[TEntity, TDatabase] //This is the service that actually connects to the database. It's a dirty implementation, but it's fine for now.
}

func NewEntityDataAccess[TEntity entity.EntityInterface, TInterface entity.EntityInterface, TDatabase any](params *NewEntityDataAccessParams[TEntity, TDatabase]) EntityDataAccessInterface[TEntity, TInterface] {
	if params.NewDatabaseAccessParams == nil {
		params.NewDatabaseAccessParams = &NewDatabaseAccessParams{}
	}

	return &EntityDataAccess[TEntity, TInterface, TDatabase]{
		BaseDatabaseAccessInterface: NewBaseDatabaseAccess(params.NewDatabaseAccessParams),
		EntityDataServiceTemp:       params.EntityDataServiceTemp,
	}
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Connect() {
	d.EntityDataServiceTemp.Connect()
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Disconnect() {
	d.EntityDataServiceTemp.Disconnect()
}

// GetByID gets an entity by its ID
func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByID(id *uuid.UUID) (TInterface, error) {
	entity, err := d.EntityDataServiceTemp.GetByID(id.String())
	if err != nil {
		log.Fatal("Failed to get entity by ID: ", err)
	}
	//convert the entity to the interface
	return interface{}(entity).(TInterface), nil
}

// GetAll gets all entities of a type/table
func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetAll() (*[]TInterface, error) {
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
func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByIDs(ids []*uuid.UUID) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
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

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) CreateBulk(entities *[]TInterface) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
	// bulkEntities := make([]TEntity, len(*entities))
	// if len(*entities) == 0 {
	// 	return nil, nil, nil
	// }
	// for i, e := range *entities {
	// 	bulkEntities[i] = interface{}(e).(TEntity)
	// }
	// errors := make(map[string]int)
	// errMap := d.EntityDataServiceTemp.CreateBulk(&bulkEntities)
	// if _, ok := errMap["transaction"]; ok {
	// 	return nil, nil, errMap["transaction"]
	// }
	// for k := range errMap {
	// 	errors[k] = 500
	// }
	// converted := make([]TInterface, len(*entities))
	// for i, e := range *entities {
	// 	converted[i] = interface{}(e).(TInterface)
	// }
	// return &converted, errors, nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Create(entity TInterface) (TInterface, error) {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	return entity, nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Update(entity TInterface) error {
	err := d.EntityDataServiceTemp.Update(*entity.GetUpdates())
	if len(err) > 0 {
		for _, e := range err {
			log.Fatal("Failed to update entity: ", e)
		}
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) UpdateBulk(entities *[]TInterface) (map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Delete(id *uuid.UUID) error {
	err := d.EntityDataServiceTemp.Delete(id.String())
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) DeleteBulk(ids []*uuid.UUID) (map[string]int, error) {
	panic("implement me") // TODO: Implement
}
