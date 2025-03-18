package databaseServiceAuth

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/user"
	"os"
	"time"
)

// Setting a basic setup for an EntityDataInterface for the user entity for service (i.e the side connected to the database)
// This is slightly redudent, as we can just use databaseService.EntityDataInterface, but this prevents us from making type checks, and allows us to add methods specific to the user entity.
type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*user.User]
}

// Struct for holding that connection
type DatabaseService struct {
	databaseService.EntityDataInterface[*user.User]
}

// params for creating a new database service for this entity.
type NewDatabaseServiceParams struct {
	*databaseService.NewEntityDataParams // leave nil for default
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	//Set defaults to prevent null reference errors
	if params.NewEntityDataParams == nil {
		params.NewEntityDataParams = &databaseService.NewEntityDataParams{}
	}

	//Create the database service, wrapped in a redis cache. Leave the creation of the actual service to the cache constructor.
	db := &DatabaseService{
		EntityDataInterface: databaseService.NewCachedEntityData[*user.User](&databaseService.NewCachedEntityDataParams{
			NewEntityDataParams: params.NewEntityDataParams,
			RedisAddr:           os.Getenv("REDIS_ADDR"),
			Password:            os.Getenv("REDIS_PASSWORD"),
			DefaultTTL:          5 * time.Minute,
		}),
		//EntityDataInterface: databaseService.NewEntityData[*user.User](params.NewEntityDataParams),
	}
	//Connect and ensure the schemea in the DB matches the User Entity
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
