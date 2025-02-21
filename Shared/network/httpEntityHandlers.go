package network

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/entity"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

func CreateNetworkEntityHandlers[T entity.EntityInterface](network NetworkInterface, entityName string, databaseManager databaseService.EntityDataInterface[T], Parse func(jsonBytes []byte) (T, error)) {
	defaults := func(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
		if requestType == "GET" || requestType == "" {
			if queryParams.Get("id") != "" {
				entity, err := databaseManager.GetByID(queryParams.Get("id"))
				if errors.Is(err, gorm.ErrRecordNotFound) {
					responseWriter.WriteHeader(http.StatusNotFound)
					return
				}
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}
				entityJSON, err := entity.EntityToJSON()
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}
				responseWriter.Write(entityJSON)
			} else {
				var entities *[]T
				var err error
				if len(queryParams) == 0 {
					entities, err = databaseManager.GetAll()
				} else if queryParams.Get("ids") != "" {
					entities, err = databaseManager.GetByIDs(strings.Split(queryParams.Get("ids"), ","))
				}
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}

				entitiesJSON, err := json.Marshal(entities)
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}
				responseWriter.Write(entitiesJSON)
			}
		} else if requestType == "POST" {
			entity, err := Parse(data)
			if err != nil {
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			err = databaseManager.Create(entity)
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			entityJSON, err := entity.EntityToJSON()
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseWriter.Write(entityJSON)
		} else if requestType == "PUT" {
			entity, err := Parse(data)
			if err != nil {
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			err = databaseManager.Update(entity)
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			entityJSON, err := entity.EntityToJSON()
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseWriter.Write(entityJSON)
		} else if requestType == "DELETE" {
			err := databaseManager.Delete(queryParams.Get("id"))
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	network.AddHandleFunc(HandlerParams{Pattern: entityName + "/", Handler: defaults})
	network.AddHandleFunc(HandlerParams{Pattern: entityName, Handler: defaults})
}
