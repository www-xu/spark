package spark

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type AppEnv string

const (
	Prod    AppEnv = "production"
	Dev     AppEnv = "development"
	Staging AppEnv = "staging"
)

var (
	ctx *ApplicationContext
)

func init() {
	ctx = NewApplicationContext()
}

type ApplicationContext struct {
	initLock           *sync.Mutex
	initialized        bool
	config             *ApplicationConfig
	initEventListeners []ApplicationInitEventListener
	stopEventListeners []ApplicationStopEventListener
	shutdownFuncs      []func()
}

func NewApplicationContext() *ApplicationContext {
	return &ApplicationContext{
		initLock:           &sync.Mutex{},
		config:             &ApplicationConfig{},
		initEventListeners: []ApplicationInitEventListener{},
		stopEventListeners: []ApplicationStopEventListener{},
		shutdownFuncs:      []func(){},
	}
}

func RegisterApplicationInitEventListener(listener ApplicationInitEventListener) {
	ctx.initEventListeners = append(ctx.initEventListeners, listener)
}

func RegisterApplicationStopEventListener(listener ApplicationStopEventListener) {
	ctx.stopEventListeners = append(ctx.stopEventListeners, listener)
}

func RegisterShutdownFunc(f func()) {
	ctx.shutdownFuncs = append(ctx.shutdownFuncs, f)
}

func (ctx *ApplicationContext) beforeInit() error {
	for _, listener := range ctx.initEventListeners {
		err := listener.BeforeInit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *ApplicationContext) afterInit() error {
	for _, listener := range ctx.initEventListeners {
		err := listener.AfterInit(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *ApplicationContext) Env() AppEnv {
	normalEnv := viper.GetString("ENV")
	if normalEnv == "" {
		return AppEnv(os.Getenv("ENV"))
	}
	return AppEnv(normalEnv)
}

func Env() AppEnv {
	return ctx.Env()
}

func (ctx *ApplicationContext) Init() error {
	ctx.initLock.Lock()
	defer ctx.initLock.Unlock()

	if ctx.initialized {
		return nil
	}

	err := ctx.beforeInit()
	if err != nil {
		return err
	}

	err = ctx.loadConfig()
	if err != nil {
		return err
	}

	err = ctx.afterInit()
	if err != nil {
		return err
	}

	shutdown, err := InitTracer(ctx.config.serverConfig.Name)
	if err != nil {
		return err
	}
	RegisterShutdownFunc(shutdown)

	ctx.initialized = true

	return nil
}

func Init() error {
	return ctx.Init()
}

// loadConfig from config file, env, flag, ssm, etc.
// priority: ssm > flag > env > config file
func (ctx *ApplicationContext) loadConfig() error {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	ctx.config.fileConfig = viper.New()
	ctx.config.fileConfig.AutomaticEnv()
	ctx.config.fileConfig.AllowEmptyEnv(true)
	ctx.config.fileConfig.SetConfigName(fmt.Sprintf(configFileName, ctx.Env()))
	ctx.config.fileConfig.AddConfigPath(defaultConfigFilePath)

	err := ctx.config.fileConfig.ReadInConfig()
	if err != nil {
		return err
	}

	err = ctx.config.fileConfig.UnmarshalKey("server", &ctx.config.serverConfig)
	if err != nil {
		return err
	}

	err = viper.MergeConfigMap(ctx.config.fileConfig.AllSettings())
	if err != nil {
		return err
	}
	return nil
}

func Close(callback func()) error {
	for _, listener := range ctx.stopEventListeners {
		listener.BeforeStop()
	}

	callback()

	// Call all registered shutdown functions in reverse order.
	for i := len(ctx.shutdownFuncs) - 1; i >= 0; i-- {
		ctx.shutdownFuncs[i]()
	}

	for _, listener := range ctx.stopEventListeners {
		listener.AfterStop()
	}

	return nil
}

func UnmarshalKey[T any](key string) (val *T, err error) {
	if ctx.config == nil {
		return nil, nil
	}

	err = ctx.UnmarshalKey(key, &val)

	return
}

func (ctx *ApplicationContext) UnmarshalKey(key string, rawVal interface{}) (err error) {
	err = viper.UnmarshalKey(key, rawVal)
	if err != nil {
		return err
	}

	return nil
}

func SererName() string {
	return ctx.config.serverConfig.Name
}

func GetConfigString(key string) (string, bool) {
	if ctx.config == nil {
		return "", false
	}

	if viper.IsSet(key) {
		return viper.GetString(key), true
	}

	return "", false
}

func GetConfigUint64(key string) (uint64, bool) {
	if ctx.config == nil {
		return 0, false
	}

	if viper.IsSet(key) {
		return viper.GetUint64(key), true
	}

	return 0, false
}

func GetConfigBool(key string) (bool, bool) {
	if ctx.config == nil {
		return false, false
	}

	if viper.IsSet(key) {
		return viper.GetBool(key), true
	}

	return false, false
}
