package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/www-xu/spark/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("gin-server")

func Observability() gin.HandlerFunc {
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

		// Add all IDs to the context using the standard context.WithValue pattern.
		ctx = context.WithValue(ctx, log.TraceIdKey, traceID)
		ctx = context.WithValue(ctx, log.SpanIdKey, spanID)

		// For consistency, let's keep request_id, but it's not part of OpenTelemetry spec
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, log.RequestIdKey, requestID)

		// Propagate all relevant IDs in the response headers.
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
		// Place the final, enriched context into the request.
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		latency := time.Since(start)
		// Use the request's context, which now contains all our IDs.
		log.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
		}).Info("request processed")
	}
}
