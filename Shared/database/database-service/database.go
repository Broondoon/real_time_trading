package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BaseDatabaseInterface interface {
	GetDBUrl() string
}
type BaseDatabase struct {
	Db_ENV_PATH string
}

type NewBaseDatabaseParams struct {
	Db_ENV_PATH string
}

func NewBaseDatabase(params NewBaseDatabaseParams) *BaseDatabase {
	return &BaseDatabase{
		Db_ENV_PATH: params.Db_ENV_PATH,
	}
}
func (d *BaseDatabase) GetDBUrl() string {
	dsn := os.Getenv(d.Db_ENV_PATH) // "DATABASE_URL" is an ENV variable that
	// is set in docker-compose.yml
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
	}
	return dsn
}

type DatabaseInterface interface {
	BaseDatabaseInterface
	Connect()
	Disconnect()
}

type PostGresDatabaseInterface interface {
	DatabaseInterface
	GetDatabaseConnection() *gorm.DB
}

type PostGresDatabase struct {
	BaseDatabaseInterface
	database *gorm.DB
}

type NewPostGresDatabaseParams struct {
	NewBaseDatabaseParams
}

func NewPostGresDatabase(params NewPostGresDatabaseParams) *PostGresDatabase {
	return &PostGresDatabase{
		BaseDatabaseInterface: NewBaseDatabase(params.NewBaseDatabaseParams),
	}
}

func (d *PostGresDatabase) Connect() {
	dsn := d.GetDBUrl()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	d.database = db
}

func (d *PostGresDatabase) Disconnect() {
	db, err := d.database.DB()
	if err != nil {
		log.Fatal("Failed to disconnect from database: ", err)
	}
	db.Close()
}

func (d *PostGresDatabase) GetDatabaseConnection() *gorm.DB {
	return d.database
}
