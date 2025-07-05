package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/www-xu/spark/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("gin-server")

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span for every request.
		// If a trace context is extracted, it becomes the parent. Otherwise, a new trace is created.
		ctx, span := tracer.Start(ctx, c.Request.URL.Path)
		defer span.End()

		spanCtx := trace.SpanContextFromContext(ctx)
		traceID := spanCtx.TraceID().String()
		spanID := spanCtx.SpanID().String()

		c.Set(string(log.TraceIdKey), traceID)
		c.Set(string(log.SpanIdKey), spanID)

		// For consistency, let's keep request_id, but it's not part of OpenTelemetry spec
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(string(log.RequestIdKey), requestID)

		// Propagate the trace context in the response header
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		latency := time.Since(start)
		log.WithContext(c).WithFields(map[string]interface{}{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
		}).Info("request processed")
	}
}
