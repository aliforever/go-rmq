package rmq_test

import (
	"context"
	"fmt"
	"github.com/aliforever/go-rmq"
	"github.com/rabbitmq/amqp091-go"
	"os"
	"testing"
	"time"
)

const envAMQPURLName = "AMQP_URL"

var amqpURL = "amqp://admin:admin@127.0.0.1:5672/"

func init() {
	url := os.Getenv(envAMQPURLName)
	if url != "" {
		amqpURL = url
		return
	}

	fmt.Printf("environment variable envAMQPURLName undefined or empty, using default: %q\n", amqpURL)
}

func initRabbitMqConnectionChannel() (*amqp091.Connection, error) {
	return amqp091.Dial(amqpURL)
}

func TestPublisher_Publish(t *testing.T) {
	conn, err := initRabbitMqConnectionChannel()
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	err = ch.Confirm(false)
	if err != nil {
		panic(err)
	}

	np := ch.NotifyReturn(make(chan amqp091.Return))
	np1 := ch.NotifyCancel(make(chan string))
	np2 := ch.NotifyClose(make(chan *amqp091.Error))
	np3 := ch.NotifyPublish(make(chan amqp091.Confirmation))

	go func() {
		for confirmation := range np {
			fmt.Println(confirmation, "Return")
		}
	}()

	go func() {
		for confirmation := range np1 {
			fmt.Println(confirmation, "Cancel")
		}
	}()

	go func() {
		for confirmation := range np2 {
			fmt.Println(confirmation, "Close")
		}
	}()

	go func() {
		for confirmation := range np3 {
			fmt.Println(confirmation, "Publish")
		}
	}()

	err = rmq.NewPublisherWithChannel(ch, "", "").New().
		WithFields(rmq.NewPublishFields().
			SetDataTypeBytes().
			SetExpiration(time.Second*5)).
		Publish(context.Background(), []byte("Hello there"))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 5)
}
