package rabbitmq

type Config struct {
	Uri      string `mapstructure:"uri"`
	Exchange string `mapstructure:"exchange"`
}
