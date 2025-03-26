package network

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/entity"
	"Shared/objects"
	"log"

	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gorm.io/gorm"
)

type BaseNetworkInterface interface {
	MatchingEngine() ClientInterface
	MicroserviceTemplate() ClientInterface
	UserManagement() ClientInterface
	Authentication() ClientInterface
	OrderInitiator() ClientInterface
	OrderExecutor() ClientInterface
	Stocks() ClientInterface
	Transactions() ClientInterface
	UserManagementDatabase() ClientInterface
	AuthDatabase() ClientInterface
}

type Network struct {
	MatchingEngineService         ClientInterface
	MicroserviceTemplateService   ClientInterface
	UserManagementService         ClientInterface
	AuthenticationService         ClientInterface
	OrderInitiatorService         ClientInterface
	OrderExecutorService          ClientInterface
	StocksService                 ClientInterface
	TransactionsService           ClientInterface
	UserManagementDatabaseService ClientInterface
	AuthDatabaseService           ClientInterface
	serviceBuilder                func(serviceString string) ClientInterface
}

func (n *Network) MatchingEngine() ClientInterface {
	if n.MatchingEngineService == nil {
		n.MatchingEngineService = n.serviceBuilder(os.Getenv("MATCHING_ENGINE_HOST") + ":" + os.Getenv("MATCHING_ENGINE_PORT"))
	}
	return n.MatchingEngineService
}

func (n *Network) MicroserviceTemplate() ClientInterface {
	if n.MicroserviceTemplateService == nil {
		n.MicroserviceTemplateService = n.serviceBuilder(os.Getenv("MICROSERVICE_TEMPLATE_HOST") + ":" + os.Getenv("MICROSERVICE_TEMPLATE_PORT"))
	}
	return n.MicroserviceTemplateService
}

func (n *Network) UserManagement() ClientInterface {
	if n.UserManagementService == nil {
		n.UserManagementService = n.serviceBuilder(os.Getenv("USER_MANAGEMENT_HOST") + ":" + os.Getenv("USER_MANAGEMENT_PORT"))
	}
	return n.UserManagementService
}

func (n *Network) Authentication() ClientInterface {
	if n.AuthenticationService == nil {
		n.AuthenticationService = n.serviceBuilder(os.Getenv("AUTH_HOST") + ":" + os.Getenv("AUTH_PORT"))
	}
	return n.AuthenticationService
}

func (n *Network) OrderInitiator() ClientInterface {
	if n.OrderInitiatorService == nil {
		n.OrderInitiatorService = n.serviceBuilder(os.Getenv("ORDER_INITIATOR_HOST") + ":" + os.Getenv("ORDER_INITIATOR_PORT"))
	}
	return n.OrderInitiatorService
}

func (n *Network) OrderExecutor() ClientInterface {
	if n.OrderExecutorService == nil {
		n.OrderExecutorService = n.serviceBuilder(os.Getenv("ORDER_EXECUTOR_HOST") + ":" + os.Getenv("ORDER_EXECUTOR_PORT"))
	}
	return n.OrderExecutorService
}

func (n *Network) Stocks() ClientInterface {
	if n.StocksService == nil {
		n.StocksService = n.serviceBuilder(os.Getenv("STOCK_DATABASE_SERVICE_HOST") + ":" + os.Getenv("STOCK_DATABASE_SERVICE_PORT"))
	}
	return n.StocksService
}

func (n *Network) Transactions() ClientInterface {
	if n.TransactionsService == nil {
		n.TransactionsService = n.serviceBuilder(os.Getenv("TRANSACTION_DATABASE_SERVICE_HOST") + ":" + os.Getenv("TRANSACTION_DATABASE_SERVICE_PORT"))
	}
	return n.TransactionsService
}

func (n *Network) UserManagementDatabase() ClientInterface {
	if n.UserManagementDatabaseService == nil {
		n.UserManagementDatabaseService = n.serviceBuilder(os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_HOST") + ":" + os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_PORT"))
	}
	return n.UserManagementDatabaseService
}

func (n *Network) AuthDatabase() ClientInterface {
	if n.AuthDatabaseService == nil {
		n.AuthDatabaseService = n.serviceBuilder(os.Getenv("AUTH_DATABASE_SERVICE_HOST") + ":" + os.Getenv("AUTH_DATABASE_SERVICE_PORT"))
	}
	return n.AuthDatabaseService
}

func NewNetwork(serviceBuilder func(serviceString string) ClientInterface) BaseNetworkInterface {
	return &Network{
		serviceBuilder: serviceBuilder,
	}
}

type NetworkInterface interface {
	BaseNetworkInterface
	Listen()
	AddHandleFuncUnprotected(params HandlerParams)
	AddHandleFuncProtected(params HandlerParams)
}

type ResponseWriter interface {
	http.ResponseWriter
	EncodeResponse(statusCode int, response map[string]interface{})
	CheckCompleted() bool
	GetStatusCode() int
}

type HandlerParams struct {
	Pattern     string
	Handler     func(ResponseWriter, []byte, url.Values, string)
	RequestType string
}

