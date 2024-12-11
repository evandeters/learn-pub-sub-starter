package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	ampq "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
    SimpleQueueDurable SimpleQueueType = iota
    SimpleQueueTransient
)

func DeclareAndBind(conn *ampq.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*ampq.Channel, ampq.Queue, error) {
    channel, err := conn.Channel()
    if err != nil {
        return nil, ampq.Queue{}, err
    }

    isDurable, autoDelete, isExclusive := false, false, false
    switch simpleQueueType {
    case SimpleQueueDurable:
        isDurable = true
    case SimpleQueueTransient:
        autoDelete = true
        isExclusive = true
    }
    queue, err := channel.QueueDeclare(queueName, isDurable, autoDelete, isExclusive, false, nil)
    if err != nil {
        return nil, ampq.Queue{}, err
    }

    err = channel.QueueBind(queueName, key, exchange, false, nil)
    if err != nil {
        return nil, ampq.Queue{}, err
    }

    return channel, queue, nil
}

func SubscribeJSON[T any](conn *ampq.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T)) error {
    channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
    if err != nil {
        log.Println("Unable to subscribe:", err)
    }

    msgChan, err := channel.Consume(queueName, "", false, false, false, false, nil)
    if err != nil {
        log.Println("Unable to consume queue:", err)
    }

    go func(c <-chan ampq.Delivery) {
        for msg := range(c) {
            body := new(T)
            if err := json.Unmarshal(msg.Body, body); err != nil {
                fmt.Println("Could not unmarshal message:", err)
                continue
            }
            handler(*body)
            msg.Ack(false)
        }
    }(msgChan)

    return nil
}
