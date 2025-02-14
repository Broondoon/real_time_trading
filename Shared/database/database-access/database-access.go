package databaseAccess

type BaseDatabaseAccessInterface interface {
	Connect()
	Disconnect()
}

type BaseDatabaseAccess struct {
}

type NewDatabaseAccessParams struct {
}

func NewBaseDatabaseAccess(params NewDatabaseAccessParams) *BaseDatabaseAccess {
	return &BaseDatabaseAccess{}
}
