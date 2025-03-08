package network

type ClientInterface interface {
	Get(route string, headers map[string]string) ([]byte, error)
	PostBulk(endpoint string, payload []interface{}) ([]byte, error)
	Post(route string, payload interface{}) ([]byte, error)
	Put(route string, payload interface{}) ([]byte, error)
	Delete(route string) ([]byte, error)
	GetBaseURL() string
}
