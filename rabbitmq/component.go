package rabbitmq

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/www-xu/spark"

	"github.com/ThreeDotsLabs/watermill"
	watermillAmqp "github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Component struct {
	ctx        *spark.ApplicationContext
	config     *Config
	amqpConfig watermillAmqp.Config
	publisher  *watermillAmqp.Publisher
	subscriber *watermillAmqp.Subscriber
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

func (c *Component) Instantiate() error {
	err := c.ctx.UnmarshalKey("rabbitmq", &c.config)
	if err != nil {
		return err
	}
	if c.config == nil {
		return errors.New("rabbitmq config isn't found")
	}

	c.amqpConfig = watermillAmqp.NewDurableQueueConfig(c.config.Uri)
	c.amqpConfig.Marshaler = watermillAmqp.DefaultMarshaler{
		PostprocessPublishing: func(publishing amqp.Publishing) amqp.Publishing {
			var messageID string = uuid.New().String()
			if value, ok := publishing.Headers[watermillAmqp.DefaultMessageUUIDHeaderKey]; ok {
				if uuid, ok := value.(string); ok {
					messageID = uuid
				}
			}
			publishing.MessageId = messageID
			return publishing
		},
	}
	c.amqpConfig.Publish.GenerateRoutingKey = func(topic string) string {
		if routingKey, ok := c.config.Topics[topic]; ok {
			return routingKey
		}
		return topic
	}
	c.amqpConfig.Exchange.GenerateName = func(topic string) string {
		return c.config.Exchange
	}
	c.amqpConfig.Exchange.Type = "topic"
	c.amqpConfig.Exchange.Durable = true
	c.amqpConfig.Exchange.AutoDeleted = true

	c.publisher, err = watermillAmqp.NewPublisher(
		c.amqpConfig,
		watermill.NewStdLogger(spark.Env() != spark.Prod, spark.Env() != spark.Prod),
	)
	if err != nil {
		panic(err)
	}

	c.subscriber, err = watermillAmqp.NewSubscriber(
		c.amqpConfig,
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	return nil
}

func Get(ctx context.Context) *watermillAmqp.Publisher {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *watermillAmqp.Publisher {
	return c.publisher
}

func (c *Component) Close() error {
	_ = c.subscriber.Close()
	_ = c.publisher.Close()

	return nil
}

func (c *Component) BeforeInit() error { return nil }

func (c *Component) AfterInit(applicationContext *spark.ApplicationContext) error {
	c.ctx = applicationContext

	return c.Instantiate()
}

func GetPublisher() *watermillAmqp.Publisher {
	return instance.publisher
}

func (c *Component) GetPublisher() *watermillAmqp.Publisher {
	return c.publisher
}

func GetSubscriber() *watermillAmqp.Subscriber {
	return instance.subscriber
}

func (c *Component) GetSubscriber() *watermillAmqp.Subscriber {
	return c.subscriber
}

func (c *Component) BeforeStop() {}
func (c *Component) AfterStop()  {}
