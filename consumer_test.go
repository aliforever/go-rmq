package rmq_test

import (
	"context"
	"fmt"
	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/aliforever/go-rmq"
	"github.com/rabbitmq/amqp091-go"
	"testing"
	"time"
)

func TestConsumer_ConsumeWithResponses(t *testing.T) {
	conn, err := initRabbitMqConnectionChannel()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	queueName := "events"

	queue, err := rmq.NewQueueBuilder(ch).Declare()
	if err != nil {
		panic(err)
	}

	responseQueue := &genericSync.Map[chan amqp091.Delivery]{}

	go consumer(conn, queue.Name)
	// go consumer2(conn, queue.Name, responseQueue)

	publisher, err := rmq.NewPublisher(conn, "", queueName)
	if err != nil {
		panic(err)
	}

	response, err := publisher.WithFields(rmq.NewPublishFields().
		SetDataTypeBytes().
		SetCorrelationID("1").
		SetResponseTimeout(time.Second*2)).New().
		PublishAwaitResponse(context.Background(), []byte("Hello 2"), responseQueue)
	if err != nil {
		panic(err)
	}

	err = response.Ack(false)
	if err != nil {
		panic(err)
	}
}

func consumer(conn *amqp091.Connection, queueName string) {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	consumer, err := rmq.NewConsumer(ch, "", queueName).
		Consume()
	if err != nil {
		panic(err)
	}

	publisher := rmq.NewPublisherWithChannel(ch, "", queueName)

	clientQueue := fmt.Sprintf("%s_client", queueName)
	fmt.Println(clientQueue)
	for delivery := range consumer {
		delivery.Ack(false)
		if delivery.CorrelationId != "" {
			err = publisher.New().WithFields(rmq.NewPublishFields().
				SetReplyToID(delivery.CorrelationId).
				SetDataTypeBytes()).
				Publish(context.Background(), []byte("Got you!"))
			if err != nil {
				panic(err)
			}
		}
	}
}
