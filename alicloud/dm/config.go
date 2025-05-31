package dm

type Config struct {
	Endpoint *string `mapstructure:"endpoint"`
	Region   *string `mapstructure:"region"`
}
