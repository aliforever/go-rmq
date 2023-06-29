package rmq_test

import (
	"fmt"
	"github.com/aliforever/go-rmq"
	"testing"
	"time"
)

func TestConsumerBuilder(t *testing.T) {
	r := rmq.New(amqpURL)

	closeChan, err := r.Connect(5, time.Second*5)
	if err != nil {
		panic(err)
	}

	ch, err := r.NewChannel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueBuilder().SetDurable().SetName("test_events").Declare()
	if err != nil {
		panic(err)
	}

	fmt.Printf("queue %q declared\n", queue.Name())

	consumer, err := ch.ConsumerBuilder("consumer1", queue.Name()).SetAutoAck().Build()
	if err != nil {
		panic(err)
	}
	defer consumer.Cancel()

	for delivery := range consumer.Messages() {
		fmt.Println(string(delivery.Body))
	}

	<-closeChan
}

// func TestConsumer_ConsumeWithResponses(t *testing.T) {
// 	conn, err := initRabbitMqConnectionChannel()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
//
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer ch.Close()
//
// 	queueName := "events"
//
// 	queue, err := rmq.NewQueueBuilderWithChannel(ch).Declare()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	responseQueue := &genericSync.Map[chan amqp091.Delivery]{}
//
// 	go consumer(conn, queue.Name)
// 	// go consumer2(conn, queue.Name, responseQueue)
//
// 	publisher, err := rmq.newPublisher(conn, "", queueName)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	response, err := publisher.WithFields(rmq.NewPublishFields().
// 		SetDataTypeBytes().
// 		SetCorrelationID("1").
// 		SetResponseTimeout(time.Second*2)).New().
// 		PublishAwaitResponse(context.Background(), []byte("Hello 2"), responseQueue)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	err = response.Ack(false)
// 	if err != nil {
// 		panic(err)
// 	}
// }
//
// func consumer(conn *amqp091.Connection, queueName string) {
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer ch.Close()
//
// 	c, err := rmq.newConsumerBuilder(conn, "", queueName)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	consumer, err := c.Consume()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	publisher := rmq.NewPublisherWithChannel(ch, "", queueName)
//
// 	clientQueue := fmt.Sprintf("%s_client", queueName)
// 	fmt.Println(clientQueue)
// 	for delivery := range consumer {
// 		delivery.Ack(false)
// 		if delivery.CorrelationId != "" {
// 			err = publisher.New().WithFields(rmq.NewPublishFields().
// 				SetReplyToID(delivery.CorrelationId).
// 				SetDataTypeBytes()).
// 				Publish(context.Background(), []byte("Got you!"))
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// }
