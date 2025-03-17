package databaseAccessAuth

import (
	"os"

	databaseAccess "Shared/database/database-access"
	user "Shared/entities/user"
	"Shared/network" // for the HTTP client (package network)
)

type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*user.User, user.UserInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	EntityDataAccessInterface
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	_networkManager network.NetworkInterface
}

// NewUserDataAccessParams holds parameters for creating a new auth data access.
type NewDatabaseAccessParams struct {
	*databaseAccess.NewEntityDataAccessHTTPParams[*user.User]
	Network network.NetworkInterface
}

// NewUserDataAccess creates an UserDataAccessInterface instance.
func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.NewEntityDataAccessHTTPParams == nil {
		params.NewEntityDataAccessHTTPParams = &databaseAccess.NewEntityDataAccessHTTPParams[*user.User]{}
	}
	if params.Network == nil {
		panic("No Network provided")
	}
	if params.NewEntityDataAccessHTTPParams.Client == nil {
		params.NewEntityDataAccessHTTPParams.Client = params.Network.AuthDatabase()
	}
	// Use an environment variable for the default route.
	if params.NewEntityDataAccessHTTPParams.DefaultRoute == "" {
		params.NewEntityDataAccessHTTPParams.DefaultRoute = os.Getenv("AUTH_SERVICE_USER_ROUTE")
	}
	if params.NewEntityDataAccessHTTPParams.Parser == nil {
		params.NewEntityDataAccessHTTPParams.Parser = user.Parse
	}
	if params.NewEntityDataAccessHTTPParams.ParserList == nil {
		params.NewEntityDataAccessHTTPParams.ParserList = user.ParseList
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*user.User, user.UserInterface](params.NewEntityDataAccessHTTPParams),
		_networkManager:           params.Network,
	}
	dba.Connect()
	return dba
}

func (a *DatabaseAccess) Connect() {
}

func (a *DatabaseAccess) Disconnect() {
}
