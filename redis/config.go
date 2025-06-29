package redis

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Db       int    `mapstructure:"db"`
}
