package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/www-xu/spark"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
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
	err := c.ctx.UnmarshalKey("mysql", &c.config)
	if err != nil {
		return err
	}

	if c.config == nil {
		return errors.New("mysql config isn't found")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		c.config.User,
		c.config.Password,
		c.config.Host,
		c.config.Port,
		c.config.DBName,
	)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(c.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(c.config.MaxLifetime))
	c.instance, err = gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}))
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
