package spark

type ApplicationInitEventListener interface {
	BeforeInit() error
	AfterInit(applicationContext *ApplicationContext) error
}

type ApplicationStopEventListener interface {
	BeforeStop()
	AfterStop()
}
