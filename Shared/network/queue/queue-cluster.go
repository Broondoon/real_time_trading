package networkQueue

import (
	"Shared/network"
	"net/url"
)

type QueueClusterInterface interface {
	QueueConnectionInterface
}

type QueueCluster struct {
	QueueConnectionInterface
	network.HandlerParams
	ExchangeKey      string
	Durable          bool
	AutoDelete       bool
	Exclusive        bool
	NoWait           bool
	Args             map[string]interface{}
	ConsumeAutoAck   bool
	ConsumeExclusive bool
	ConsumeNoLocal   bool
	ConsumeNoWait    bool
	ConsumeArgs      map[string]interface{}
}

type NewQueueClusterParams struct {
	*NewNetworkQueueConnectionParams
	Durable          bool
	AutoDelete       bool
	Exclusive        bool
	NoWait           bool
	Args             map[string]interface{}
	ConsumeAutoAck   bool
	ConsumeExclusive bool
	ConsumeNoLocal   bool
	ConsumeNoWait    bool
	ConsumeArgs      map[string]interface{}
}

func NewQueueCluster(exchangeKey string, handler network.HandlerParams, params *NewQueueClusterParams) QueueClusterInterface {
	if params.NewNetworkQueueConnectionParams == nil {
		params.NewNetworkQueueConnectionParams = &NewNetworkQueueConnectionParams{}
	}
	return &QueueCluster{
		QueueConnectionInterface: NewNetworkQueueConnection(params.NewNetworkQueueConnectionParams),
		HandlerParams:            handler,
		ExchangeKey:              exchangeKey,
		Durable:                  params.Durable,
		AutoDelete:               params.AutoDelete,
		Exclusive:                params.Exclusive,
		NoWait:                   params.NoWait,
		Args:                     params.Args,
	}
}

func GetDefaults() *NewQueueClusterParams {
	return &NewQueueClusterParams{
		NewNetworkQueueConnectionParams: &NewNetworkQueueConnectionParams{},
		Durable:                         true,
		AutoDelete:                      false,
		Exclusive:                       false,
		NoWait:                          false,
		Args:                            nil,
		ConsumeAutoAck:                  true,
		ConsumeExclusive:                false,
		ConsumeNoLocal:                  false,
		ConsumeNoWait:                   false,
		ConsumeArgs:                     nil,
	}
}

// Exchange Key. Bind this queue to an exchange with this key. We then filter incomming messages by pattern
func (n *QueueCluster) SpawnQueue() {
	exchangeParams := ExchangeParamsDefaults()
	exchangeParams.Name = n.ExchangeKey
	ch := n.SpawnChannel(exchangeParams)
	defer n.CloseChannel(ch)
	q, err := ch.QueueDeclare(
		n.HandlerParams.Pattern, // name
		n.Durable,               // durable
		n.AutoDelete,            // delete when unused
		n.Exclusive,             // exclusive
		n.NoWait,                // no-wait
		n.Args,                  // arguments
	)
	failOnError(err, "Failed to declare a queue")
	err = ch.QueueBind(
		q.Name,                  // queue name
		n.HandlerParams.Pattern, // routing key
		n.ExchangeKey,           // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msg, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		n.ConsumeAutoAck,
		n.ConsumeExclusive,
		n.ConsumeNoLocal,
		n.ConsumeNoWait,
		n.ConsumeArgs,
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msg {
			//response writer
			//queryParams
			//request type
			n.HandlerParams.Handler(d.Body, url.Values{}, n.HandlerParams.RequestType)
		}
	}()
	<-make(chan struct{})
}
