package network

type ClientInterface interface {
	Get(route string, headers map[string]string) ([]byte, error)
	GetBulk(route string, ids []string, headers map[string]string) (BulkReturn, error)
	PostBulk(endpoint string, payload []interface{}) (BulkReturn, error)
	Post(route string, payload interface{}) ([]byte, error)
	Put(route string, payload []interface{}) (BulkReturn, error)
	Patch(route string, id string) error
	PatchBulk(route string, ids []string) (BulkReturn, error)
	DeleteBulk(route string, payload []string) (BulkReturn, error)
	Delete(route string) ([]byte, error)
	GetBaseURL() string
}
