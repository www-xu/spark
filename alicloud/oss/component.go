package oss

import (
	"context"
	"errors"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/www-xu/spark"
)

type Component struct {
	ctx      *spark.ApplicationContext
	config   *Config
	instance *oss.Client
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
	err = c.ctx.UnmarshalKey("alicloud_oss", &c.config)
	if err != nil {
		return err
	}

	if c.config == nil {
		return errors.New("alicloud_oss config isn't found")
	}

	c.instance, err = oss.New(c.config.Endpoint, c.config.AccessKeyID, c.config.AccessKeySecret)
	if err != nil {
		return err
	}

	return nil
}

func Get(ctx context.Context) *oss.Client {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *oss.Client {
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
