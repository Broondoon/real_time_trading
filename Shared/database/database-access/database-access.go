package databaseAccess

type BaseDatabaseAccessInterface interface {
}

type BaseDatabaseAccess struct {
}

type DatabaseAccessInterface interface {
	Connect()
	Disconnect()
}

type NewDatabaseAccessParams struct {
}

func NewBaseDatabaseAccess(params *NewDatabaseAccessParams) BaseDatabaseAccessInterface {
	return &BaseDatabaseAccess{}
}
