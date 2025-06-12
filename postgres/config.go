package postgres

type Config struct {
	User         string  `mapstructure:"user"`
	Password     string  `mapstructure:"password"`
	Host         string  `mapstructure:"host"`
	Port         string  `mapstructure:"port"`
	DBName       string  `mapstructure:"db_name"`
	SSLMode      string  `mapstructure:"ssl_mode"`
	MaxOpenConns int     `mapstructure:"max_open_conns"`
	MaxIdleConns int     `mapstructure:"max_idle_conns"`
	MaxLifetime  int     `mapstructure:"max_life_time"`
	scheme       *string `mapstructure:"scheme"`
}
