package databaseServiceAuth

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/user"
	"os"
	"time"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*user.User]
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*user.User]
}

type NewDatabaseServiceParams struct {
	*databaseService.NewEntityDataParams // leave nil for default
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	if params.NewEntityDataParams == nil {
		params.NewEntityDataParams = &databaseService.NewEntityDataParams{}
	}
	

	db := &DatabaseService{
<<<<<<< HEAD
		EntityDataInterface: cachedUser,

	}
	db.Connect()
	db.GetDatabaseSession().AutoMigrate(&user.User{})
	return db
}

func (d *DatabaseService) Connect() {
	d.EntityDataInterface.Connect()
}

func (d *DatabaseService) Disconnect() {
	d.EntityDataInterface.Disconnect()
}
