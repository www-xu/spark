package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/www-xu/spark"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	config *Config
	engine *gin.Engine
}

func NewServer(middleware ...gin.HandlerFunc) *Server {
	ginEngine := gin.New()
	ginEngine.ContextWithFallback = true
	ginEngine.Use(middleware...)

	return &Server{
		engine: ginEngine,
	}
}

func (s *Server) Run() (err error) {
	err = spark.Init()
	if err != nil {
		return err
	}

	s.config, err = spark.UnmarshalKey[Config]("server")
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    s.config.Address,
		Handler: s.engine,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	_ = spark.Close(func() {
		_ = server.Close()
	})

	return nil
}
