package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// declareExchange sets up the topic exchange
//
// An exchange in RabbitMQ is a message routing agent.
// It's essentially the middleman between the publisher (who sends messages) and the queues (where messages are stored).
//
// Key points about exchanges:
//
// Publishers send messages to exchanges, not directly to queues.
// Exchanges distribute message copies to queues based on rules called "bindings".
// There are several types of exchanges (direct, topic, fanout, headers), each with different routing rules.
// Here, we're using a "topic" exchange named "logs_topic".
func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name of the exchange
		"topic",      // type of the exchange
		true,         // durable (survives broker restarts)
		false,        // auto-deleted when no queues are bound to it
		false,        // internal (can't be used directly by publishers)
		false,        // no-wait (don't wait for a server confirmation)
		nil,          // arguments (optional)
	)
}

// declareRandomQueue declares a random, exclusive queue
func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",    // name (empty string means generate a random name)
		false, // durable (queue won't survive broker restarts)
		false, // delete when unused
		true,  // exclusive (only accessible by this connection)
		false, // no-wait
		nil,   // arguments
	)
}

/**
A channel in RabbitMQ is a virtual connection inside a real TCP connection. It's like a lightweight socket inside the connection.
Key points about channels:

Channels allow you to multiplex a connection, performing multiple operations concurrently.
Most operations in the RabbitMQ API are performed on channels, not on the connection itself.
Channels are less expensive to open and close than connections.
In our code, we create a channel with consumer.conn.Channel().

Think of a channel as a "virtual connection" that allows you to interact with RabbitMQ without the overhead of creating multiple TCP connections.
*/
