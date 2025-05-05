package spark

import "github.com/spf13/viper"

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ApplicationConfig struct {
	fileConfig   *viper.Viper //
	serverConfig *ServerConfig
}

const (
	configFileName              = "cfg.%s" // config.{env}
	defaultConfigFilePath       = "config/"
	customizeConfigFilePathFlag = "config_path"
)
