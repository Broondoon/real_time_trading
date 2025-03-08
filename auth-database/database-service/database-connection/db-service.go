package databaseServiceAuth

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/user"
)

type UserDataServiceInterface interface {
	databaseService.EntityDataInterface[*user.User]
}

type DatabaseServiceInterface interface {
	databaseService.DatabaseInterface
	User() UserDataServiceInterface
}

type DatabaseService struct {
	UserInterface UserDataServiceInterface
	databaseService.DatabaseInterface
}

type NewDatabaseServiceParams struct {
	UserParams *databaseService.NewEntityDataParams // leave nil for default
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	if params.UserParams == nil {
		params.UserParams = &databaseService.NewEntityDataParams{
			NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
		}
	}
	var newDBConnection databaseService.PostGresDatabaseInterface
	if params.UserParams.Existing != nil {
		newDBConnection = params.UserParams.Existing
	} else {
		newDBConnection = databaseService.NewPostGresDatabase(params.UserParams.NewPostGresDatabaseParams)
		params.UserParams.Existing = newDBConnection
	}

	db := &DatabaseService{
		UserInterface:     databaseService.NewEntityData[*user.User](params.UserParams),
		DatabaseInterface: newDBConnection,
	}
	db.Connect()
	db.User().GetDatabaseSession().AutoMigrate(&user.User{})
	return db
}

func (d *DatabaseService) User() UserDataServiceInterface {
	return d.UserInterface
}

func (d *DatabaseService) Connect() {
	d.User().Connect()
	d.User().Connect()
}

func (d *DatabaseService) Disconnect() {
	d.User().Disconnect()
	d.User().Disconnect()
}
