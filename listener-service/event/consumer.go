package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer represents a RabbitMQ consumer
type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

// NewConsumer creates and sets up a new Consumer
func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

// setup prepares the consumer by creating a channel and declaring the exchange
func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

// Payload represents the structure of incoming messages
type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Listen starts consuming messages for the specified topics
func (consumer *Consumer) Listen(topics []string) error {
	// Create a channel for communication with RabbitMQ
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare a random queue for this consumer
	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	// Bind the queue to each topic
	for _, s := range topics {
		err = ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	// Start consuming messages
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Create a channel to keep the goroutine running
	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			// Process each message in a separate goroutine
			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for messages [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

// handlePayload processes different types of payloads
func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		// Log the payload
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}

	case "auth":
		// Authentication logic would go here

	default:
		// Log any unknown payload types
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

// logEvent sends the payload to a logging service
func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service:8080/log"

	// Create a new HTTP request
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the logging service accepted the request
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	return nil
}
