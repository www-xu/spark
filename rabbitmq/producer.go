package rabbitmq

import (
	"context"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// Producer is a thread-safe RabbitMQ producer.
// It ensures that publishing to a channel is a mutually exclusive operation.
type Producer struct {
	channel *amqp091.Channel
	mu      sync.Mutex
}

// NewProducer creates a new thread-safe producer.
func NewProducer(channel *amqp091.Channel) *Producer {
	return &Producer{
		channel: channel,
	}
}

// PublishWithContext sends a message to RabbitMQ in a thread-safe manner.
// It injects the OpenTelemetry trace context into the message headers before publishing.
func (p *Producer) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Inject the trace context.
	if msg.Headers == nil {
		msg.Headers = make(amqp091.Table)
	}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, AMQPTextMapCarrier(msg.Headers))

	return p.channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}
