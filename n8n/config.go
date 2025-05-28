package n8n

type Config struct {
	Host            string `mapstructure:"host"`
	AuthHeaderKey   string `mapstructure:"auth_header_key"`
	AuthHeaderValue string `mapstructure:"auth_header_value"`
}
