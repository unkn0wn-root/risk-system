// Package messaging provides RabbitMQ-based message publishing and consumption capabilities.
package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQ wraps a RabbitMQ connection and channel for message operations.
// It provides methods for queue management, message publishing, and consumption.
type RabbitMQ struct {
	conn    *amqp.Connection // RabbitMQ connection
	channel *amqp.Channel    // RabbitMQ channel for operations
}

// NewRabbitMQ creates a new RabbitMQ client instance and establishes connection.
// It opens both a connection and a channel for message operations.
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

// DeclareQueue creates a durable queue with the specified name if it doesn't exist.
// The queue is configured to survive broker restarts but not exclusive to this connection.
func (r *RabbitMQ) DeclareQueue(name string) error {
	_, err := r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// Publish sends a message to the specified queue after JSON marshaling.
// The message is published with JSON content type and logged for debugging.
func (r *RabbitMQ) Publish(queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to queue %s: %s", queueName, string(body))
	return nil
}

// Consume starts consuming messages from the specified queue with auto-acknowledgment.
// It processes messages using the provided handler function and logs message receipts.
func (r *RabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received message from queue %s: %s", queueName, string(d.Body))
			if err := handler(d.Body); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()

	log.Printf("Waiting for messages from queue %s. To exit press CTRL+C", queueName)
	<-forever

	return nil
}

// Close properly closes the RabbitMQ channel and connection.
// It should be called when the RabbitMQ client is no longer needed to prevent resource leaks.
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
