// In auth-database/databaseAccess/db-access.go
package databaseAccessAuth

import (
	"errors"
	"fmt"
	"os"

	databaseAccess "Shared/database/database-access"
	"Shared/entities/user"
	"Shared/network"
)

// AuthDataAccessInterface defines the operations required by auth-service.
type AuthDataAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	// GetUserByUsername (or similar) is the method to check for an existing username.
	GetUserByUsername(username string) (*user.User, error)
	CreateUser(u *user.User) error
	// ... other methods as needed.
}

// AuthDataAccess is the concrete implementation.
type AuthDataAccess struct {
	databaseAccess.EntityDataAccessInterface[*user.User, user.UserInterface]
	_networkManager network.NetworkInterface
}

// NewDatabaseAccessParams are the parameters needed to create the auth database access.
type NewDatabaseAccessParams struct {
	UserParams *databaseAccess.NewEntityDataAccessHTTPParams[*user.User]
	Network    network.NetworkInterface
}

// NewDatabaseAccess creates a new auth database access implementation.
func NewDatabaseAccess(params *NewDatabaseAccessParams) AuthDataAccessInterface {
	if params.UserParams == nil {
		params.UserParams = &databaseAccess.NewEntityDataAccessHTTPParams[*user.User]{}
	}
	if params.Network == nil {
		panic("No Network provided.")
	}
	if params.UserParams.Client == nil {
		params.UserParams.Client = params.Network.AuthDatabase()
	}
	// Set default route if needed.
	if params.UserParams.DefaultRoute == "" {
		params.UserParams.DefaultRoute = os.Getenv("AUTH_SERVICE_USER_ROUTE")
	}
	if params.UserParams.Parser == nil {
		params.UserParams.Parser = user.Parse
	}
	if params.UserParams.ParserList == nil {
		params.UserParams.ParserList = user.ParseList
	}

	dba := &AuthDataAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*user.User, user.UserInterface](params.UserParams),
		_networkManager:           params.Network,
	}
	// Optionally call Connect() if needed.
	dba.Connect()
	return dba
}

func (a *AuthDataAccess) Connect() {
	// Implementation can be empty if the HTTP–based access doesn't require an explicit connection.
}

func (a *AuthDataAccess) Disconnect() {
	// Similarly, provide a disconnect if needed.
}

// GetUserByUsername fetches a user using the "username" column.
func (a *AuthDataAccess) GetUserByUsername(username string) (*user.User, error) {
	// Assuming the generic access provides GetByForeignID:
	users, err := a.GetByForeignID("username", username)
	if err != nil {
		return nil, err
	}
	if len(*users) == 0 {
		return nil, errors.New("user not found")
	}
	// Assuming the first result is the one we want.
	u, ok := (*users)[0].(*user.User)
	if !ok {
		return nil, fmt.Errorf("failed to convert result")
	}
	return u, nil
}

// CreateUser calls the generic Create method.
func (a *AuthDataAccess) CreateUser(u *user.User) error {
	_, err := a.Create(u)
	return err
}
