package rabbitmq

import (
	"context"
	"errors"

	"github.com/www-xu/spark"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
)

type Component struct {
	ctx        *spark.ApplicationContext
	config     *Config
	amqpConfig amqp.Config
	publisher  *amqp.Publisher
	subscriber *amqp.Subscriber
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

	c.amqpConfig = amqp.NewDurableQueueConfig(c.config.Uri)

	c.publisher, err = amqp.NewPublisher(
		c.amqpConfig,
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	c.subscriber, err = amqp.NewSubscriber(
		c.amqpConfig,
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	return nil
}

func Get(ctx context.Context) *amqp.Publisher {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *amqp.Publisher {
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

func GetPublisher() *amqp.Publisher {
	return instance.publisher
}

func (c *Component) GetPublisher() *amqp.Publisher {
	return c.publisher
}

func GetSubscriber() *amqp.Subscriber {
	return instance.subscriber
}

func (c *Component) GetSubscriber() *amqp.Subscriber {
	return c.subscriber
}

func (c *Component) BeforeStop() {}
func (c *Component) AfterStop()  {}
