package rmq

import (
	"context"
	"fmt"
	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type Publisher struct {
	ch         *Channel
	exchange   string
	routingKey string

	retryCount int
	retryDelay time.Duration

	fields *PublishFields
}

func newPublisher(
	ch *Channel,
	retryCount int,
	retryDelay time.Duration,
	exchange string,
	routingKey string,
	fields *PublishFields,
) *Publisher {
	return &Publisher{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
		retryCount: retryCount,
		retryDelay: retryDelay,
		fields:     fields,
	}
}

func (p *Publisher) WithFields(fields *PublishFields) *Publisher {
	p.fields = fields

	return p
}

func (p *Publisher) Fields() *PublishFields {
	return p.fields
}

func (p *Publisher) Publish(ctx context.Context, data interface{}) error {
	body, err := p.fields.makeData(data)
	if err != nil {
		return err
	}

	retried := p.retryCount

	for retried > 0 {
		err = p.ch.channel().PublishWithContext(ctx,
			p.exchange,
			p.routingKey,
			p.fields.mandatory,
			p.fields.immediate,
			amqp091.Publishing{
				Headers:       p.fields.headers,
				ContentType:   p.fields.contentType,
				DeliveryMode:  p.fields.deliveryMode,
				CorrelationId: p.fields.correlationID,
				ReplyTo:       p.fields.replyToID,
				Expiration:    p.fields.expiration,
				Body:          body,
			})
		if err != nil {
			retried--
			time.Sleep(p.retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to publish message after %d retries: %s", p.retryCount, err)
}

// PublishAwaitResponse creates a channel and stores it in responseMap
//
//	If an outsider writes the response to that map it checks for the reply to id to see if it matches the correlation id
//	Fails upon an invalid correlation id or in case of a timeout
func (p *Publisher) PublishAwaitResponse(
	ctx context.Context,
	data interface{},
	responseMap *genericSync.Map[chan amqp091.Delivery],
) (amqp091.Delivery, error) {
	if responseMap == nil {
		return amqp091.Delivery{}, ResponseMapNotSetError
	}

	if p.fields.correlationID == "" {
		return amqp091.Delivery{}, CorrelationIdNotSetError
	}

	ch := make(chan amqp091.Delivery)

	responseMap.Store(p.fields.correlationID, ch)

	body, err := p.fields.makeData(data)
	if err != nil {
		return amqp091.Delivery{}, err
	}

	tried := p.retryCount

	for tried > 0 {
		err = p.ch.channel().PublishWithContext(
			ctx,
			p.exchange,
			p.routingKey,
			p.fields.mandatory,
			p.fields.immediate,
			amqp091.Publishing{
				Headers:       p.fields.headers,
				ContentType:   p.fields.contentType,
				DeliveryMode:  p.fields.deliveryMode,
				CorrelationId: p.fields.correlationID,
				ReplyTo:       p.fields.replyToID,
				Expiration:    p.fields.expiration,
				Body:          body,
			})
		if err != nil {
			tried--
			time.Sleep(p.retryDelay)
			continue
		}

		result, err := p.waitForResponse(ctx, p.fields.correlationID, ch)
		if err != nil {
			return amqp091.Delivery{}, err
		}

		return result, nil
	}

	return amqp091.Delivery{}, fmt.Errorf("failed to publish message after %d retries: %s", p.retryCount, err)
}

func (p *Publisher) PublishWithConfirmation(ctx context.Context, data interface{}) (bool, error) {
	body, err := p.fields.makeData(data)
	if err != nil {
		return false, err
	}

	tried := p.retryCount

	var (
		ch *amqp091.DeferredConfirmation
	)

	for tried > 0 {
		ch, err = p.ch.channel().PublishWithDeferredConfirmWithContext(
			ctx,
			p.exchange,
			p.routingKey,
			p.fields.mandatory,
			p.fields.immediate,
			amqp091.Publishing{
				Headers:       p.fields.headers,
				ContentType:   p.fields.contentType,
				DeliveryMode:  p.fields.deliveryMode,
				CorrelationId: p.fields.correlationID,
				ReplyTo:       p.fields.replyToID,
				Expiration:    p.fields.expiration,
				Body:          body,
			})
		if err != nil {
			tried--
			time.Sleep(p.retryDelay)
			continue
		}

		return ch.WaitContext(ctx)
	}

	return false, fmt.Errorf("failed to publish message after %d retries: %s", p.retryCount, err)
}

func (p *Publisher) waitForResponse(
	ctx context.Context,
	correlationID string,
	ch chan amqp091.Delivery,
) (amqp091.Delivery, error) {

	for {
		select {
		case <-ctx.Done():
			return amqp091.Delivery{}, ctx.Err()
		case result := <-ch:
			if result.ReplyTo != correlationID {
				return amqp091.Delivery{}, PublishResponseInvalidReplyToIdError
			}
			return result, nil
		}
	}
}
