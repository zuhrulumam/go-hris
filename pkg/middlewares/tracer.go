package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // capture to buffer
	return w.ResponseWriter.Write(b) // write to original
}

func TracerLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)

		if span == nil || !span.IsRecording() {
			c.Next()
			return
		}

		// Wrap response writer
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		// Read request body if JSON
		var requestBody string
		if strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
			if bodyBytes, err := io.ReadAll(c.Request.Body); err == nil {
				requestBody = string(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				span.RecordError(err)
			}
		}

		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic: %v", rec)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.AddEvent("panic recovered", trace.WithAttributes(
					attribute.String("panic", fmt.Sprint(rec)),
				))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}

			// Add request + response attributes
			attrs := []attribute.KeyValue{
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.RequestURI),
				attribute.Int("http.status_code", c.Writer.Status()),
				attribute.String("request.body", requestBody),
				attribute.String("response.body", bw.body.String()),
			}

			if len(c.Errors) > 0 {
				for _, ginErr := range c.Errors {
					span.RecordError(ginErr)
					span.AddEvent("handler error", trace.WithAttributes(
						attribute.String("error", ginErr.Error()),
					))
				}
				span.SetStatus(codes.Error, c.Errors.Last().Error())
			} else if c.Writer.Status() >= 500 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			}

			span.AddEvent("http.transaction", trace.WithAttributes(attrs...))
		}()

		c.Next()
	}
}
