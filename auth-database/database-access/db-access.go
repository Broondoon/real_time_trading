package databaseAccessAuth

import (
	"os"

	databaseAccess "Shared/database/database-access"
	user "Shared/entities/user"
	"Shared/network" // for the HTTP client (package network)
)

// Base data access interface set to the User entity. This gives us the methods to interact with the database on a generic level.
type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*user.User, user.UserInterface] // [Base entity, Interface for base entity so we can convert when necceessary]

// DatabaseAccessInterface is the interface for the database access for user. Any extra methods can be inserted here.
type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface //Basic methods, such as connect and disconnect for the database
	EntityDataAccessInterface              //methods for interacting with the database specific to the user. Stuff like Create or Update.
}

// Struct to hold the generic database access for the user entity, and the network manager to communicate with the database services
type DatabaseAccess struct {
	EntityDataAccessInterface
	_networkManager network.NetworkInterface
}

// NewUserDataAccessParams holds parameters for creating a new auth data access.
type NewDatabaseAccessParams struct {
	*databaseAccess.NewEntityDataAccessNetworkParams[*user.User]
	Network network.NetworkInterface //Network to be used for the database access to access the service. Currently we use a HTTP client.
}

// NewUserDataAccess creates an UserDataAccessInterface instance.
func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	//Set defaults to prevent null reference errors
	if params.NewEntityDataAccessNetworkParams == nil {
		params.NewEntityDataAccessNetworkParams = &databaseAccess.NewEntityDataAccessNetworkParams[*user.User]{}
	}
	//We need a network. System won't work without it. Panic if we don't have one.
	if params.Network == nil {
		panic("No Network provided")
	}
	//Default the base pathway for listening to this service
	if params.NewEntityDataAccessNetworkParams.Client == nil {
		params.NewEntityDataAccessNetworkParams.Client = params.Network.AuthDatabase()
	}
	// Use an environment variable for the default route.
	if params.NewEntityDataAccessNetworkParams.DefaultRoute == "" {
		params.NewEntityDataAccessNetworkParams.DefaultRoute = os.Getenv("AUTH_SERVICE_USER_ROUTE")
	}
	//Set parsers for the user object. These are found in the shared/entities/user/user.go file.
	if params.NewEntityDataAccessNetworkParams.Parser == nil {
		params.NewEntityDataAccessNetworkParams.Parser = user.Parse
	}
	if params.NewEntityDataAccessNetworkParams.ParserList == nil {
		params.NewEntityDataAccessNetworkParams.ParserList = user.ParseList
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*user.User, user.UserInterface](params.NewEntityDataAccessNetworkParams),
		_networkManager:           params.Network,
	}
	dba.Connect()
	return dba
}

func (a *DatabaseAccess) Connect() {
}

func (a *DatabaseAccess) Disconnect() {
}
