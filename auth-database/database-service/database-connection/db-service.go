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
		EntityDataInterface: databaseService.NewCachedEntityData[*user.User](&databaseService.NewCachedEntityDataParams{
			NewEntityDataParams: params.NewEntityDataParams,
			RedisAddr:           os.Getenv("REDIS_ADDR"),
			Password:            os.Getenv("REDIS_PASSWORD"),
			DefaultTTL:          5 * time.Minute,
		}),
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
