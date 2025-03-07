package network

type ClientInterface interface {
	Get(route string, headers map[string]string) ([]byte, error)
	PostBulk(endpoint string, payload []interface{}) ([]byte, error)
	Post(route string, payload interface{}) ([]byte, error)
	Put(route string, payload []interface{}) error
	Patch(route string, id string) error
	PatchBulk(route string, ids []string) error
	DeleteBulk(route string, payload []string) ([]byte, error)
	Delete(route string) ([]byte, error)
	GetBaseURL() string
}
