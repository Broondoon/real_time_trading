package databaseAccess

import (
	"Shared/entities/entity"
	"Shared/network"
	"Shared/objects"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Database access for entities defined in shared entities, and implmenting the databseAccessInterface
// This is the client that connects to the database service using something that implements the network.ClientInterface
// TEntity is an entity struct type, such as User.
// TInterface is the interface for that entity, such as UserInterface
// Using these, make sure that a Network/network.go.CreateNetworkEntityHandlers is setup on the recieivng service. This will not work without it.
type EntityDataAccessClient[TEntity entity.EntityInterface, TInterface entity.EntityInterface] struct {
	BaseDatabaseAccessInterface
	_client      network.ClientInterface          // The client that connects to the database service
	PostRoute    string                           // The route for POST requests.  Defaults to DefaultRoute if empty.
	GetRoute     string                           // The route for GET requests.  Defaults to DefaultRoute if empty.
	PutRoute     string                           // The route for PUT requests.  Defaults to DefaultRoute if empty.
	DeleteRoute  string                           // The route for DELETE requests.  Defaults to DefaultRoute if empty.
	DefaultRoute string                           // The default route for requests.  Defaults to an environment variable if empty.
	Parser       func([]byte) (TEntity, error)    // The parser for a single entity.  Can't be empty. The parser method is usually found in the entity file in question.
	ParserList   func([]byte) (*[]TEntity, error) //parser for a list of entities.  Can't be empty. The parser method is usually found in the entity file in question.
}

// See the struct above.
type NewEntityDataAccessNetworkParams[TEntity entity.EntityInterface] struct {
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

func NewEntityDataAccessNetwork[TEntity entity.EntityInterface, TInterface entity.EntityInterface](params *NewEntityDataAccessNetworkParams[TEntity]) EntityDataAccessInterface[TEntity, TInterface] {
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

// Connect to DB service. Usually this is just a health check to see if the service is up.
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

	//for number of retries, ping the service healthcheck and wait for timeout. If not, wait interval then try again.
	for i := 0; i < retries; i++ { // try with converted retries count
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
		resp, err := client.Get(baseURL) //Ping the health check route
		if err != nil {
			log.Printf("Database not ready yet, retrying... (%d/%d)", i+1, retries)
			time.Sleep(time.Duration(interval) * time.Second)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return
		}
	}
	//Assume critical error if we can't reach the service.
	log.Fatal("Database connection failed after multiple attempts: ", err)

}

func (d *EntityDataAccessClient[TEntity, TInterface]) Disconnect() {

}

// GetByID gets an entity by its ID
func (d *EntityDataAccessClient[TEntity, TInterface]) GetByID(id *uuid.UUID) (TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	//We attach the ID to the route and send a GET request to the database service.
	jsonBytes, err := d._client.Get(d.GetRoute+"/"+id.String(), nil)
	if err != nil {
		var zero TInterface
		return zero, err
	}
	//We parse the JSON response into the entity struct.
	entity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		log.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	//We convert the entity struct to the interface and return it.
	return interface{}(entity).(TInterface), nil
}

// GetAll gets all entities of a type/table
func (d *EntityDataAccessClient[TEntity, TInterface]) GetAll() (*[]TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	//We send a GET request to the database service to get all entities. By not indicating any ID, we get all entities.
	jsonBytes, err := d._client.Get(d.GetRoute, nil)
	if err != nil {
		var zero []TInterface
		log.Println("Failed to get all entities: ", err)
		return &zero, err
	}
	//We parse the JSON response into a list of entity structs.
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

// GetByIDs gets entities by their IDs. This is a bulk operation.
// It will return all the entities that match the IDs, and a map of any errors that occurred, mapped to their respective IDs.
func (d *EntityDataAccessClient[TEntity, TInterface]) GetByIDs(ids []*uuid.UUID) (*[]TInterface, map[string]int, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	queryParams := map[string]string{} //Empty because we don't use this for anything here.
	idsStr := make([]string, len(ids)) //Convert the UUIDs to strings. Easier to manage here.
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	//We send a GET request to the database service to get entities by their IDs.
	bulkReturn, err := d._client.GetBulk(d.GetRoute, idsStr, queryParams)
	if err != nil {
		//If the entire thing fails, we send back empty values and a generic error.
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to get entities by IDs: ", err)
		return &zero, mapErrs, err
	}
	//We parse the JSON response into a list of entity structs.
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByPairedID(idColumn1 string, idColumn2 string, ids objects.Pair) (TInterface, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	//We attach the ID to the route and send a GET request to the database service.
	queryParams := map[string]string{"IdColumn1": idColumn1, "IdColumn2": idColumn2, "Id1": ids.ID1, "Id2": ids.ID2}
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero TInterface
		return zero, err
	}
	//We parse the JSON response into the entity struct.
	entity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		log.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	//We convert the entity struct to the interface and return it.
	return interface{}(entity).(TInterface), nil
}

func (d *EntityDataAccessClient[TEntity, TInterface]) GetByPairedIDBulk(idColumn1 string, idColumn2 string, ids *[]objects.Pair) (*[]TInterface, map[string]int, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	//We send a GET request to the database service to get entities by their paired IDs.
	queryParams := map[string]string{"IdColumn1": idColumn1, "IdColumn2": idColumn2}
	//We convert the list of paired IDs to the right format for the bulk request.
	var interfaces []string
	for _, v := range *ids {
		interfaces = append(interfaces, v.String())
	}
	//We send a GET request to the database service to get entities by their paired IDs.
	bulkReturn, err := d._client.GetBulk(d.GetRoute, interfaces, queryParams)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to get entities by paired IDs: ", err)
		return &zero, mapErrs, err
	}
	//We parse the JSON response into a list of entity structs.
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil
}

// GetByForeignID gets entities by a foreign key.
func (d *EntityDataAccessClient[TEntity, TInterface]) GetByForeignID(foreignIDColumn string, foreignID string) (*[]TInterface, error) {
	//log.Println("Getting by foreign ID")
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
		//log.Printf("[DEBUG] GetRoute was empty, set to DefaultRoute: %s\n", d.DefaultRoute)
	}
	//For the network handler side, we indicate this is a foreign key search by passing the foreign key column and the foreign key in the query/header.
	queryParams := map[string]string{"foreignKey": foreignIDColumn, "id": foreignID}
	//We send a GET request to the database service to get entities by their foreign key.
	jsonBytes, err := d._client.Get(d.GetRoute, queryParams)
	if err != nil {
		var zero []TInterface
		log.Printf("[DEBUG] Failed to get entities by foreignKey: %v\n", err)
		return &zero, err
	}
	//We parse the JSON response into a list of entity structs.
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, nil
}

// GetByForeignIDBulk gets entities by a foreign key. This is a bulk operation. Try to avoid using this if possible, as it's an O(n^2) operation, unless the DB has a better alogirthm built in.
// It will return all the entities that match the foreign IDs, and a map of any errors that occurred, mapped to their respective foreign IDs.
func (d *EntityDataAccessClient[TEntity, TInterface]) GetByForeignIDBulk(foreignIDColumn string, foreignIDs []string) (*[]TInterface, map[string]int, error) {
	return d.GetByFilteredForeignIDBulk(foreignIDColumn, foreignIDs, "", "")
}
func (d *EntityDataAccessClient[TEntity, TInterface]) GetByFilteredForeignIDBulk(foreignIDColumn string, foreignIDs []string, filterKey string, filterVal string) (*[]TInterface, map[string]int, error) {
	if d.GetRoute == "" {
		d.GetRoute = d.DefaultRoute
	}
	// mark this as a foreign key search by passing the foreign key column name
	queryParams := map[string]string{"foreignKey": foreignIDColumn, "filteredForeignKey": filterKey, "filteredForeignID": filterVal}
	//We send a GET request to the database service to get entities by their foreign key.
	bulkReturn, err := d._client.GetBulk(d.GetRoute, foreignIDs, queryParams)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to get entities by foreignKey: ", err)
		return &zero, mapErrs, err
	}
	//We parse the JSON response into a list of entity structs.
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil
}

