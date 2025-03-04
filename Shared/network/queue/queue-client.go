package networkQueue

import (
	"Shared/network"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueClientInterface interface {
	QueueConnectionInterface
	network.ClientInterface
	SendWithReturn(route string, message []byte, params SendParams, onReturn func([]byte) ([]byte, error)) ([]byte, error)
}

type QueueClient struct {
	QueueConnectionInterface
	ExchangeRoute string
}

type NewQueueClientParams struct {
	*NewNetworkQueueConnectionParams
}

func NewQueueClient(exchangeRoute string, params *NewQueueClientParams) QueueClientInterface {
	if params.NewNetworkQueueConnectionParams == nil {
		params.NewNetworkQueueConnectionParams = &NewNetworkQueueConnectionParams{}
	}
	return &QueueClient{
		QueueConnectionInterface: NewNetworkQueueConnection(params.NewNetworkQueueConnectionParams),
		ExchangeRoute:            exchangeRoute,
	}
}

type SendParams struct {
	Mandatory bool
	Immediate bool
	Timeout   time.Duration
}

func DefaultPublishParams() SendParams {
	return SendParams{
		Mandatory: false,
		Immediate: false,
		Timeout:   5 * time.Second,
	}
}

func (n *QueueClient) SendWithReturn(route string, message []byte, params SendParams, onReturn func([]byte) ([]byte, error)) ([]byte, error) {
	exchangeParams := ExchangeParams{
		Name: n.ExchangeRoute,
	}
	ch := n.SpawnChannel(exchangeParams)
	if params.Timeout == 0 {
		params.Timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()
	defer n.CloseChannel(ch)
	returnQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a return queue")
	msg, err := ch.Consume(
		returnQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")
	corrID := generateRandomID()
	err = ch.PublishWithContext(
		ctx,
		n.ExchangeRoute,
		route,
		params.Mandatory,
		params.Immediate,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			ReplyTo:       returnQueue.Name,
			Body:          message,
		})
	failOnError(err, "Failed to publish a message")
	for d := range msg {
		if corrID == d.CorrelationId {
			println("Response received")
			return onReturn(d.Body)
		}
	}
	println("No response received")
	return nil, nil
}

// need a RPC
func (n *QueueClient) Get(route string, headers map[string]string) ([]byte, error) {
	data := QueueJSONData{
		Headers:     headers,
		MessageType: "GET",
		Payload:     nil,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return n.SendWithReturn(route, jsonData, DefaultPublishParams(), func(response []byte) ([]byte, error) {
		return response, nil
	})
}

func (n *QueueClient) Post(route string, payload interface{}) ([]byte, error) {
	data := QueueJSONData{
		Headers:     nil,
		MessageType: "POST",
		Payload:     payload,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		println("Error marshalling payload")
		return nil, err
	}
	return n.SendWithReturn(route, jsonData, DefaultPublishParams(), func(response []byte) ([]byte, error) {
		return response, nil
	})
}

func (n *QueueClient) Put(route string, payload interface{}) ([]byte, error) {
	data := QueueJSONData{
		Headers:     nil,
		MessageType: "PUT",
		Payload:     payload,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		println("Error marshalling payload")
		return nil, err
	}
	return n.SendWithReturn(route, jsonData, DefaultPublishParams(), func(response []byte) ([]byte, error) {
		return response, nil
	})
}

func (n *QueueClient) Delete(route string) ([]byte, error) {
	//header will have id as the last part of the route
	id := strings.Split(route, "/")[len(strings.Split(route, "/"))-1]
	headers := map[string]string{"id": id}
	data := QueueJSONData{
		Headers:     headers,
		MessageType: "DELETE",
		Payload:     nil,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		println("Error marshalling payload")
		return nil, err
	}
	return n.SendWithReturn(route, jsonData, DefaultPublishParams(), func(response []byte) ([]byte, error) {
		return response, nil
	})
}

func generateRandomID() string {
	// Generate a new UUID as the stock ID
	return uuid.New().String()
}

type QueueJSONData struct {
	Headers     map[string]string `json:"headers"`
	MessageType string            `json:"messageType"`
	Payload     interface{}       `json:"payload"`
}
