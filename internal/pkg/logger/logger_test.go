package logger_test

import (
	"context"
	"testing"

	"github.com/JLugagne/ha-dash/internal/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerFromContext(t *testing.T) {
	t.Run("returns default logger when nil context", func(t *testing.T) {
		entry := logger.LoggerFromContext(nil)
		assert.NotNil(t, entry)
	})

	t.Run("returns default logger when not in context", func(t *testing.T) {
		ctx := context.Background()
		entry := logger.LoggerFromContext(ctx)
		assert.NotNil(t, entry)
	})

	t.Run("returns custom logger when stored in context", func(t *testing.T) {
		custom := logrus.New().WithField("service", "test")
		ctx := logger.ContextWithLogger(context.Background(), custom)
		entry := logger.LoggerFromContext(ctx)
		assert.Equal(t, custom, entry)
	})
}
