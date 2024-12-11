package main

import (
	"fmt"
	"log"

	"github.com/evandeters/learn-pub-sub-starter/internal/gamelogic"
	"github.com/evandeters/learn-pub-sub-starter/internal/pubsub"
	"github.com/evandeters/learn-pub-sub-starter/internal/routing"

	ampq "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
    
    connString := "amqp://guest:guest@localhost:5672/"
    
    conn, err := ampq.Dial(connString)
    if err != nil {
        log.Fatalln("Failed to connect to RabbitMQ:", err)
    }
    defer conn.Close()

    username, err := gamelogic.ClientWelcome()
    if err != nil {
        log.Fatalln("Failed to get username:", err)
    }

    queueName := routing.PauseKey + "." + username
    _, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueTransient)
    if err != nil {
        log.Fatalln("Failed to create Queue:", err)
    }

    gs := gamelogic.NewGameState(username)
    err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs))
    if err != nil {
        log.Fatalln("Could not subscribe to Game State:", err)
    }


    moveKey := routing.ArmyMovesPrefix + "." + username
    moveRoutingKey := routing.ArmyMovesPrefix + ".*"
    moveChan, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, moveKey, moveRoutingKey, pubsub.SimpleQueueTransient)
    if err != nil {
        log.Fatalln("Could not bind to move routing key:", err)
    }

    err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, moveKey, moveRoutingKey, pubsub.SimpleQueueTransient, handlerMove(gs))
    if err != nil {
        log.Fatalln("Could not subscribe to Move Queue:", err)
    }

    for {
        input := gamelogic.GetInput()

        if len(input) == 0 {
            continue
        }

        switch input[0] {
        case "spawn":
            err := gs.CommandSpawn(input)
            if err != nil {
                fmt.Println("Failed to spawn unit:", err)
            }
        case "move":
            armyMove, err := gs.CommandMove(input)
            if err != nil {
                fmt.Println("Failed to move unit:", err)
            } else {
                err = pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, moveKey, armyMove)
                fmt.Println("Move was successful!")
            }
        case "status":
            gs.CommandStatus()
        case "help":
            gamelogic.PrintClientHelp()
        case "spam":
            fmt.Println("Spamming not allowed yet!")
        case "quit":
            gamelogic.PrintQuit()
            return
        default:
            fmt.Println("ERROR: Unknown command.")
        }
    }
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
    return func(ps routing.PlayingState) {
        defer fmt.Print("> ")
        gs.HandlePause(ps)
    }
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
    return func(am gamelogic.ArmyMove) {
        defer fmt.Print("> ")
        gs.HandleMove(am)
    }
}