// CreateBulk creates entities in bulk. This is a bulk operation.
// It will return all the entities that were created, and a map of any errors that occurred, mapped to their respective entities Unique Identifier. Make sure this unqiue identifier is set before calling this.
func (d *EntityDataAccessClient[TEntity, TInterface]) CreateBulk(entitiesList *[]TInterface) (*[]TInterface, map[string]int, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	//converted to make sure the bulk request is in the right format.
	var interfaces []interface{}
	for _, v := range *entitiesList {
		interfaces = append(interfaces, v)
	}
	//We send a POST request to the database service to create entities in bulk.
	bulkReturn, err := d._client.PostBulk(d.PostRoute, interfaces)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to create entities: ", err)
		return &zero, mapErrs, err
	}
	//We parse the JSON response into a list of entity structs.
	jsonBytes := bulkReturn.Entities
	entities, err := d.ParserList(jsonBytes)
	if err != nil {
		var zero []TInterface
		var mapErrs map[string]int
		log.Println("Failed to unmarshal entities: ", err)
		return &zero, mapErrs, err
	}
	//We convert the list of entity structs to a list of interfaces and return it.
	converted := make([]TInterface, len(*entities))
	for i, e := range *entities {
		converted[i] = interface{}(e).(TInterface)
	}
	return &converted, bulkReturn.Errors, nil

}

