package databaseAccess

import (
	"Shared/entities/entity"
	"Shared/network"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	retriesStr := os.Getenv("HEALTHCHECK_RETRIES")
	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		retries = 10
	}
	intervalStr := os.Getenv("HEALTHCHECK_INTERVAL")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 1
	}
	timeoutStr := os.Getenv("HEALTHCHECK_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		interval = 1
	}
	baseURL := fmt.Sprintf("%s/health", d._client.GetBaseURL())

	for i := 0; i < retries; i++ { // try with converted retries count
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
		resp, err := client.Get(baseURL)
		if err != nil {
			log.Printf("Database not ready yet, retrying... (%d/%d)", i+1, retries)
			time.Sleep(time.Duration(interval) * time.Second)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return
		}
	}
	log.Fatal("Database connection failed after multiple attempts: ", err)

}

func (d *EntityDataAccessClient[TEntity, TInterface]) Disconnect() {

}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByID(id *uuid.UUID) (TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Get(d.GetRoute+"/"+id.String(), nil)
	if err != nil {
		var zero TInterface
		return zero, err
	}
	entity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		log.Println("Failed to unmarshal entity: ", err)
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
		log.Println("Failed to get all entities: ", err)
		return &zero, err
	}
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByIDs(ids []*uuid.UUID) (*[]TInterface, map[string]int, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{}
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	bulkReturn, err := d._client.GetBulk(d.GetRoute, idsStr, queryParams)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to get entities by IDs: ", err)
		return &zero, mapErrs, err
	}
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
	log.Println("Getting by foreign ID")
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
		log.Printf("[DEBUG] GetRoute was empty, set to DefaultRoute: %s\n", d.DefaultRoute)
	}
	queryParams := map[string]string{"foreignKey": foreignIDColumn, "id": foreignID}
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero []TInterface
		log.Printf("[DEBUG] Failed to get entities by foreignKey: %v\n", err)
		return &zero, err
	}
	log.Printf("[DEBUG] Received JSON response: %s\n", string(jsonBytes))
	entities, err := d.ParserList(jsonBytes)

	if err != nil {
		var zero []TInterface
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	log.Printf("[DEBUG] Parsed entities: %v\n", *entities)
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, map[string]int, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{"foreignKey": foreignIDColumn}
	bulkReturn, err := d._client.GetBulk(d.GetRoute, foreignIDs, queryParams)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to get entities by foreignKey: ", err)
		return &zero, mapErrs, err
	}
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) CreateBulk(entitiesList *[]TInterface) (*[]TInterface, map[string]int, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	var interfaces []interface{}
	for _, v := range *entitiesList {
		interfaces = append(interfaces, v)
	}

	bulkReturn, err := d._client.PostBulk(d.PostRoute, interfaces)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to create entities: ", err)
		return &zero, mapErrs, err
	}
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil

}

func (d *EntityDataAccessClient[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	jsonBytes, err := d._client.Post(d.PostRoute, entity)
	if err != nil {
		var zero TInterface
		log.Println("Failed to create entity: ", err)
		return zero, err
	}
	newEntity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		log.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	return interface{}(newEntity).(TInterface), nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) Update(entity TInterface) error {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}

	var updatesInterface []interface{}
	for _, u := range *entity.GetUpdates() {
		updatesInterface = append(updatesInterface, u)
	}

	bulkReturn, err := d._client.Put(d.PutRoute, updatesInterface)
	if err != nil {
		log.Println("Failed to update entity: ", err)
		return err
	}
	if len(bulkReturn.Errors) > 0 {
		log.Println("Failed to update entity: ", bulkReturn.Errors)
		err = fmt.Errorf("Failed to update entity: %v", bulkReturn.Errors)
		return err
	}
	return nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) UpdateBulk(entities *[]TInterface) (map[string]int, error) {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	var interfaces []interface{}
	for _, v := range *entities {
		for _, u := range *v.GetUpdates() {
			interfaces = append(interfaces, u)
		}
	}
	bulkReturn, err := d._client.Put(d.PutRoute, interfaces)
	if err != nil {
		var mapErrs map[string]int
		log.Println("Failed to update entities: ", err)
		return mapErrs, err
	}
	return bulkReturn.Errors, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) Delete(id *uuid.UUID) error {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	_, err := d._client.Delete(d.DeleteRoute + "/" + id.String())
	if err != nil {
		log.Println("Failed to delete entity: ", err)
		return err
	}
	return nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) DeleteBulk(ids []*uuid.UUID) (map[string]int, error) {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, id.String())
	}
	bulkReturn, err := d._client.DeleteBulk(d.DeleteRoute, idsStr)
	if err != nil {
		var mapErrs map[string]int
		log.Println("Failed to delete entities: ", err)
		return mapErrs, err
	}
	return bulkReturn.Errors, nil
}
