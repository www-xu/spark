package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/www-xu/spark"
	"github.com/www-xu/spark/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Component struct {
	ctx      *spark.ApplicationContext
	config   *Config
	instance *gorm.DB
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
	err := c.ctx.UnmarshalKey("postgres", &c.config)
	if err != nil {
		return err
	}

	if c.config == nil {
		return errors.New("postgres config isn't found")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.config.Host,
		c.config.User,
		c.config.Password,
		c.config.DBName,
		c.config.Port,
		c.config.SSLMode,
	)

	if c.config.Scheme != nil {
		dsn += fmt.Sprintf(" search_path=%s", *c.config.Scheme)
	}

	c.instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         log.NewGormLogger(),
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func Get(ctx context.Context) *gorm.DB {
	return instance.Get(ctx)
}

func (c *Component) Get(ctx context.Context) *gorm.DB {
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
