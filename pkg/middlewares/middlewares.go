package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zuhrulumam/go-hris/pkg/ctxkeys"
	"go.uber.org/zap"
)

func RequestContextMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate correlation ID if not present
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Create context with values
		ctx := context.WithValue(c.Request.Context(), ctxkeys.CtxKeyCorrelationID, correlationID)
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyApp, "attendance-service")
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyRuntime, "go")
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyEnv, "production") // or from env
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyAppVersion, "v1.0.0")
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyPath, c.FullPath())
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyMethod, c.Request.Method)
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyIP, c.ClientIP())
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyPort, c.Request.URL.Port())
		ctx = context.WithValue(ctx, ctxkeys.CtxKeySrcIP, c.Request.RemoteAddr)
		ctx = context.WithValue(ctx, ctxkeys.CtxKeyHeader, c.Request.Header)

		// Attach context back to request
		c.Request = c.Request.WithContext(ctx)

		// Continue to next middleware/handler
		c.Next()

		// Logging after response
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("HTTP Request",
			zap.String("path", c.FullPath()),
			zap.String("method", c.Request.Method),
			zap.String("correlation_id", correlationID),
			zap.Int("status", statusCode),
			zap.String("duration", duration.String()),
			zap.Any("header", c.Request.Header),
		)
	}
}
