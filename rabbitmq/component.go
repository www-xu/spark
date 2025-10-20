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
	ctx         *spark.ApplicationContext
	config      *Config
	publishers  map[string]*watermillAmqp.Publisher  // exchange name -> publisher
	subscribers map[string]*watermillAmqp.Subscriber // exchange name -> subscriber
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

	// 初始化 publishers 和 subscribers maps
	c.publishers = make(map[string]*watermillAmqp.Publisher)
	c.subscribers = make(map[string]*watermillAmqp.Subscriber)

	// 为每个 exchange 创建独立的 publisher 和 subscriber
	// watermillAmqp 会在创建时自动声明 exchange
	for exchangeName, exchangeConfig := range c.config.Exchanges {
		// 创建针对该 exchange 的配置
		amqpConfig := watermillAmqp.NewDurableQueueConfig(c.config.Uri)
		amqpConfig.Marshaler = watermillAmqp.DefaultMarshaler{
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

		// 为这个 exchange 配置 routing key 生成器
		amqpConfig.Publish.GenerateRoutingKey = func(topic string) string {
			if routingKey, ok := c.config.RoutingKeys[topic]; ok {
				return routingKey
			}
			// 如果没有配置 routing key，默认使用 topic 名称
			return topic
		}

		// 固定返回当前 exchange 的名称
		currentExchangeName := exchangeName
		amqpConfig.Exchange.GenerateName = func(topic string) string {
			return currentExchangeName
		}

		// 设置 exchange 配置
		exchangeType := exchangeConfig.Type
		amqpConfig.Exchange.Type = exchangeType
		amqpConfig.Exchange.Durable = exchangeConfig.Durable
		amqpConfig.Exchange.AutoDeleted = exchangeConfig.AutoDeleted
		amqpConfig.Exchange.Arguments = exchangeConfig.Args

		// 创建 publisher，watermillAmqp 会自动声明 exchange
		publisher, err := watermillAmqp.NewPublisher(
			amqpConfig,
			watermill.NewStdLogger(spark.Env() != spark.Prod, spark.Env() != spark.Prod),
		)
		if err != nil {
			return err
		}
		c.publishers[exchangeName] = publisher

		// 创建 subscriber，watermillAmqp 会自动声明 exchange
		subscriber, err := watermillAmqp.NewSubscriber(
			amqpConfig,
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			return err
		}
		c.subscribers[exchangeName] = subscriber
	}

	return nil
}

func Get(ctx context.Context, exchangeName string) (*watermillAmqp.Publisher, error) {
	return instance.Get(ctx, exchangeName)
}

func (c *Component) Get(ctx context.Context, exchangeName string) (*watermillAmqp.Publisher, error) {
	return c.GetPublisher(exchangeName)
}

func (c *Component) Close() error {
	for _, subscriber := range c.subscribers {
		_ = subscriber.Close()
	}
	for _, publisher := range c.publishers {
		_ = publisher.Close()
	}

	return nil
}

func (c *Component) BeforeInit() error { return nil }

func (c *Component) AfterInit(applicationContext *spark.ApplicationContext) error {
	c.ctx = applicationContext

	return c.Instantiate()
}

func GetPublisher(exchangeName string) (*watermillAmqp.Publisher, error) {
	return instance.GetPublisher(exchangeName)
}

func (c *Component) GetPublisher(exchangeName string) (*watermillAmqp.Publisher, error) {
	publisher, ok := c.publishers[exchangeName]
	if !ok {
		return nil, errors.New("publisher not found for exchange: " + exchangeName)
	}

	return publisher, nil
}

func GetSubscriber(exchangeName string) (*watermillAmqp.Subscriber, error) {
	return instance.GetSubscriber(exchangeName)
}

func (c *Component) GetSubscriber(exchangeName string) (*watermillAmqp.Subscriber, error) {
	subscriber, ok := c.subscribers[exchangeName]
	if !ok {
		return nil, errors.New("subscriber not found for exchange: " + exchangeName)
	}

	return subscriber, nil
}

func (c *Component) BeforeStop() {}
func (c *Component) AfterStop()  {}
