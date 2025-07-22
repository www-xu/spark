package sms

import (
	"context"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/www-xu/spark"
)

type Component struct {
	ctx      *spark.ApplicationContext
	config   *Config
	instance *dysmsapi20170525.Client
}

func NewComponent() *Component {
	return &Component{}
}

var instance *Component

func init() {
	instance = NewComponent()

	spark.RegisterApplicationInitEventListener(instance)
	spark.RegisterApplicationStopEventListener(instance)
}

func (c *Component) Instantiate() (err error) {
	// err = c.ctx.UnmarshalKey("alicloud_sms", &c.config)
	// if err != nil {
	// 	return err
	// }

	// if c.config == nil {
	// 	return errors.New("alicloud_sms config isn't found")
	// }

	keyId, _ := spark.GetConfigString("ALIBABA_CLOUD_ACCESS_KEY_ID")
	keySecret, _ := spark.GetConfigString("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

	c.instance, err = dysmsapi20170525.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(keyId),
		AccessKeySecret: tea.String(keySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	})
	if err != nil {
		return err
	}

	return nil
}

func Get(ctx context.Context) *dysmsapi20170525.Client {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *dysmsapi20170525.Client {
	return c.instance
}

func (c *Component) Close() error {

	return nil
}

func (c *Component) BeforeInit() error {
	return nil
}

func (c *Component) AfterInit(applicationContext *spark.ApplicationContext) error {
	c.ctx = applicationContext

	return c.Instantiate()
}

func (c *Component) BeforeStop() {
	return
}

func (c *Component) AfterStop() {
	_ = c.Close()

	return
}
