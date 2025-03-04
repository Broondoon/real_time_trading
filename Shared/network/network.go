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
}

type HandlerParams struct {
	Pattern     string
	Handler     func(ResponseWriter, []byte, url.Values, string)
	RequestType string
}

func CreateNetworkEntityHandlers[T entity.EntityInterface](network NetworkInterface, entityName string, databaseManager databaseService.EntityDataInterface[T], Parse func(jsonBytes []byte) (T, error)) {
	defaults := func(responseWriter ResponseWriter, data []byte, queryParams url.Values, requestType string) {
		fmt.Println("-----------------\nRequest:")
		if requestType == "POST" || requestType == "PUT" {
			fmt.Println("data: ", string(data))
		}
		fmt.Println("queryParams: ", queryParams.Encode())
		fmt.Println("requestType: ", requestType)
		fmt.Println("-----------------")
		if requestType == "GET" || requestType == "" {
			if queryParams.Get("id") != "" {
				if queryParams.Get("foreignKey") != "" {
					entities, err := databaseManager.GetByForeignID(queryParams.Get("foreignKey"), queryParams.Get("id"))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						responseWriter.WriteHeader(http.StatusNotFound)
						return
					}
					if err != nil {
						fmt.Println("error: ", err.Error())
						responseWriter.WriteHeader(http.StatusInternalServerError)
						return
					}
					entitiesJSON, err := json.Marshal(entities)
					if err != nil {
						fmt.Println("error: ", err.Error())
						responseWriter.WriteHeader(http.StatusInternalServerError)
						return
					}
					responseWriter.Write(entitiesJSON)
				} else {
					entity, err := databaseManager.GetByID(queryParams.Get("id"))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						responseWriter.WriteHeader(http.StatusNotFound)
						return
					}
					if err != nil {
						fmt.Println("error: ", err.Error())
						responseWriter.WriteHeader(http.StatusInternalServerError)
						return
					}
					entityJSON, err := entity.ToJSON()
					if err != nil {
						fmt.Println("error: ", err.Error())
						responseWriter.WriteHeader(http.StatusInternalServerError)
						return
					}
					responseWriter.Write(entityJSON)
				}
			} else {
				var entities *[]T
				var err error
				if queryParams.Get("ids") != "" {
					entities, err = databaseManager.GetByIDs(strings.Split(queryParams.Get("ids"), ","))
				} else {
					entities, err = databaseManager.GetAll()
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					responseWriter.WriteHeader(http.StatusNotFound)
					return
				}
				if err != nil {
					fmt.Println("error: ", err.Error())
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}

				entitiesJSON, err := json.Marshal(entities)
				if err != nil {
					fmt.Println("error: ", err.Error())
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}
				responseWriter.Write(entitiesJSON)
			}
		} else if requestType == "POST" {
			entity, err := Parse(data)
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			err = databaseManager.Create(entity)
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			entityJSON, err := entity.ToJSON()
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseWriter.Write(entityJSON)
		} else if requestType == "PUT" {
			entity, err := Parse(data)
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			err = databaseManager.Update(entity)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			}
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			entityJSON, err := entity.ToJSON()
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseWriter.Write(entityJSON)
		} else if requestType == "DELETE" {
			err := databaseManager.Delete(queryParams.Get("id"))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			}
			if err != nil {
				fmt.Println("error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	network.AddHandleFuncUnprotected(HandlerParams{Pattern: entityName + "/", Handler: defaults})
	network.AddHandleFuncUnprotected(HandlerParams{Pattern: entityName, Handler: defaults})
}
