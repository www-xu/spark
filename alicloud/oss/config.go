package oss

type Config struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
	Region   string `json:"region" mapstructure:"region"`
}
