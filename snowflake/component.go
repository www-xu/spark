package snowflake

import (
	"context"
	"github.com/sony/sonyflake"
	"github.com/www-xu/spark"
	"math/rand"
)

var instance *Snowflake

func init() {
	instance = NewSnowflake()

	spark.RegisterApplicationInitEventListener(instance)
	spark.RegisterApplicationStopEventListener(instance)
}

type Snowflake struct {
	ctx      *spark.ApplicationContext
	setting  *sonyflake.Settings
	instance *sonyflake.Sonyflake
}

func NewSnowflake() *Snowflake {
	return &Snowflake{}
}

func (c *Snowflake) Instantiate() (err error) {
	c.setting = &sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return uint16(rand.Uint32()), nil
		},
	}

	c.instance, err = sonyflake.New(*c.setting)
	if instance == nil || err != nil {
		return err
	}

	return nil
}

func Get(ctx context.Context) *sonyflake.Sonyflake {
	return instance.Get(ctx)
}

func (c *Snowflake) Get(ctx context.Context) *sonyflake.Sonyflake {
	return c.instance
}

func (c *Snowflake) Close() error {
	return nil
}

func (c *Snowflake) BeforeInit() error {
	return nil
}

func (c *Snowflake) AfterInit(applicationContext *spark.ApplicationContext) error {
	c.ctx = applicationContext

	return c.Instantiate()
}

func (c *Snowflake) BeforeStop() {
	return
}

func (c *Snowflake) AfterStop() {
	_ = c.Close()

	return
}
