package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"Shared/objects"
	"errors"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByID(id *uuid.UUID, shardKey string) (TInterface, error) {
	entity, err := d.EntityDataServiceTemp.GetByID(id.String(), shardKey)
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
func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByIDs(ids []*uuid.UUID, shardKeys *[]string) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByForeignID(foreignIDColumn string, foreignID string, shardKey string) (*[]TInterface, error) {
	entities, err := d.EntityDataServiceTemp.GetByForeignID(foreignIDColumn, foreignID, shardKey)
	if err != nil {
		log.Fatal("Failed to get entities by ForeignKey: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByPairedID(idColumn1 string, idColumn2 string, ids objects.Pair) (TInterface, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByPairedIDBulk(idColumn1 string, idColumn2 string, ids *[]objects.Pair) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByFilteredForeignIDBulk(foreignIDKey string, foreignIDs []string, filterKey string, filterVal string, shardKeys *[]string) (*[]TInterface, map[string]int, error) {
	panic("implement me") // TODO: Implement
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string, shardKeys *[]string) (*[]TInterface, map[string]int, error) {
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
	updates := *entity.GetUpdates()
	entity.ClearUpdates()
	err := d.EntityDataServiceTemp.Update(*entity.GetUpdates())
	if len(err) > 0 {
		for _, singleErr := range err {
			log.Println("Failed to update entity: ", singleErr)
			break
		}
		entity.SetUpdates(&updates)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) UpdateBulk(entities *[]TInterface) (map[string]int, error) {
	changesPerEntity := make(map[string][]*entity.EntityUpdateData)
	entitiesById := make(map[string]TInterface)
	//We convert the list of updates to the right format for the bulk request.
	var interfaces []*entity.EntityUpdateData
	for _, v := range *entities {
		entitiesById[v.GetIdString()] = v
		changesPerEntity[v.GetIdString()] = *v.GetUpdates()
		interfaces = append(interfaces, *v.GetUpdates()...)
		v.ClearUpdates()
	}
	//We don't actually have a put bulk request, because it's always just a list of updates. So we just send a put request with the updates. This one just has updates for multiple entities, and reutrns errors for mulitple entities.
	err := d.EntityDataServiceTemp.Update(interfaces)
	errorList := make(map[string]int)
	for id, err := range err {
		log.Println("Failed to update entity: ", id, " with Error: ", err, ". Re-adding Updates to entity.")
		//We put the updates back in the entity, so they can be tried again.
		tempUpdates := changesPerEntity[id]
		entitiesById[id].SetUpdates(&tempUpdates)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorList[id] = 404
		} else {
			errorList[id] = 500
		}
	}
	return errorList, nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) Delete(id *uuid.UUID, shardKey string) error {
	err := d.EntityDataServiceTemp.Delete(id.String(), shardKey)
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
	return nil
}

func (d *EntityDataAccess[TEntity, TInterface, TDatabase]) DeleteBulk(ids []*uuid.UUID, shardKeys *[]string) (map[string]int, error) {
	panic("implement me") // TODO: Implement
}
