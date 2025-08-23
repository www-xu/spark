package oss

type Config struct {
	Endpoint        string `json:"endpoint" mapstructure:"endpoint"`
	AccessKeyID     string `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret"`
}
