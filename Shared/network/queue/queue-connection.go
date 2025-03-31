package networkQueue

import (
	"log"
	"os"
	"strconv"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueConnectionInterface interface {
	GetConnection() *amqp.Connection
	Connect()
	Disconnect()
	GetDialAddress() string
	SetDialAddress(address string)
	SpawnChannel(params ExchangeParams) *amqp.Channel
	CloseChannel(channel *amqp.Channel)
	//UpdateChannelQos(params QosParams)
}

type NetworkQueueConnection struct {
	DialAddress string
	connected   bool
	connection  *amqp.Connection
	channels    chan *amqp.Channel // buffered channel of channels
	mutex       sync.Mutex
}

type NewNetworkQueueConnectionParams struct {
	DialAddress string
	Connection  *amqp.Connection
}

func NewNetworkQueueConnection(params *NewNetworkQueueConnectionParams) QueueConnectionInterface {
	if params.DialAddress == "" {
		params.DialAddress = os.Getenv("BASE_DIAL_AMQP_URL")
	}
	maxChannelsStr := os.Getenv("MAX_CHANNELS_PER_CONNECTION_ENTITY")
	maxChannels, err := strconv.Atoi(maxChannelsStr)
	if err != nil {
		log.Printf("Invalid MAX_CHANNELS_PER_CONNECTION_ENTITY, using capacity of 0: %v", err)
		maxChannels = 0
	}

	n := &NetworkQueueConnection{
		connected:   false,
		DialAddress: params.DialAddress,
		connection:  params.Connection,
		channels:    make(chan *amqp.Channel, maxChannels),
	}
	n.Connect()
	for i := 0; i < cap(n.channels); i++ {
		ch, err := n.connection.Channel()
		if err == nil {
			ch.Qos(50, 0, false) // set a larger prefetch
			n.channels <- ch
		} else {
			// handle error if channel creation fails
		}
	}
	return n
}

// Acquire a channel
func (n *NetworkQueueConnection) GetChannel() *amqp.Channel {
	ch := <-n.channels
	return ch
}

func (b *NetworkQueueConnection) GetConnection() *amqp.Connection {
	log.Println("Getting connection")
	if !b.connected {
		b.Connect()
	}
	log.Println("Returning connection")
	return b.connection
}

func (b *NetworkQueueConnection) GetDialAddress() string {
	return b.DialAddress
}

func (b *NetworkQueueConnection) SetDialAddress(address string) {
	b.DialAddress = address
}

func (n *NetworkQueueConnection) Connect() {
	if n.connection != nil && n.connection.IsClosed() == false {
		n.connected = true
		return
	}
	conn, err := amqp.Dial(n.GetDialAddress())
	failOnError(err, "Failed to connect to RabbitMQ")
	n.connected = true
	n.connection = conn
}

func (n *NetworkQueueConnection) Disconnect() {
	n.connection.Close()
	n.connected = false
}

type ExchangeParams struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

func ExchangeParamsDefaults() ExchangeParams {
	return ExchangeParams{
		Name:       "",
		Type:       "topic",
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
}

func (n *NetworkQueueConnection) SpawnChannel(params ExchangeParams) *amqp.Channel {
	if !n.connected {
		n.Connect()
	}
	// if len(n.Channels) == cap(n.Channels) {
	// 	log.Printf("Max channels reached, not spawning new channel")
	// 	//provide random number to select a channel to return to caller
	// 	index := rand.Intn(len(n.Channels))
	// 	return n.Channels[index]
	// }

	ch := <-n.channels

	if params.Name == "" {
		panic("Not implemented correctly")
	}
	if params.Type == "" {
		params.Type = "topic"
	}

	err := ch.ExchangeDeclare(
		params.Name,
		params.Type,
		params.Durable,
		params.AutoDelete,
		params.Internal,
		params.NoWait,
		params.Args,
	)
	failOnError(err, "Failed to declare an exchange")

	return ch
}

func (n *NetworkQueueConnection) CloseChannel(channel *amqp.Channel) {
	// err := channel.Close()
	// failOnError(err, "Failed to close a channel")
	n.channels <- channel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("Queue %s: %s", msg, err)
		//log.Panicf("%s: %s", msg, err)
	}
}

type QosParams struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

// func (n *NetworkQueueConnection) UpdateChannelQos(params QosParams) {
// 	// if len(n.Channels) == 0 {
// 	// 	return
// 	// }
// 	if params.PrefetchCount == 0 {
// 		params.PrefetchCount = 1
// 	}

// 	for _, ch := range n.Channels {
// 		err := ch.Qos(
// 			params.PrefetchCount, // prefetch count
// 			params.PrefetchSize,  // prefetch size
// 			params.Global,        // global)
// 		)
// 		failOnError(err, "Failed to set QoS")
// 	}
// }
