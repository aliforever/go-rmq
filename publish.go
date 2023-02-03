package rmq

import (
	"context"
	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type Publish struct {
	ch         *amqp091.Channel
	exchange   string
	routingKey string

	fields *PublishFields
}

func NewPublish(ch *amqp091.Channel, exchange string, routingKey string, fields *PublishFields) *Publish {
	return &Publish{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,

		fields: fields,
	}
}

func (p *Publish) WithFields(fields *PublishFields) *Publish {
	p.fields = fields

	return p
}

func (p *Publish) Fields() *PublishFields {
	return p.fields
}

func (p *Publish) Publish(ctx context.Context, data interface{}) error {
	body, err := p.fields.makeData(data)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, p.fields.mandatory, p.fields.immediate, amqp091.Publishing{
		Headers:       p.fields.headers,
		ContentType:   p.fields.contentType,
		DeliveryMode:  p.fields.deliveryMode,
		CorrelationId: p.fields.correlationID,
		ReplyTo:       p.fields.replyToID,
		Expiration:    p.fields.expiration,
		Body:          body,
	})
}

// PublishAwaitResponse creates a channel and stores it in responseMap
//	If an outsider writes the response to that map it checks for the reply to id to see if it matches the correlation id
//	Fails upon an invalid correlation id or in case of a timeout
func (p *Publish) PublishAwaitResponse(
	ctx context.Context, data interface{}, responseMap *genericSync.Map[DeliveryChannel]) (amqp091.Delivery, error) {

	if p.fields.responseTimeout == 0 {
		return amqp091.Delivery{}, ResponseTimeoutNotSetError
	}

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

	err = p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, p.fields.mandatory, p.fields.immediate, amqp091.Publishing{
		Headers:       p.fields.headers,
		ContentType:   p.fields.contentType,
		DeliveryMode:  p.fields.deliveryMode,
		CorrelationId: p.fields.correlationID,
		ReplyTo:       p.fields.replyToID,
		Expiration:    p.fields.expiration,
		Body:          body,
	})
	if err != nil {
		return amqp091.Delivery{}, err
	}

	result, err := p.waitForResponse(ctx, p.fields.correlationID, ch)
	if err != nil {
		return amqp091.Delivery{}, err
	}

	return result, nil
}

func (p *Publish) PublishWithConfirmation(ctx context.Context, data interface{}) (bool, error) {
	body, err := p.fields.makeData(data)
	if err != nil {
		return false, err
	}

	ch, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, p.exchange, p.routingKey, p.fields.mandatory, p.fields.immediate, amqp091.Publishing{
		Headers:       p.fields.headers,
		ContentType:   p.fields.contentType,
		DeliveryMode:  p.fields.deliveryMode,
		CorrelationId: p.fields.correlationID,
		ReplyTo:       p.fields.replyToID,
		Expiration:    p.fields.expiration,
		Body:          body,
	})
	if err != nil {
		return false, err
	}

	return ch.WaitContext(ctx)
}

func (p *Publish) waitForResponse(ctx context.Context, correlationID string, ch chan amqp091.Delivery) (amqp091.Delivery, error) {
	ticker := time.NewTicker(p.fields.responseTimeout)

	for {
		select {
		case <-ctx.Done():
			return amqp091.Delivery{}, ctx.Err()
		case <-ticker.C:
			return amqp091.Delivery{}, PublishResponseTimeoutError
		case result := <-ch:
			if result.ReplyTo != correlationID {
				return amqp091.Delivery{}, PublishResponseInvalidReplyToIdError
			}
			return result, nil
		}
	}
}
