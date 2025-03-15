package databaseServiceAuth

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/user"
	"os"
	"time"
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

	cachedUser := databaseService.NewCachedEntityData[*user.User](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.UserParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	db := &DatabaseService{
		UserInterface:     cachedUser,
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
