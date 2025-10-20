package rabbitmq

type ExchangeConfig struct {
	Type        string         `mapstructure:"type"`
	Durable     bool           `mapstructure:"durable"`
	AutoDeleted bool           `mapstructure:"auto_deleted"`
	Args        map[string]any `mapstructure:"args"` // exchange type 需要的参数
}

type Config struct {
	Uri         string                    `mapstructure:"uri"`
	Exchanges   map[string]ExchangeConfig `mapstructure:"exchanges"`    // key 是 exchange name
	RoutingKeys map[string]string         `mapstructure:"routing_keys"` // topic -> routing key
}
