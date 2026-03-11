# Spark

Spark is a lightweight, modular, and cloud-native Go application framework designed to scaffold enterprise-grade services quickly. It provides a robust core for configuration management, application lifecycle handling, dependency injection, and out-of-the-box OpenTelemetry tracing.

## Core Features
*   **Application Context Management (`context.go`)**: Manages the application lifecycle, including graceful initialization (`BeforeInit`, `AfterInit`), shutdown (`BeforeStop`, `AfterStop`), and component registration.
*   **Environment & Configuration (`config.go`)**: Built on top of [Viper](https://github.com/spf13/viper), supporting robust configuration management. Loads configs smoothly from `config/cfg.{env}.yaml`, environment variables, or flags.
*   **OpenTelemetry Built-in (`otel.go`)**: Native integration with the OpenTelemetry SDK. Components are automatically instrumented with W3C Trace Context propagation.
*   **Component Architecture (`component.go`)**: Standardized interfaces `IComponent`, `ISingleSourceComponent`, and `IMultiSourceComponent` for creating plug-and-play modules.

## Supported Integrations (Components)
Spark includes several pre-built, ready-to-use components. These integrate seamlessly into the Spark App lifecycle and include observability hooks by default:
*   `gin`: A Web Server wrapper extending `gin-gonic/gin` with observability middleware and graceful shutdown.
*   `redis`: Redis client integration via `go-redis/redis/v8`, including `redsync` for distributed locks and `redisotel` for tracing.
*   `mysql` & `postgres`: Relational database connections.
*   `rabbitmq`: Message broker integration.
*   `snowflake`: Distributed unique ID generation.
*   `log`: Standardized logging wrapper.
*   `n8n`, `dify`, `alicloud`: Integrations for popular third-party cloud services and automation platforms.

## Installation

```bash
go get github.com/www-xu/spark
```

## Quick Start

### 1. Define Initialization

Create a configuration file `config/cfg.development.yaml` (example):
```yaml
server:
  name: "my-app"
  address: ":8080"
redis:
  address: "localhost:6379"
```

### 2. Basic Gin Server Setup

Here is a minimal example using Spark with the wrapped Gin component:

```go
package main

import (
	"log"

	sgin "github.com/www-xu/spark/gin"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Spark Gin server (comes with Observability and Recovery middleware)
	srv := sgin.NewServer()

	// Register your routes
	srv.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Run initializes the spark context, configuration, tracers, and starts the server
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
```

### 3. Component Usage Example (Redis)

Components seamlessly hook into the Spark initialization phase. To use Redis:

```go
package main

import (
	"context"
	"fmt"

	"github.com/www-xu/spark"
	sredis "github.com/www-xu/spark/redis"
)

func main() {
    // Initialize Spark AppContext manually if not using a wrapped server like sgin.Server
    if err := spark.Init(); err != nil {
        panic(err)
    }

    // Defer graceful shutdown 
    defer spark.Close(func() {
        fmt.Println("Shutting down the application...")
    })

    // Get the global Redis client instance
    client := sredis.Get(context.Background())
    
    // Use it
    err := client.Set(context.Background(), "key", "value", 0).Err()
    if err != nil {
        panic(err)
    }
}
```

## Creating Custom Components
You can easily build custom modules that tie into the Spark lifecycle by implementing the `ApplicationInitEventListener` or `ApplicationStopEventListener` interfaces:

```go
type MyCustomPlugin struct{}

func (p *MyCustomPlugin) BeforeInit() error { return nil }
func (p *MyCustomPlugin) AfterInit(ctx *spark.ApplicationContext) error { 
    // Perform initialization logic here
    return nil 
}
func (p *MyCustomPlugin) BeforeStop() {}
func (p *MyCustomPlugin) AfterStop() {}

func init() {
    plugin := &MyCustomPlugin{}
    spark.RegisterApplicationInitEventListener(plugin)
    spark.RegisterApplicationStopEventListener(plugin)
}
```

## License
MIT
