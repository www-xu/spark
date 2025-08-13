package rabbitmq

type Config struct {
	Uri    string            `mapstructure:"uri"`
	Topics map[string]string `mapstructure:"topics"`
}
