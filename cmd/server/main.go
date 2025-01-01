package main

import (
	"fmt"
	"log"

	"github.com/evandeters/learn-pub-sub-starter/internal/gamelogic"
	"github.com/evandeters/learn-pub-sub-starter/internal/pubsub"
	"github.com/evandeters/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connString := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalln("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()
	fmt.Println("Connection to RabbitMQ was successful!")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalln("Failed to create channel:", err)
	}

	err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Println("Failed to publish message:", err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, "game_logs", routing.GameLogSlug, pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalln("Failed to create game_logs topic")
	}

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Game is pausing")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Println("Failed to pause game:", err)
			}
		case "resume":
			fmt.Println("Game is resuming")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Println("Failed to resume game:", err)
			}
		case "quit":
			fmt.Println("Game is stopping")
			return
		default:
			fmt.Println("I don't understand that command")
		}

	}
}
