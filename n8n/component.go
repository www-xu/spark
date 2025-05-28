package n8n

import (
	"context"
	"errors"
	"github.com/www-xu/spark"
)

type Component struct {
	ctx      *spark.ApplicationContext
	config   *Config
	instance *Client
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

func (c *Component) Instantiate() error {
	err := c.ctx.UnmarshalKey("mysql", &c.config)
	if err != nil {
		return err
	}

	if c.config == nil {
		return errors.New("mysql config isn't found")
	}

	c.instance = NewClient(c.config.Host, c.config.AuthHeaderKey, c.config.AuthHeaderValue)

	return nil
}

func Get(ctx context.Context) *Client {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *Client {
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
