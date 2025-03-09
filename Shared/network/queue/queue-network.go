package networkQueue

import (
	"Shared/network"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NetworkQueue struct {
	network.BaseNetworkInterface
	QueueConnectionInterface
	QueueClusters map[string]QueueClusterInterface
	ExchangeKey   string
}

func NewNetworkQueue(Connection *amqp.Connection, exchangeKey string) network.NetworkInterface {
	return &NetworkQueue{
		BaseNetworkInterface: network.NewNetwork(func(serviceString string) network.ClientInterface {
			return NewQueueClient(serviceString, &NewQueueClientParams{
				NewNetworkQueueConnectionParams: &NewNetworkQueueConnectionParams{
					Connection: Connection,
				},
			})
		}),
		QueueConnectionInterface: NewNetworkQueueConnection(&NewNetworkQueueConnectionParams{
			Connection: Connection,
		}),
		QueueClusters: make(map[string]QueueClusterInterface),
		ExchangeKey:   exchangeKey,
	}
}

func (n *NetworkQueue) AddHandleFuncUnprotected(params network.HandlerParams) {
	// handledHandler := network.HandlerParams{
	// 	Pattern: params.Pattern,
	// 	Handler: params.Handler,
	// }

	log.Println("Adding handle func unprotected: ", params.Pattern)
	n.QueueClusters[params.Pattern] = NewQueueCluster(n.ExchangeKey, params, &NewQueueClusterParams{
		NewNetworkQueueConnectionParams: &NewNetworkQueueConnectionParams{
			Connection: n.GetConnection(),
		},
		Durable:          true,
		AutoDelete:       false,
		Exclusive:        false,
		NoWait:           false,
		Args:             nil,
		ConsumeAutoAck:   false,
		ConsumeExclusive: false,
		ConsumeNoLocal:   false,
		ConsumeNoWait:    false,
		ConsumeArgs:      nil,
	})
	log.Println("Added handle func unprotected: ", params.Pattern)
	log.Println("Queue clusters: ", len(n.QueueClusters))
}

func (n *NetworkQueue) AddHandleFuncProtected(params network.HandlerParams) {
	panic("Internal Queues should not be used for External Requests")
}

func (n *NetworkQueue) Listen() {
	log.Println("Listening")
	log.Println("Queue clusters: ", len(n.QueueClusters))
	for _, params := range n.QueueClusters {
		log.Println("Spawning queue")
		go params.SpawnQueue()
		log.Println("Spawned queue")
	}
}

// func handleFunc(params network.HandlerParams, w http.ResponseWriter, r *http.Request) {
// 	// fmt.Println("Handling request for: ", r.URL.Path)
// 	var body []byte
// 	var err error
// 	queryParams := make(url.Values)
// 	queryParams, err = url.ParseQuery(r.URL.RawQuery)
// 	if err != nil {
// 		fmt.Println("Error, there was an issue with reading the message:", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodPut {
// 		//decode params
// 		queryParams = r.URL.Query()
// 		id := strings.TrimPrefix(r.URL.Path, "/"+params.Pattern)
// 		if id != "" {
// 			queryParams.Add("id", id)
// 		}
// 	}

// 	if r.Method == http.MethodPost || r.Method == http.MethodPut {
// 		body, err = io.ReadAll(r.Body)
// 		if err != nil {
// 			fmt.Println("Error, there was an issue with reading the message:", err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		defer r.Body.Close()
// 	}

// 	// The type assertion here is failing; r.Context().Value(userIDKey) returns a uint.
// 	// So we need to change that.
// 	if r.Context().Value(userIDKey) != nil {
// 		// queryParams.Add("userID", r.Context().Value(userIDKey).(string))
// 		if userID, ok := r.Context().Value(userIDKey).(uint); ok {
// 			queryParams.Add("userID", fmt.Sprintf("%d", userID)) // Convert to string
// 		} else if userID, ok := r.Context().Value(userIDKey).(string); ok {
// 			queryParams.Add("userID", userID)
// 		}
// 	}

// 	params.Handler(w, body, queryParams, r.Method)
// 	//w.WriteHeader(http.StatusOK)
// }
