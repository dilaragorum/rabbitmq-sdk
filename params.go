package rabbitmq_sdk

import amqp "github.com/rabbitmq/amqp091-go"

type QueueParam struct {
	name       string
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	args       amqp.Table
}

type ExchangeParam struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       amqp.Table
}

type QueueBindingParam struct {
	name     string
	key      string
	exchange string
	noWait   bool
	args     amqp.Table
}

type ConsumeParam struct {
	queue     string
	consumer  string
	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table
}
