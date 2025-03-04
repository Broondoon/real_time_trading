package databaseAccess

import (
	"Shared/entities/entity"
	"Shared/network"
	"fmt"
	"strings"
)

type EntityDataAccessClient[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	_client      network.ClientInterface
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
	Client                   network.ClientInterface
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
	return &EntityDataAccessClient[TEntity, TInterface]{
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

func (d *EntityDataAccessClient[TEntity, TInterface]) Connect() {

}

func (d *EntityDataAccessClient[TEntity, TInterface]) Disconnect() {

}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByID(id string) (TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute+"/"+id, nil)
	if err != nil {
		var zero TInterface
		return zero, err
	}
	entity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	return interface{}(entity).(TInterface), nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetAll() (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute, nil)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to get all entities: ", err)
		return &zero, err
	}
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByIDs(ids []string) (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{"ids": strings.Join(ids, ",")}
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to get entities by IDs: ", err)
		return &zero, err
	}
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
	println("Getting by foreign ID")
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
	}
	fmt.Printf("[DEBUG] Received JSON response: %s\n", string(jsonBytes))
	entities, err := d.ParserList(jsonBytes)

	if err != nil {
		var zero []TInterface
		fmt.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	fmt.Printf("[DEBUG] Parsed entities: %v\n", *entities)
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Post(d.PostRoute, entity)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to create entity: ", err)
		return zero, err
	}
	newEntity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		fmt.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	return interface{}(newEntity).(TInterface), nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) Update(entity TInterface) error {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	_, err := d._client.Put(d.PutRoute, entity)
	if err != nil {
		fmt.Println("Failed to update entity: ", err)
		return err
	}
	return nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) Delete(id string) error {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	_, err := d._client.Delete(d.DeleteRoute + "/" + id)
	if err != nil {
		fmt.Println("Failed to delete entity: ", err)
		return err
	}
	return nil
}
