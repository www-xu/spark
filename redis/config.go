package redis

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	Db       int    `mapstructure:"db"`
}
