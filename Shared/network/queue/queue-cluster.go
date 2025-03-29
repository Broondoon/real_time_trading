package networkQueue

import (
	"Shared/network"
	"context"
	"encoding/json"
	"log"
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
	QueueName        string
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
	log.Println("New Queue Cluster")
	log.Println("Exchange Key: ", exchangeKey)
	return &QueueCluster{
		QueueConnectionInterface: NewNetworkQueueConnection(params.NewNetworkQueueConnectionParams),
		HandlerParams:            handler,
		ExchangeKey:              exchangeKey,
		Durable:                  params.Durable,
		AutoDelete:               params.AutoDelete,
		Exclusive:                params.Exclusive,
		NoWait:                   params.NoWait,
		Args:                     params.Args,
		QueueName:                handler.Pattern,
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
	// log.Println("#######")
	// log.Println("Spawning Queue")
	// log.Println("ExchangeKey: ", n.ExchangeKey)
	// log.Println("QueueCluster: ", n.HandlerParams.Pattern)
	// log.Println("#######")
	defer n.CloseChannel(ch)
	q, err := ch.QueueDeclare(
		n.QueueName+"_Queue", // name. We set this to make sure that any services sharing this queue all use the same queue, rather than declaring their own, which causes duplication issues.
		n.Durable,            // durable
		n.AutoDelete,         // delete when unused
		n.Exclusive,          // exclusive
		n.NoWait,             // no-wait
		n.Args,               // arguments
	)
	failOnError(err, "Failed to declare a queue")
	//log.Println("Queue: ", q.Name)
	err = ch.QueueBind(
		q.Name,                  // queue name
		n.HandlerParams.Pattern, // routing key
		n.ExchangeKey,           // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")
	// log.Println("Queue Bound: ", q.Name)

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
	//log.Println("Consumer registered: ", q.Name)
	go func() {
		for d := range msg {
			// log.Println("Received message")
			// log.Println("Message: ", string(d.Body))
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
	d          amqp091.Delivery
	ch         *amqp.Channel
	Completed  bool
	statusCode int
}

func NewQueueResponseHandler(d amqp091.Delivery, ch *amqp.Channel) network.ResponseWriter {
	return &QueueResponseHandler{
		d:          d,
		ch:         ch,
		Completed:  false,
		statusCode: http.StatusOK,
	}
}

func (n *QueueResponseHandler) WriteHeader(statusCode int) {
	if !n.CheckCompleted() {
		n.statusCode = statusCode
		//log.Println("Writing header: ", statusCode)
		switch statusCode {
		case http.StatusOK:
			// log.Println("Acking")
			// n.d.Ack(false)
			n.Write([]byte("OK")) //Bad situation here, since we need to make a few adjustments to the response. We have to send back a body right now
		case http.StatusNotFound:
			log.Println("Not found on: ", n.d.ReplyTo)
			n.d.Nack(false, false)
		case http.StatusBadRequest:
			log.Println("Bad Request on: ", n.d.ReplyTo)
			n.d.Nack(false, false)
		case http.StatusInternalServerError:
			log.Println("Internal Service error on: ", n.d.ReplyTo)
			n.d.Nack(false, false)
		default:
			log.Println("Unknown on: ", n.d.ReplyTo)
			n.d.Nack(false, false)
		}
		n.Completed = true
	}
}

func (n *QueueResponseHandler) Write(body []byte) (int, error) {
	if !n.CheckCompleted() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// log.Println("Writing body: ", string(body))
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
			log.Println("Failed to publish response: ", err.Error())
			defer n.d.Nack(false, true)
			n.statusCode = http.StatusInternalServerError
			return http.StatusInternalServerError, err
		}
		// log.Println("Response published")
		defer n.d.Ack(false)
		n.Completed = true
		n.statusCode = http.StatusOK
		return http.StatusOK, nil
	}
	return n.statusCode, nil
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

func (n *QueueResponseHandler) EncodeResponse(statusCode int, data map[string]interface{}) {
	if !n.CheckCompleted() {
		jsonData, err := json.Marshal(data)
		if err != nil {
			n.WriteHeader(http.StatusInternalServerError)
			return
		}
		n.WriteHeader(statusCode)
		n.Write(jsonData)
	}
}

func (n *QueueResponseHandler) CheckCompleted() bool {
	return n.Completed
}

func (n *QueueResponseHandler) GetStatusCode() int {
	return n.statusCode
}
