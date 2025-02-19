package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"Shared/network"
	"encoding/json"
	"log"
	"strings"
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
	Create(entity TInterface) TInterface
	Update(entity TInterface)
	Delete(id string)
}

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

func (d *EntityDataAccess[TEntity, TInterface]) Create(entity TInterface) TInterface {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	return entity
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

type EntityDataAccessHTTP[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	_client      network.HttpClientInterface
	PostRoute    string
	GetRoute     string
	PutRoute     string
	DeleteRoute  string
	DefaultRoute string
}

type NewEntityDataAccessHTTPParams[TEntity entity.EntityInterface] struct {
	*NewDatabaseAccessParams // leave nil for default. usually fine.
	Client                   network.HttpClientInterface
	PostRoute                string
	GetRoute                 string
	PutRoute                 string
	DeleteRoute              string
	DefaultRoute             string
}

func NewEntityDataAccessHTTP[TEntity entity.EntityInterface, TInterface entity.EntityInterface](params *NewEntityDataAccessHTTPParams[TEntity]) EntityDataAccessInterface[TEntity, TInterface] {
	if params.NewDatabaseAccessParams == nil {
		params.NewDatabaseAccessParams = &NewDatabaseAccessParams{}
	}
	return &EntityDataAccessHTTP[TEntity, TInterface]{
		BaseDatabaseAccessInterface: NewBaseDatabaseAccess(params.NewDatabaseAccessParams),
		_client:                     params.Client,
		PostRoute:                   params.PostRoute,
		GetRoute:                    params.GetRoute,
		PutRoute:                    params.PutRoute,
		DeleteRoute:                 params.DeleteRoute,
		DefaultRoute:                params.DefaultRoute,
	}
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Connect() {

}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Disconnect() {

}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetByID(id string) TInterface {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute+"/"+id, nil)
	if err != nil {
		log.Fatal("Failed to get entity by ID: ", err)
	}
	var entity TEntity
	err = json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		log.Fatal("Failed to unmarshal entity: ", err)
	}
	return interface{}(entity).(TInterface)
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetAll() *[]TInterface {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute, nil)
	if err != nil {
		log.Fatal("Failed to get all entities: ", err)
	}
	var entities []TEntity
	err = json.Unmarshal(jsonBytes, &entities)
	if err != nil {
		log.Fatal("Failed to unmarshal entities: ", err)
	}
	converted := make([]TInterface, len(entities))
	for i, e := range entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetByIDs(ids []string) *[]TInterface {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{"ids": strings.Join(ids, ",")}
	jsonBytes, err := d._client.Get(d.PostRoute, queryParams)
	if err != nil {
		log.Fatal("Failed to get entities by IDs: ", err)
	}
	var entities []TEntity
	err = json.Unmarshal(jsonBytes, &entities)
	if err != nil {
		log.Fatal("Failed to unmarshal entities: ", err)
	}
	converted := make([]TInterface, len(entities))
	for i, e := range entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Create(entity TInterface) TInterface {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Post(d.PostRoute, entity)
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	var newEntity TEntity
	err = json.Unmarshal(jsonBytes, &newEntity)
	if err != nil {
		log.Fatal("Failed to unmarshal entity: ", err)
	}
	return interface{}(newEntity).(TInterface)
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Update(entity TInterface) {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	_, err := d._client.Put(d.PutRoute, entity)
	if err != nil {
		log.Fatal("Failed to update entity: ", err)
	}
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Delete(id string) {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	_, err := d._client.Delete(d.DeleteRoute + "/" + id)
	if err != nil {
		log.Fatal("Failed to delete entity: ", err)
	}
}
