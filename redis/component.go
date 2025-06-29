package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/www-xu/spark"
)

type Component struct {
	ctx      *spark.ApplicationContext
	config   *RedisConfig
	instance *redis.Client
	locker   *redsync.Redsync
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
	err := c.ctx.UnmarshalKey("redis", &c.config)
	if err != nil {
		return err
	}

	if c.config == nil {
		return errors.New("redis config isn't found")
	}

	c.instance = redis.NewClient(&redis.Options{
		Addr:     c.config.Address,
		Password: c.config.Password,
		DB:       c.config.Db,
	})
	c.instance.AddHook(redisotel.NewTracingHook())

	pool := goredis.NewPool(c.instance)
	c.locker = redsync.New(pool)

	return nil
}

func Get(ctx context.Context) *redis.Client {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *redis.Client {
	return c.instance
}

func Locker(ctx context.Context) *redsync.Redsync {
	return instance.Locker(ctx)
}

func (c *Component) Locker(ctx context.Context) *redsync.Redsync {
	return c.locker
}

func (c *Component) Close() error {
	return c.instance.Close()
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
