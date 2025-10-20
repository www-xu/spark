package rabbitmq

type Config struct {
	Uri             string            `mapstructure:"uri"`
	DefaultExchange string            `mapstructure:"default_exchange"`
	Topics          map[string]string `mapstructure:"topics"`
	Exchanges       map[string]string `mapstructure:"exchanges"`
}
