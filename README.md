# go-rmq
RabbitMQ Wrappers for [amqp091-go](github.com/rabbitmq/amqp091-go)

## Features:
- Using builder pattern to make it much easier working with exchanges, queues, consumers, etc...
- KeepAlive functionality to persist the connection upon temporary failures

## Install
`go get -u github.com/aliforever/go-rmq`

## Usage:
To initialize the connection with 5 retry times and 10 seconds of delay on each try:
```go
r := rmq.New(address)

errChan, err := r.Connect(5, 10*time.Second)
if err != nil {
    return nil, err
}

panic(<-errChan)
```
