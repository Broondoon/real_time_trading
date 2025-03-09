package network

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/entity"

	"encoding/json"
	"errors"
	"fmt"
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
	WriteHeader(statusCode int)
	Write([]byte) (int, error)
	Header() http.Header
	EncodeResponse(statusCode int, response map[string]interface{})
}

type HandlerParams struct {
	Pattern     string
	Handler     func(ResponseWriter, []byte, url.Values, string)
	RequestType string
}

func CreateNetworkEntityHandlers[T entity.EntityInterface](network NetworkInterface, entityName string, databaseManager databaseService.EntityDataInterface[T], Parse func(jsonBytes []byte) (T, error), ParseList func(jsonBytes []byte) (*[]T, error)) {
	defaults := func(responseWriter ResponseWriter, data []byte, queryParams url.Values, requestType string) {
		fmt.Println("-----------------\nRequest:")
		if requestType == "POST" || requestType == "PUT" {
			fmt.Println("data: ", string(data))
		}
		fmt.Println("queryParams: ", queryParams.Encode())
		fmt.Println("requestType: ", requestType)
		fmt.Println("-----------------")
		bulkRequest := false
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
			if queryParams.Get("isBulk") != "" {
				ids := strings.Split(queryParams.Get("ids"), ",")
				if foreignKey := queryParams.Get("foreignKey"); foreignKey != "" {
					entities, errorsReceived = databaseManager.GetByForeignIDBulk(foreignKey, ids)
				} else {
					entities, errorsReceived = databaseManager.GetByIDs(ids)
				}
				useEntities = true
				bulkRequest = true
			} else if id := queryParams.Get("id"); id != "" {
				if foreignKey := queryParams.Get("foreignKey"); foreignKey != "" {
					entities, err = databaseManager.GetByForeignID(foreignKey, id)
					useEntities = true
				} else {
					entityObj, err = databaseManager.GetByID(id)
				}
			} else {
				entities, err = databaseManager.GetAll()
				useEntities = true
			}

		case "POST":
			var isBulk bool
			if queryParams.Get("isBulk") != "" {
				entities, err = ParseList(data)
				isBulk = true
			} else {
				entityObj, err = Parse(data)
			}
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			if isBulk {
				errorsReceived = databaseManager.CreateBulk(entities)
				useEntities = true
				bulkRequest = true
			} else {
				err = databaseManager.Create(entityObj)
			}
		case "PUT":
			updates := make([]*entity.EntityUpdateData, 0)
			err = json.Unmarshal(data, &updates)
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			errorsReceived = databaseManager.Update(updates)
			if isBulk := queryParams.Get("isBulk"); isBulk != "" {
				bulkRequest = true
			}
			noReturns = true
		case "DELETE":
			if isBulk := queryParams.Get("isBulk"); isBulk != "" {
				errorsReceived = databaseManager.DeleteBulk(strings.Split(queryParams.Get("ids"), ","))
				bulkRequest = true
			} else {
				err = databaseManager.Delete(queryParams.Get("id"))
			}
		default:
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		if errorsReceived != nil {
			if err = errorsReceived["transaction"]; err == nil {
				for id, err := range errorsReceived {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						errorList[id] = http.StatusNotFound
					} else {
						errorList[id] = http.StatusInternalServerError
					}
				}
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
		if noReturns {
			responseWriter.WriteHeader(http.StatusOK)
			return
		}

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
			fmt.Println("error: ", err.Error())
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		if bulkRequest {
			jsonVal, err = json.Marshal(BulkReturn{Entities: jsonVal, Errors: errorList})
			if err != nil {
				fmt.Println("error: ", err.Error())
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
