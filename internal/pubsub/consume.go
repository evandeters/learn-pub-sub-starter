package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type Acktype int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	isDurable, autoDelete, isExclusive := false, false, false
	switch simpleQueueType {
	case SimpleQueueDurable:
		isDurable = true
	case SimpleQueueTransient:
		autoDelete = true
		isExclusive = true
	}

	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := channel.QueueDeclare(queueName, isDurable, autoDelete, isExclusive, false, args)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) Acktype) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer channel.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				msg.Ack(false)
				log.Println("Message acknowleged")
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("NackRequeue")
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("Message Nack and discard")
			}
		}
	}()
	return nil
}
