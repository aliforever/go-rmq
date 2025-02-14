package main

import (
	"fmt"
	"time"

	"github.com/aliforever/go-rmq"
)

func main() {
	client := rmq.New("amqp://guest:guest@localhost:1884/")

	closeChan, err := client.Connect(2, time.Second*5, func(err error) {
		fmt.Println("Err", err)
	})
	if err != nil {
		panic(err)
	}

	go func() {
		chn, err := client.NewChannel()
		if err != nil {
			panic(err)
		}

		q, err := chn.QueueBuilder().SetName("test_events").Declare()
		if err != nil {
			panic(err)
		}

		fmt.Println(q.ConsumerCount())

		consumer, err := chn.ConsumerBuilder("consumer1", q.Name()).SetAutoAck().Build()
		if err != nil {
			panic(err)
		}

		fmt.Println("reading messages")

		for delivery := range consumer.Messages() {
			fmt.Println(string(delivery.Body))
		}

		fmt.Println("consumer closed")
	}()

	err = <-closeChan

	fmt.Println("client closed", err)
}