// Create creates an entity.
func (d *EntityDataAccessClient[TEntity, TInterface]) Create(entity TInterface) (TInterface, error) {
	if d.PostRoute == "" {
		d.PostRoute = d.DefaultRoute
	}
	//We send a POST request to the database service to create an entity.
	jsonBytes, err := d._client.Post(d.PostRoute, entity)
	if err != nil {
		var zero TInterface
		log.Println("Failed to create entity: ", err)
		return zero, err
	}
	//We parse the JSON response into the entity struct.
	newEntity, err := d.Parser(jsonBytes)
	if err != nil {
		var zero TInterface
		log.Println("Failed to unmarshal entity: ", err)
		return zero, err
	}
	//We convert the entity struct to the interface and return it.
	return interface{}(newEntity).(TInterface), nil
}

// Update updates an entity. This one can be a little complicated, as the entitiy has been generating a list of updates to be made as you've been working with it. THis was automatic.
func (d *EntityDataAccessClient[TEntity, TInterface]) Update(entity TInterface) error {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	//We convert the list of updates to the right format for the bulk request.
	updates := *entity.GetUpdates()
	entity.ClearUpdates()
	var updatesInterface []interface{}
	for _, u := range updates {
		updatesInterface = append(updatesInterface, u)
	}
	//We send a PUT request to the database service to update the entity.
	bulkReturn, err := d._client.Put(d.PutRoute, updatesInterface)
	if err != nil {
		log.Println("Failed to update entity: ", err)
		entity.SetUpdates(&updates)
		return err
	}
	//If there are any errors, we log them and return an error. Otherwise, we return nil, because you already have the updated entity. You don't need a new one.
	if len(bulkReturn.Errors) > 0 {
		log.Println("Failed to update entity: ", bulkReturn.Errors)
		err = fmt.Errorf("failed to update entity: %v", bulkReturn.Errors)
		entity.SetUpdates(&updates)
		return err
	}
	return nil
}

// UpdateBulk updates entities in bulk. This is a bulk operation.
// It will return a map of any errors that occurred, mapped to their respective entities ID.
func (d *EntityDataAccessClient[TEntity, TInterface]) UpdateBulk(entities *[]TInterface) (map[string]int, error) {
	if d.PutRoute == "" {
		d.PutRoute = d.DefaultRoute
	}
	changesPerEntity := make(map[string][]*entity.EntityUpdateData)
	entitiesById := make(map[string]TInterface)
	//We convert the list of updates to the right format for the bulk request.
	var interfaces []interface{}
	for _, v := range *entities {
		entitiesById[v.GetIdString()] = v
		changesPerEntity[v.GetIdString()] = *v.GetUpdates()
		for _, u := range *v.GetUpdates() {
			interfaces = append(interfaces, u)
		}
		v.ClearUpdates()
	}
	//We don't actually have a put bulk request, because it's always just a list of updates. So we just send a put request with the updates. This one just has updates for multiple entities, and reutrns errors for mulitple entities.
	bulkReturn, err := d._client.Put(d.PutRoute, interfaces)
	if err != nil {
		var mapErrs map[string]int
		log.Println("Failed to update entities: ", err)
		return mapErrs, err
	}
	for id, err := range bulkReturn.Errors {
		log.Println("Failed to update entity: ", id, " with Error Code: ", err, ". Re-adding Updates to entity.")
		//We put the updates back in the entity, so they can be tried again.
		tempUpdates := changesPerEntity[id]
		entitiesById[id].SetUpdates(&tempUpdates)
	}
	return bulkReturn.Errors, nil
}

// Delete deletes an entity.
func (d *EntityDataAccessClient[TEntity, TInterface]) Delete(id *uuid.UUID) error {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	//We send a DELETE request to the database service to delete the entity.
	_, err := d._client.Delete(d.DeleteRoute + "/" + id.String())
	if err != nil {
		log.Println("Failed to delete entity: ", err)
		return err
	}
	//Entity is deleted. No reason to return anything.
	return nil
}

// DeleteBulk deletes entities in bulk. This is a bulk operation.
// It will return a map of any errors that occurred, mapped to their respective entities ID.
func (d *EntityDataAccessClient[TEntity, TInterface]) DeleteBulk(ids []*uuid.UUID) (map[string]int, error) {
	if d.DeleteRoute == "" {
		d.DeleteRoute = d.DefaultRoute
	}
	//We convert the list of IDs to strings. Easier to manage here.
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, id.String())
	}
	//We send a DELETE request to the database service to delete entities by their IDs.
	bulkReturn, err := d._client.DeleteBulk(d.DeleteRoute, idsStr)
	if err != nil {
		var mapErrs map[string]int
		log.Println("Failed to delete entities: ", err)
		return mapErrs, err
	}
	return bulkReturn.Errors, nil
}
