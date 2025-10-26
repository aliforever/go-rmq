package rmq

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

var (
	ErrConnectionClosed        = fmt.Errorf("connection closed")
	fileUpdateUnmarshalErrText = "error on file update unmarshal"
	ackErrText                 = "error on ack file update"
)

type OnUpdateOptions struct {
	prefetchCount          int
	returnOnUnmarshalError bool
	returnOnHandlerError   bool
	createQueue            bool
	durableQueue           bool
	logger                 *slog.Logger
}

func NewOnUpdateOptions() *OnUpdateOptions {
	return &OnUpdateOptions{
		prefetchCount:          10,
		createQueue:            true,
		returnOnUnmarshalError: false,
		returnOnHandlerError:   true,
		durableQueue:           true,
		logger:                 nil,
	}
}

func (oup *OnUpdateOptions) SetCreateQueue(createQueue, isDurable bool) *OnUpdateOptions {
	oup.createQueue = createQueue
	oup.durableQueue = isDurable
	return oup
}

func (oup *OnUpdateOptions) SetPrefetchCount(count int) *OnUpdateOptions {
	oup.prefetchCount = count
	return oup
}

func (oup *OnUpdateOptions) SetReturnOnUnmarshalError(returnOnUnmarshalError bool) *OnUpdateOptions {
	oup.returnOnUnmarshalError = returnOnUnmarshalError
	return oup
}

func (oup *OnUpdateOptions) SetReturnOnHandlerError(returnOnHandlerError bool) *OnUpdateOptions {
	oup.returnOnHandlerError = returnOnHandlerError
	return oup
}

func (oup *OnUpdateOptions) SetLogger(logger *slog.Logger) *OnUpdateOptions {
	oup.logger = logger
	return oup
}

func OnUpdate[T any](
	client *RMQ,
	queueName string,
	prefetchCount int,
	handler func(*T) error,
	opts *OnUpdateOptions,
) error {
	if opts == nil {
		opts = NewOnUpdateOptions()
	}

	consumerChannel, err := client.NewChannel()
	if err != nil {
		return err
	}

	consumer, err := consumerChannel.
		ConsumerBuilder("", queueName).
		SetPrefetch(prefetchCount).
		Build()
	if err != nil {
		return err
	}

	if opts.createQueue {
		builder := consumerChannel.
			QueueBuilder().
			SetName(queueName)

		if opts.durableQueue {
			builder.SetDurable()
		}

		_, err = builder.Declare()
		if err != nil {
			return err
		}
	}

	for delivery := range consumer.Messages() {
		var message *T

		if err = json.Unmarshal(delivery.Body, &message); err != nil {
			logErrorOnNotNil(
				opts.logger,
				fileUpdateUnmarshalErrText,
				slog.String("body", string(delivery.Body)),
				slog.Any("error", err),
			)

			if opts.returnOnUnmarshalError {
				return fmt.Errorf("couldnt unmarshal payload: %s", err)
			}

			continue
		}

		if err = handler(message); err != nil {
			logErrorOnNotNil(
				opts.logger,
				fmt.Sprintf("error on %s queue", queueName),
				slog.Any("error", err),
			)

			if opts.returnOnHandlerError {
				return fmt.Errorf("error on %s queue: %s", queueName, err)
			}

			continue
		}

		if err = delivery.Ack(false); err != nil {
			logErrorOnNotNil(
				opts.logger,
				ackErrText,
				slog.Any("error", err),
			)

			return fmt.Errorf("error on ack file update: %s", err)
		}
	}

	return ErrConnectionClosed
}

func logErrorOnNotNil(logger *slog.Logger, msg string, args ...any) {
	if logger == nil {
		return
	}

	logger.Error(
		msg,
		args...,
	)
}
