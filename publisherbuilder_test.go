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

var amqpURL = "amqp://guest:guest@localhost:5672/"

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
	r := rmq.New(amqpURL)

	closeChan, err := r.Connect(5, time.Second*6, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("connected to %q\n", amqpURL)

	fmt.Println("creating channel")
	ch, err := r.NewChannel()
	if err != nil {
		panic(err)
	}
	fmt.Println("channel created")

	fmt.Println("sleeping for 10 seconds")
	time.Sleep(time.Second * 10)

	pb := ch.PublisherBuilder("", "test_events").
		WithFields(rmq.NewPublishFields().
			SetDataTypeBytes().
			SetExpiration(time.Second * 5),
		).New()

	err = pb.Publish(context.Background(), []byte("Hello there"))
	if err != nil {
		panic(err)
	}

	go r.Close()

	<-closeChan
}

// func TestPublisher_Publish2(t *testing.T) {
// 	conn, err := initRabbitMqConnectionChannel()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	err = ch.Confirm(false)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	np := ch.NotifyReturn(make(chan amqp091.Return))
// 	np1 := ch.NotifyCancel(make(chan string))
// 	np2 := ch.NotifyClose(make(chan *amqp091.Error))
// 	np3 := ch.NotifyPublish(make(chan amqp091.Confirmation))
//
// 	go func() {
// 		for confirmation := range np {
// 			fmt.Println(confirmation, "Return")
// 		}
// 	}()
//
// 	go func() {
// 		for confirmation := range np1 {
// 			fmt.Println(confirmation, "Cancel")
// 		}
// 	}()
//
// 	go func() {
// 		for confirmation := range np2 {
// 			fmt.Println(confirmation, "Close")
// 		}
// 	}()
//
// 	go func() {
// 		for confirmation := range np3 {
// 			fmt.Println(confirmation, "Publish")
// 		}
// 	}()
//
// 	err = rmq.NewPublisherWithChannel(ch, "", "").New().
// 		WithFields(rmq.NewPublishFields().
// 			SetDataTypeBytes().
// 			SetExpiration(time.Second*5)).
// 		Publish(context.Background(), []byte("Hello there"))
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	time.Sleep(time.Second * 5)
// }
