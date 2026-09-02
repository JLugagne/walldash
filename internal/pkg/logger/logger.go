package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextKey string

const loggerContextKey contextKey = "ha-dash-logger"

// LoggerFromContext retrieves the logger entry from the context, or returns a default logger entry.
func LoggerFromContext(ctx context.Context) *logrus.Entry {
	if ctx != nil {
		if entry, ok := ctx.Value(loggerContextKey).(*logrus.Entry); ok && entry != nil {
			return entry
		}
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

// ContextWithLogger associates a logger entry with the context.
func ContextWithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerContextKey, entry)
}
