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

	queue, err := rmq.NewQueueBuilder(ch, queueName).Declare()
	if err != nil {
		panic(err)
	}

	responseQueue := &genericSync.Map[rmq.DeliveryChannel]{}

	go consumer(conn, queue.Name)
	// go consumer2(conn, queue.Name, responseQueue)

	response, err := rmq.NewPublisherWithChannel(conn, "", queueName).
		SetDataTypeBytes().
		SetCorrelationID("1").
		SetResponseTimeout(time.Second*2).
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

	clientQueue := fmt.Sprintf("%s_client", queueName)
	fmt.Println(clientQueue)
	for delivery := range consumer {
		delivery.Ack(false)
		if delivery.CorrelationId != "" {
			err = rmq.NewPublisherWithChannel(conn, "", queueName).
				SetReplyToID(delivery.CorrelationId).
				SetDataTypeBytes().
				Publish(context.Background(), []byte("Got you!"))
			if err != nil {
				panic(err)
			}
		}
	}
}

func consumerClient(conn *amqp091.Connection, queueName string, responseQueue *genericSync.Map[rmq.DeliveryChannel]) {
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

	for delivery := range consumer {
		delivery.Ack(false)
		if delivery.CorrelationId != "" {
			err = rmq.NewPublisherWithChannel(conn, "", queueName).
				SetReplyToID(delivery.CorrelationId).
				SetDataTypeBytes().
				Publish(context.Background(), []byte("Got you!"))
			if err != nil {
				panic(err)
			}
		}
	}
}
