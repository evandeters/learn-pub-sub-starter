package pubsub

import (
	"context"
	"encoding/json"

	ampq "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *ampq.Channel,exchange, key string, val T) error {
    valJson, err := json.Marshal(val)
    if err != nil {
        return err
    }

    ctx := context.Background()
    err = ch.PublishWithContext(ctx, exchange, key, false, false, ampq.Publishing{ContentType: "application/json", Body: valJson})
    if err != nil {
        return err
    }

    return nil
}

