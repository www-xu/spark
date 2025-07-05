package gin

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/www-xu/spark"
	"github.com/www-xu/spark/gin/middleware"
)

type Server struct {
	config *Config
	*gin.Engine
}

func NewServer(middlewares ...gin.HandlerFunc) *Server {
	ginEngine := gin.New()
	ginEngine.ContextWithFallback = true

	// Add custom logger and recovery middleware
	ginEngine.Use(middleware.Observability(), gin.Recovery())
	ginEngine.Use(middlewares...)

	return &Server{
		Engine: ginEngine,
	}
}

func (s *Server) Init() (err error) {
	err = spark.Init()
	if err != nil {
		return err
	}

	return nil
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
		Handler: s.Engine,
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
