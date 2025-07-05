package log

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	TraceIdKey   contextKey = "trace_id"
	RequestIdKey contextKey = "request_id"
	SpanIdKey    contextKey = "span_id"
)

// contextHook is a logrus hook that adds trace_id and request_id from the context to log entries.
type contextHook struct{}

// Levels returns all log levels, indicating this hook will fire on all of them.
func (h *contextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus for each log entry. It extracts context values and adds them as fields.
func (h *contextHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	if traceId := entry.Context.Value(TraceIdKey); traceId != nil {
		entry.Data["trace_id"] = traceId
	}
	if requestId := entry.Context.Value(RequestIdKey); requestId != nil {
		entry.Data["request_id"] = requestId
	}
	if spanId := entry.Context.Value(SpanIdKey); spanId != nil {
		entry.Data["span_id"] = spanId
	}

	return nil
}

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := frame.File
			// 1. Remove module version string from path if present.
			if atIndex := strings.Index(fileName, "@"); atIndex != -1 {
				if slashIndex := strings.Index(fileName[atIndex:], "/"); slashIndex != -1 {
					fileName = fileName[:atIndex] + fileName[atIndex+slashIndex:]
				}
			}

			// 2. Take the last 3 path components.
			parts := strings.Split(fileName, "/")
			if len(parts) > 3 {
				fileName = strings.Join(parts[len(parts)-3:], "/")
			} else if len(parts) > 1 {
				fileName = strings.Join(parts[len(parts)-2:], "/")
			}

			return frame.Function, fmt.Sprintf("%s:%d", fileName, frame.Line)
		},
	})
	logrus.AddHook(&contextHook{})
}

// WithContext returns a new logrus entry with the provided context for subsequent chained logging calls.
func WithContext(ctx context.Context) *logrus.Entry {
	return logrus.WithContext(ctx)
}

// Info logs a message at level Info.
func Info(args ...interface{}) {
	logrus.Info(args...)
}

// Infof logs a formatted message at level Info.
func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// Warn logs a message at level Warn.
func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

// Warnf logs a formatted message at level Warn.
func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

// Error logs a message at level Error.
func Error(args ...interface{}) {
	logrus.Error(args...)
}

// Errorf logs a formatted message at level Error.
func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

// Fatal logs a message at level Fatal then the process will exit.
func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

// Fatalf logs a formatted message at level Fatal then the process will exit.
func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

// Panic logs a message at level Panic then the process will panic.
func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

// Panicf logs a formatted message at level Panic then the process will panic.
func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}
