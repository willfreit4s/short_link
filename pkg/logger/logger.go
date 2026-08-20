// Package logger defines middleware for logging requests in Gin
package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/willfreit4s/short_link/configs"

	"github.com/gin-gonic/gin"
)

type contextKey struct{}

const LoggerKey = "logger"

func InitLogger(cfg *configs.Config) (*slog.Logger, error) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	log.Info(
		"Starting application",
		"service_name", cfg.ServiceName,
		"environment", cfg.Environment,
		"server_port", cfg.ServerPort,
	)

	return log, nil
}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID, _ := c.Get("request_id")

		reqLogger := logger.With(
			slog.String("path", c.Request.URL.Path),
			slog.String("method", c.Request.Method),
			slog.String("request_id", requestID.(string)),
		)

		c.Request = c.Request.WithContext(WithContext(c.Request.Context(), reqLogger))
		c.Set(LoggerKey, reqLogger)

		c.Next()

		reqLogger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("agent", c.Request.UserAgent()),
		)
	}
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}

	return slog.Default()
}
