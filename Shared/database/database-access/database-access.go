package databaseAccess

import (
	"Shared/database/database-service"
	"Shared/entities/entity"
	"Shared/network"
	"fmt"
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
	GetByID(id string) (TInterface, error)
	GetAll() (*[]TInterface, error)
	GetByIDs(ids []string) (*[]TInterface, error)
	GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error)
	Create(entity TInterface) (TInterface, error)
	Update(entity TInterface) error
	Delete(id string) error
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

func (d *EntityDataAccess[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	err := d.EntityDataServiceTemp.Create(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to create entity: ", err)
	}
	return entity, nil
}

func (d *EntityDataAccess[TEntity, TInterface]) Update(entity TInterface) error {
	err := d.EntityDataServiceTemp.Update(interface{}(entity).(TEntity))
	if err != nil {
		log.Fatal("Failed to update entity: ", err)
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

type EntityDataAccessHTTP[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	_client      network.HttpClientInterface
	PostRoute    string
	GetRoute     string
	PutRoute     string
	DeleteRoute  string
	DefaultRoute string
	Parser       func([]byte) (TEntity, error)
	ParserList   func([]byte) (*[]TEntity, error)
}

type NewEntityDataAccessHTTPParams[TEntity entity.EntityInterface] struct {
	*NewDatabaseAccessParams // leave nil for default. usually fine.
	Client                   network.HttpClientInterface
	PostRoute                string
	GetRoute                 string
	PutRoute                 string
	DeleteRoute              string
	DefaultRoute             string
	Parser                   func([]byte) (TEntity, error)
	ParserList               func([]byte) (*[]TEntity, error)
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
		Parser:                      params.Parser,
		ParserList:                  params.ParserList,
	}
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Connect() {

}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Disconnect() {

}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetByID(id string) (TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute+"/"+id, nil)
	if err != nil {
		var zero TInterface
		return zero, err
		log.Fatal("Failed to get entity by ID: ", err)
	}
	entity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to unmarshal entity: ", err)
		return zero, err
		log.Fatal("Failed to unmarshal entity: ", err)
	}
	return interface{}(entity).(TInterface), nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetAll() (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute, nil)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to get all entities: ", err)
		return &zero, err
		log.Fatal("Failed to get all entities: ", err)
	}
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
		log.Fatal("Failed to unmarshal entities: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetByIDs(ids []string) (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{"ids": strings.Join(ids, ",")}
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to get entities by IDs: ", err)
		return &zero, err
		log.Fatal("Failed to get entities by IDs: ", err)
	}
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
		log.Fatal("Failed to unmarshal entities: ", err)
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
		fmt.Printf("[DEBUG] GetRoute was empty, set to DefaultRoute: %s\n", d.DefaultRoute)
	}
	queryParams := map[string]string{"foreignKey": foreignIDColumn, "id": foreignID}
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero []TInterface
		fmt.Printf("[DEBUG] Failed to get entities by foreignKey: %v\n", err)
		return &zero, err
		log.Fatal("Failed to get entities by foreignKey: ", err)
	}
	fmt.Printf("[DEBUG] Received JSON response: %s\n", string(jsonBytes))
	entities, err := d.ParserList(jsonBytes)

	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
		log.Fatal("Failed to unmarshal entities: ", err)
	}
	fmt.Printf("[DEBUG] Parsed entities: %v\n", *entities)
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Post(d.PostRoute, entity)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to create entity: ", err)
		return zero, err
		log.Fatal("Failed to create entity: ", err)
	}
	newEntity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to unmarshal entity: ", err)
		return zero, err
		log.Fatal("Failed to unmarshal entity: ", err)
	}
	return interface{}(newEntity).(TInterface), nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Update(entity TInterface) error {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	_, err := d._client.Put(d.PutRoute, entity)
	if err != nil {
		fmt.Println("Failed to update entity: ", err)
		return err
		log.Fatal("Failed to update entity: ", err)
	}
	return nil
}

func (d *EntityDataAccessHTTP[TEntity, TInterface]) Delete(id string) error {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	_, err := d._client.Delete(d.DeleteRoute + "/" + id)
	if err != nil {
		fmt.Println("Failed to delete entity: ", err)
		return err
		log.Fatal("Failed to delete entity: ", err)
	}
	return nil
}
