package databaseAccessAuth

import (
	"errors"
	"fmt"
	"os"

	databaseAccess "Shared/database/database-access"
	user "Shared/entities/user"
	"Shared/network" // for the HTTP client (package network)
)

// UserDataAccessInterface defines the operations that auth-service needs.
type UserDataAccessInterface interface {
	databaseAccess.EntityDataAccessInterface[*user.User, user.UserInterface]
	// GetUserByUsername queries by the "username" field.
	GetUserByUsername(username string) (user.UserInterface, error)
	// CreateUser creates a new user.
	CreateUser(u *user.User) error
}

// UserDataAccess is the concrete implementation.
type UserDataAccess struct {
	databaseAccess.EntityDataAccessInterface[*user.User, user.UserInterface]
}

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	User() UserDataAccessInterface
}

type DatabaseAccess struct {
	UserDataAccessInterface
	_networkManager network.NetworkInterface
}

// NewUserDataAccessParams holds parameters for creating a new auth data access.
type NewDatabaseAccessParams struct {
	UserParams *databaseAccess.NewEntityDataAccessHTTPParams[*user.User]
	Network    network.NetworkInterface
}

// NewUserDataAccess creates an UserDataAccessInterface instance.
func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.UserParams == nil {
		params.UserParams = &databaseAccess.NewEntityDataAccessHTTPParams[*user.User]{}
	}
	if params.Network == nil {
		panic("No Network provided")
	}
	if params.UserParams.Client == nil {
		params.UserParams.Client = params.Network.AuthDatabase()
	}
	// Use an environment variable for the default route.
	if params.UserParams.DefaultRoute == "" {
		params.UserParams.DefaultRoute = os.Getenv("AUTH_SERVICE_USER_ROUTE")
	}
	if params.UserParams.Parser == nil {
		params.UserParams.Parser = user.Parse
	}
	if params.UserParams.ParserList == nil {
		params.UserParams.ParserList = user.ParseList
	}

	dba := &DatabaseAccess{
		UserDataAccessInterface: &UserDataAccess{
			EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*user.User, user.UserInterface](params.UserParams),
		},
		_networkManager: params.Network,
	}
	dba.Connect()
	return dba
}

func (a *DatabaseAccess) Connect() {
}

func (a *UserDataAccess) Disconnect() {
}

func (d *DatabaseAccess) User() UserDataAccessInterface {
	return d.UserDataAccessInterface
}

// GetUserByUsername uses the generic GetByForeignID method from the shared layer.
func (d *UserDataAccess) GetUserByUsername(username string) (user.UserInterface, error) {
	users, err := d.GetByForeignID("Username", username)
	if err != nil {
		return nil, err
	}
	if len(*users) == 0 {
		return nil, errors.New("user not found")
	}
	u, ok := (*users)[0].(*user.User)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to *user.User")
	}
	return u, nil
}

// CreateUser calls the generic Create method.
func (d *UserDataAccess) CreateUser(u *user.User) error {
	_, err := d.Create(u)
	return err
}
