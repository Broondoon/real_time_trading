package networkQueue

import (
	"Shared/network"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueClusterInterface interface {
	QueueConnectionInterface
	SpawnQueue()
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
	println("New Queue Cluster")
	println("Exchange Key: ", exchangeKey)
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
	println("Exchange Key: ", n.ExchangeKey)
	exchangeParams.Name = n.ExchangeKey
	ch := n.SpawnChannel(exchangeParams)
	println("#######")
	println("Spawning Queue")
	println("ExchangeKey: ", n.ExchangeKey)
	println("QueueCluster: ", n.HandlerParams.Pattern)
	println("#######")
	defer n.CloseChannel(ch)
	q, err := ch.QueueDeclare(
		"",           // name
		n.Durable,    // durable
		n.AutoDelete, // delete when unused
		n.Exclusive,  // exclusive
		n.NoWait,     // no-wait
		n.Args,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	println("Queue: ", q.Name)
	err = ch.QueueBind(
		q.Name,                  // queue name
		n.HandlerParams.Pattern, // routing key
		n.ExchangeKey,           // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")
	println("Queue Bound: ", q.Name)

	msg, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		n.ConsumeExclusive,
		n.ConsumeNoLocal,
		n.ConsumeNoWait,
		n.ConsumeArgs,
	)
	failOnError(err, "Failed to register a consumer")
	println("Consumer registered: ", q.Name)
	go func() {
		for d := range msg {
			println("Received message")
			println("Message: ", string(d.Body))
			responseHandler := NewQueueResponseHandler(d, ch)
			data := QueueJSONData{}
			if json.Unmarshal(d.Body, &data) != nil {
				responseHandler.WriteHeader(http.StatusBadRequest)
				return
			}
			payload, err := json.Marshal(data.Payload)
			if err != nil {
				responseHandler.WriteHeader(http.StatusBadRequest)
				return
			}
			queryParams := url.Values{}
			for k, v := range data.Headers {
				queryParams.Add(k, v)
			}
			n.HandlerParams.Handler(responseHandler, payload, queryParams, data.MessageType)
		}
	}()
	<-make(chan struct{})
}

type QueueResponseHandler struct {
	d  amqp091.Delivery
	ch *amqp.Channel
}

func NewQueueResponseHandler(d amqp091.Delivery, ch *amqp.Channel) network.ResponseWriter {
	return &QueueResponseHandler{
		d:  d,
		ch: ch,
	}
}

func (n *QueueResponseHandler) WriteHeader(statusCode int) {
	println("Writing header: ", statusCode)
	switch statusCode {
	case http.StatusOK:
		// println("Acking")
		// n.d.Ack(false)
		n.Write([]byte("OK")) //Bad situation here, since we need to make a few adjustments to the response. We have to send back a body right now
	case http.StatusNotFound:
		n.d.Nack(false, false)
	case http.StatusBadRequest:
		n.d.Nack(false, false)
	case http.StatusInternalServerError:
		n.d.Nack(false, true)
	default:
		n.d.Nack(false, false)
	}
}

func (n *QueueResponseHandler) Write(body []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	println("Writing body: ", string(body))
	err := n.ch.PublishWithContext(
		ctx,
		"",
		n.d.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: n.d.CorrelationId,
			Body:          body,
		})

	if err != nil {
		println("Failed to publish response: ", err.Error())
		defer n.d.Nack(false, true)
		return http.StatusInternalServerError, err
	}
	println("Response published")
	defer n.d.Ack(false)
	return http.StatusOK, nil
}

func (n *QueueResponseHandler) Header() http.Header {
	header := http.Header{}
	for k, v := range n.d.Headers {
		jsonData, err := json.Marshal(v)
		if err != nil {
			panic("Failed to marshal header")
		}
		header.Add(k, string(jsonData))
	}
	return header
}