func CreateNetworkEntityHandlers[T entity.EntityInterface, TDatabase any](network NetworkInterface, entityName string, databaseManager databaseService.EntityDataInterface[T, TDatabase], Parse func(jsonBytes []byte) (T, error), ParseList func(jsonBytes []byte) (*[]T, error)) {
	defaults := func(responseWriter ResponseWriter, data []byte, queryParams url.Values, requestType string) {
		// log.Println("-----------------\nRequest:")
		// log.Println("entityName: ", entityName)
		// if requestType == "POST" || requestType == "PUT" {
		// 	log.Println("data: ", string(data))
		// }
		// log.Println("queryParams: ", queryParams.Encode())
		// log.Println("requestType: ", requestType)
		// log.Println("-----------------")
		bulkRequest := queryParams.Get("Isbulk") != ""
		useEntities := false
		noReturns := false
		errorList := make(map[string]int)
		errorsReceived := make(map[string]error)
		var err error
		var entityObj T
		var entities *[]T
		if requestType == "" {
			requestType = "GET"
		}
		switch requestType {
		case "GET":
			if bulkRequest {
				ids := strings.Split(queryParams.Get("Ids"), ",")
				if key1 := queryParams.Get("IdColumn1"); key1 != "" {
					pairedIds := make([]objects.Pair, len(ids)/2)
					for i := 0; i < len(ids); i += 2 {
						pairedIds[i/2] = objects.Pair{ID1: ids[i], ID2: ids[i+1]}
					}
					entities, errorsReceived = databaseManager.GetByPairedIDBulk(key1, queryParams.Get("IdColumn2"), &pairedIds)
				} else if foreignKey := queryParams.Get("foreignKey"); foreignKey != "" {
					if filterKey := queryParams.Get("filterKey"); filterKey != "" {
						entities, errorsReceived = databaseManager.GetByFilteredForeignIDBulk(foreignKey, ids, filterKey, queryParams.Get("filterVal"))
					} else {
						entities, errorsReceived = databaseManager.GetByForeignIDBulk(foreignKey, ids)
					}
				} else {
					entities, errorsReceived = databaseManager.GetByIDs(ids)
				}
				useEntities = true
			} else if id := queryParams.Get("id"); id != "" {
				if foreignKey := queryParams.Get("foreignKey"); foreignKey != "" {
					entities, err = databaseManager.GetByForeignID(foreignKey, id)
					useEntities = true
				} else {
					entityObj, err = databaseManager.GetByID(id)
				}
			} else if key1 := queryParams.Get("IdColumn1"); key1 != "" {
				entityObj, err = databaseManager.GetByPairedID(key1, queryParams.Get("IdColumn2"), objects.Pair{ID1: queryParams.Get("Id1"), ID2: queryParams.Get("Id2")})
			} else {
				entities, err = databaseManager.GetAll()
				useEntities = true
			}

		case "POST":
			if bulkRequest {
				entities, err = ParseList(data)
			} else {
				entityObj, err = Parse(data)
			}
			if err != nil {
				log.Println("network POST error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			if bulkRequest {
				errorsReceived = databaseManager.CreateBulk(entities)
				useEntities = true
			} else {
				err = databaseManager.Create(entityObj)
			}
		case "PUT":
			updates := make([]*entity.EntityUpdateData, 0)
			err = json.Unmarshal(data, &updates)
			if err != nil {
				log.Println("network PUT error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			errorsReceived = databaseManager.Update(updates)
			noReturns = true
		case "DELETE":
			if bulkRequest {
				errorsReceived = databaseManager.DeleteBulk(strings.Split(queryParams.Get("Ids"), ","))
			} else {
				err = databaseManager.Delete(queryParams.Get("id"))
			}
			noReturns = true
		default:
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		if errorsReceived != nil {
			if _, ok := errorsReceived["transaction"]; !ok {
				for id, err := range errorsReceived {
					log.Println("Transfer Error: ", err)
					if errors.Is(err, gorm.ErrRecordNotFound) {
						errorList[id] = http.StatusNotFound
					} else {
						errorList[id] = http.StatusInternalServerError
					}
				}
			} else {
				log.Printf("Transaction error: %v\n", errorsReceived["transaction"])
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			} else {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		var jsonVal []byte

		if useEntities {
			jsonVal, err = json.Marshal(entities)
		} else if noReturns {
			if bulkRequest {
				jsonVal = []byte{}
			} else {
				responseWriter.WriteHeader(http.StatusOK)
				return
			}
		} else {
			jsonVal, err = entityObj.ToJSON()
		}
		if err != nil {
			log.Println("Network General marshal error: ", err.Error())
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		if bulkRequest {
			jsonVal, err = json.Marshal(BulkReturn{Entities: jsonVal, Errors: errorList})
			if err != nil {
				log.Println("Networ Bulkify marshal error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		responseWriter.Write(jsonVal)
	}

	network.AddHandleFuncUnprotected(HandlerParams{Pattern: entityName + "/", Handler: defaults})
	network.AddHandleFuncUnprotected(HandlerParams{Pattern: entityName, Handler: defaults})
}

type BulkReturn struct {
	Entities []byte         `json:"entities"`
	Errors   map[string]int `json:"errors"`
}
