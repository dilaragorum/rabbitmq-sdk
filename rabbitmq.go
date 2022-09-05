package rabbitmq_sdk

import (
	"github.com/labstack/gommon/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	connURL      string
	errCh        <-chan *amqp.Error
	messageChan  <-chan amqp.Delivery
	retryAttempt int
}

type RabbitMQOptions struct {
	URL          string
	RetryAttempt int
}

func NewRabbitMQ(options RabbitMQOptions) (*RabbitMQ, error) {
	rabbitMQ := &RabbitMQ{
		connURL:      options.URL,
		retryAttempt: options.RetryAttempt,
	}

	if err := rabbitMQ.connect(); err != nil {
		return nil, err
	}

	return rabbitMQ, nil
}

func (rmq *RabbitMQ) connect() error {
	var err error

	rmq.connection, err = amqp.Dial(rmq.connURL)
	if err != nil {
		log.Error("Error when creating connection", err.Error())
		return err
	}

	rmq.channel, err = rmq.connection.Channel()
	if err != nil {
		log.Error("Error when creating channel", err.Error())
		return err
	}

	rmq.errCh = rmq.connection.NotifyClose(make(chan *amqp.Error))

	return nil
}

func (rmq *RabbitMQ) reconnect() error {
	attempt := rmq.retryAttempt

	for attempt != 0 {
		log.Info("Attempting rabbitmq reconnection")
		if err := rmq.connect(); err != nil {
			attempt--
			log.Error("Rabbitmq retry connection error", err.Error())
			continue
		}
		return nil
	}

	if attempt == 0 {
		log.Fatal("Rabbitmq retry connection is failed")
	}

	return nil
}

func (rmq *RabbitMQ) Close() {
	rmq.channel.Close()
	rmq.connection.Close()
}

func (rmq *RabbitMQ) ConsumeMessageChannel() (jsonBytes []byte, err error) {
	select {
	case err := <-rmq.errCh:
		log.Warn("Rabbitmq comes error from notifyCloseChan", err.Error())
		rmq.reconnect()
	case msg := <-rmq.messageChan:
		return msg.Body, nil
	}

	return nil, nil
}

func (rmq *RabbitMQ) CreateQueue(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args map[string]interface{}) (amqp.Queue, error) {
	queue, err := rmq.channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
	if err != nil {
		log.Error("Error when creating queue", err.Error())
		return amqp.Queue{}, err
	}

	return queue, nil
}

func (rmq *RabbitMQ) CreateExchange(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args map[string]interface{}) error {
	if err := rmq.channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args); err != nil {
		log.Error("Error when creating exchange", err.Error())
		return err
	}

	return nil
}

func (rmq *RabbitMQ) BindExchangeWithQueue(name string, key string, exchange string, noWait bool, args map[string]interface{}) error {
	if err := rmq.channel.QueueBind(name, key, exchange, noWait, args); err != nil {
		log.Error("Error when queue binding", err.Error())
		return err
	}

	return nil
}

func (rmq *RabbitMQ) CreateMessageChannel(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args map[string]interface{}) error {
	var err error
	rmq.messageChan, err = rmq.channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if err != nil {
		log.Error("Error when consuming message", err.Error())
		return err
	}

	return nil
}
