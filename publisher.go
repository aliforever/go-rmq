package rmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/rabbitmq/amqp091-go"
)

type PublisherImpl interface {
	WithFields(fields *PublishFields) PublisherImpl
	Fields() *PublishFields
	Publish(ctx context.Context, data interface{}) error
	PublishAwaitResponse(
		ctx context.Context,
		data interface{},
		responseMap *genericSync.Map[chan amqp091.Delivery],
	) (amqp091.Delivery, error)
	PublishWithConfirmation(ctx context.Context, data interface{}) (bool, error)
}

type Publisher struct {
	ch         *Channel
	exchange   string
	routingKey string

	retryCount int
	retryDelay time.Duration

	fields *PublishFields
	mu     sync.RWMutex
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
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fields = fields
	return p
}

func (p *Publisher) Fields() *PublishFields {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.fields
}

func (p *Publisher) Publish(ctx context.Context, data interface{}) error {
	p.mu.RLock()
	fields := p.fields
	p.mu.RUnlock()

	body, err := fields.makeData(data)
	if err != nil {
		return err
	}

	backoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second

	for attempt := 0; attempt < p.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if channel is healthy
		if !p.ch.IsHealthy() {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		ch := p.ch.channel()
		if ch == nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		err = ch.PublishWithContext(ctx,
			p.exchange,
			p.routingKey,
			fields.mandatory,
			fields.immediate,
			amqp091.Publishing{
				Headers:       fields.headers,
				ContentType:   fields.contentType,
				DeliveryMode:  fields.deliveryMode,
				CorrelationId: fields.correlationID,
				ReplyTo:       fields.replyToID,
				Expiration:    fields.expiration,
				Body:          body,
			})
		if err != nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to publish message after %d retries: %w", p.retryCount, err)
}

func (p *Publisher) PublishAwaitResponse(
	ctx context.Context,
	data interface{},
	responseMap *genericSync.Map[chan amqp091.Delivery],
) (amqp091.Delivery, error) {
	if responseMap == nil {
		return amqp091.Delivery{}, ResponseMapNotSetError
	}

	p.mu.RLock()
	fields := p.fields
	p.mu.RUnlock()

	if fields.correlationID == "" {
		return amqp091.Delivery{}, CorrelationIdNotSetError
	}

	respCh := make(chan amqp091.Delivery, 1)
	responseMap.Store(fields.correlationID, respCh)
	defer responseMap.Delete(fields.correlationID)

	body, err := fields.makeData(data)
	if err != nil {
		return amqp091.Delivery{}, err
	}

	backoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second

	for attempt := 0; attempt < p.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return amqp091.Delivery{}, ctx.Err()
		default:
		}

		// Check if channel is healthy
		if !p.ch.IsHealthy() {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		ch := p.ch.channel()
		if ch == nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		err = ch.PublishWithContext(ctx,
			p.exchange,
			p.routingKey,
			fields.mandatory,
			fields.immediate,
			amqp091.Publishing{
				Headers:       fields.headers,
				ContentType:   fields.contentType,
				DeliveryMode:  fields.deliveryMode,
				CorrelationId: fields.correlationID,
				ReplyTo:       fields.replyToID,
				Expiration:    fields.expiration,
				Body:          body,
			})
		if err != nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		// Wait for response
		result, err := p.waitForResponse(ctx, fields.correlationID, respCh)
		if err != nil {
			return amqp091.Delivery{}, err
		}

		return result, nil
	}

	return amqp091.Delivery{}, fmt.Errorf("failed to publish message after %d retries: %w", p.retryCount, err)
}

func (p *Publisher) PublishWithConfirmation(ctx context.Context, data interface{}) (bool, error) {
	p.mu.RLock()
	fields := p.fields
	p.mu.RUnlock()

	body, err := fields.makeData(data)
	if err != nil {
		return false, err
	}

	backoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second

	for attempt := 0; attempt < p.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// Check if channel is healthy
		if !p.ch.IsHealthy() {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		ch := p.ch.channel()
		if ch == nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx,
			p.exchange,
			p.routingKey,
			fields.mandatory,
			fields.immediate,
			amqp091.Publishing{
				Headers:       fields.headers,
				ContentType:   fields.contentType,
				DeliveryMode:  fields.deliveryMode,
				CorrelationId: fields.correlationID,
				ReplyTo:       fields.replyToID,
				Expiration:    fields.expiration,
				Body:          body,
			})
		if err != nil {
			if attempt < p.retryCount-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		return confirm.WaitContext(ctx)
	}

	return false, fmt.Errorf("failed to publish message after %d retries: %w", p.retryCount, err)
}

func (p *Publisher) waitForResponse(
	ctx context.Context,
	correlationID string,
	respCh chan amqp091.Delivery,
) (amqp091.Delivery, error) {
	select {
	case <-ctx.Done():
		return amqp091.Delivery{}, ctx.Err()
	case result := <-respCh:
		if result.CorrelationId != correlationID {
			return amqp091.Delivery{}, fmt.Errorf("correlation ID mismatch: expected %s, got %s", correlationID, result.CorrelationId)
		}
		return result, nil
	}
}
