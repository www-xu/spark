package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/www-xu/spark"
	"github.com/www-xu/spark/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AMQPTextMapCarrier implements the TextMapCarrier interface for AMQP headers.
type AMQPTextMapCarrier amqp091.Table

// Get returns the value associated with the given key.
func (c AMQPTextMapCarrier) Get(key string) string {
	if val, ok := c[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Set sets the value for the given key.
func (c AMQPTextMapCarrier) Set(key string, value string) {
	c[key] = value
}

// Keys returns a slice of all keys in the carrier.
func (c AMQPTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

var tracer = otel.Tracer("rabbitmq-component")

// consumerConfig holds the settings for a consumer to allow for re-registration.
type consumerConfig struct {
	queue        string
	autoAck      bool
	numConsumers int
	handler      func(ctx context.Context, delivery amqp091.Delivery)
}

type Component struct {
	appCtx *spark.ApplicationContext
	config *Config

	// For managing state and reconnection
	connCtx    context.Context
	connCancel context.CancelFunc
	wg         sync.WaitGroup

	mu              sync.Mutex
	connection      *amqp091.Connection
	producer        *Producer
	consumerConfigs []*consumerConfig
}

func NewComponent() *Component {
	return &Component{}
}

var instance *Component

func init() {
	instance = NewComponent()
	spark.RegisterApplicationInitEventListener(instance)
	spark.RegisterApplicationStopEventListener(instance)
}

// AfterInit starts the connection manager goroutine.
func (c *Component) AfterInit(applicationContext *spark.ApplicationContext) error {
	c.appCtx = applicationContext
	err := c.appCtx.UnmarshalKey("rabbitmq", &c.config)
	if err != nil {
		return err
	}
	if c.config == nil {
		return errors.New("rabbitmq config isn't found")
	}

	c.connCtx, c.connCancel = context.WithCancel(context.Background())
	c.wg.Add(1)
	go c.connectionManager()
	return nil
}

// connectionManager is a long-running goroutine that maintains the connection.
func (c *Component) connectionManager() {
	defer c.wg.Done()
	log.WithContext(c.connCtx).Info("starting rabbitmq connection manager")

	ticker := time.NewTicker(5 * time.Second) // Reconnect interval
	defer ticker.Stop()

	for {
		if c.connCtx.Err() != nil {
			log.WithContext(c.connCtx).Info("stopping rabbitmq connection manager")
			return
		}

		conn, err := amqp091.Dial(c.config.Uri)
		if err != nil {
			log.WithContext(c.connCtx).Errorf("failed to connect to rabbitmq: %v", err)
			<-ticker.C // Wait before retrying
			continue
		}

		if c.handleConnection(conn) {
			return // Graceful shutdown requested
		}
		// Connection lost, loop to reconnect
	}
}

// handleConnection manages an active connection and its resources.
// Returns true if shutdown was initiated, false if connection was lost.
func (c *Component) handleConnection(conn *amqp091.Connection) (gracefulShutdown bool) {
	c.setupResources(conn)
	log.WithContext(c.connCtx).Info("rabbitmq connection established and resources configured")

	closeCh := conn.NotifyClose(make(chan *amqp091.Error))

	select {
	case <-c.connCtx.Done():
		log.WithContext(c.connCtx).Info("shutting down rabbitmq connection gracefully")
		c.cleanUpConnections()
		return true
	case err := <-closeCh:
		log.WithContext(c.connCtx).Warnf("rabbitmq connection lost: %v. Reconnecting...", err)
		c.cleanUpConnections()
		return false
	}
}

// setupResources creates a new producer and launches all registered consumers.
func (c *Component) setupResources(conn *amqp091.Connection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connection = conn

	producerChannel, err := c.connection.Channel()
	if err != nil {
		log.WithContext(c.connCtx).Errorf("failed to create producer channel, will force reconnect: %v", err)
		_ = c.connection.Close()
		return
	}
	c.producer = NewProducer(producerChannel)

	for _, cfg := range c.consumerConfigs {
		c.launchConsumer(cfg)
	}
}

// launchConsumer starts the goroutines for a consumer config. Must be called with mutex held.
func (c *Component) launchConsumer(cfg *consumerConfig) {
	for i := 0; i < cfg.numConsumers; i++ {
		ch, err := c.connection.Channel()
		if err != nil {
			log.WithContext(c.connCtx).Errorf("failed to create channel for consumer on queue %s: %v", cfg.queue, err)
			continue
		}
		c.runConsumerGoroutine(ch, cfg)
	}
}

func (c *Component) runConsumerGoroutine(ch *amqp091.Channel, cfg *consumerConfig) {
	if !cfg.autoAck {
		if err := ch.Qos(1, 0, false); err != nil {
			_ = ch.Close()
			log.WithContext(c.connCtx).Errorf("failed to set QoS for queue %s: %v", cfg.queue, err)
			return
		}
	}
	msgs, err := ch.Consume(cfg.queue, "", cfg.autoAck, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		log.WithContext(c.connCtx).Errorf("failed to start consuming from queue %s: %v", cfg.queue, err)
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			// Channel is closed when connection is lost, no need to close here.
			if r := recover(); r != nil {
				log.Errorf("rabbitmq consumer loop for queue [%s] panicked: %v", cfg.queue, r)
			}
		}()
		log.Infof("consumer started for queue [%s]", cfg.queue)
		for d := range msgs {
			c.processMessage(d, cfg)
		}
		log.Infof("consumer stopped for queue [%s]", cfg.queue)
	}()
}

func (c *Component) processMessage(d amqp091.Delivery, cfg *consumerConfig) {
	var ctx context.Context
	defer func() {
		if r := recover(); r != nil {
			log.WithContext(ctx).Errorf("handler panic for queue: %v", r)
		}
	}()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(context.Background(), AMQPTextMapCarrier(d.Headers))
	ctx, span := tracer.Start(ctx, "rabbitmq.consume")
	defer span.End()

	spanCtx := trace.SpanContextFromContext(ctx)
	ctx = context.WithValue(ctx, log.TraceIdKey, spanCtx.TraceID().String())
	ctx = context.WithValue(ctx, log.SpanIdKey, spanCtx.SpanID().String())
	ctx = context.WithValue(ctx, log.RequestIdKey, uuid.New().String())

	cfg.handler(ctx, d)

	if !cfg.autoAck {
		if err := d.Ack(false); err != nil {
			log.WithContext(ctx).Errorf("failed to ack rabbitmq message: %v", err)
		}
	}
}

// cleanUpConnections closes the producer channel and connection.
func (c *Component) cleanUpConnections() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.producer != nil && c.producer.channel != nil {
		_ = c.producer.channel.Close()
	}
	c.producer = nil
	if c.connection != nil && !c.connection.IsClosed() {
		_ = c.connection.Close()
	}
}

// GetProducer returns the current producer.
func (c *Component) GetProducer() *Producer {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.producer
}

// RegisterConsumer stores a consumer's configuration to be launched by the connection manager.
func (c *Component) RegisterConsumer(queue string, autoAck bool, numConsumers int, handler func(ctx context.Context, delivery amqp091.Delivery)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newCfg := &consumerConfig{
		queue:        queue,
		autoAck:      autoAck,
		numConsumers: numConsumers,
		handler:      handler,
	}
	c.consumerConfigs = append(c.consumerConfigs, newCfg)

	if c.connection != nil && !c.connection.IsClosed() {
		c.launchConsumer(newCfg)
	}
	return nil
}

func RegisterConsumer(queue string, autoAck bool, numConsumers int, handler func(ctx context.Context, delivery amqp091.Delivery)) error {
	return instance.RegisterConsumer(queue, autoAck, numConsumers, handler)
}

// Close signals the connection manager to shut down and waits for it to finish.
func (c *Component) Close() error {
	log.WithContext(context.Background()).Info("closing rabbitmq component")
	if c.connCancel != nil {
		c.connCancel()
	}
	c.wg.Wait()
	return nil
}

// Deprecated: These methods are no longer relevant in the new design.
func (c *Component) BeforeInit() error                 { return nil }
func (c *Component) BeforeStop()                       {}
func (c *Component) AfterStop()                        { _ = c.Close() }
func (c *Component) Instantiate() error                { return nil }
func Get(ctx context.Context) *Producer                { return instance.GetProducer() }
func (c *Component) Get(ctx context.Context) *Producer { return c.GetProducer() }
