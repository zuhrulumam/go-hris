package metrics

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	slokmid "github.com/slok/go-http-metrics/middleware"
)

type (
	SkipHandler struct {
		Path   string
		Method string
	}

	metricMiddleware struct {
		ServiceName    string
		SkipHandler    map[string]bool
		RequestCounter *prometheus.CounterVec
	}

	reporter struct {
		c *gin.Context
	}
)

type GinMiddleware interface {
	Use() gin.HandlerFunc
	Skip(handler []SkipHandler)
	RequestCounterMiddleware() gin.HandlerFunc
}

func (r *reporter) Method() string           { return r.c.Request.Method }
func (r *reporter) Context() context.Context { return r.c.Request.Context() }
func (r *reporter) URLPath() string          { return r.c.FullPath() }
func (r *reporter) StatusCode() int          { return r.c.Writer.Status() }
func (r *reporter) BytesWritten() int64      { return int64(r.c.Writer.Size()) }

// Use : handler metric
func (middleware metricMiddleware) Use() gin.HandlerFunc {
	mm := slokmid.New(slokmid.Config{
		Service:  middleware.ServiceName,
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	return func(ctx *gin.Context) {
		// if this handler is skipped
		// then skip this middleware
		if middleware.SkipHandler[ctx.FullPath()] {
			ctx.Next()
			return
		}

		// filter by path
		if _, disallowMetrics := middleware.SkipHandler[ctx.Request.URL.Path]; disallowMetrics {
			ctx.Next()
			return
		}

		rep := &reporter{c: ctx}
		mm.Measure(rep.c.FullPath(), rep, func() {
			ctx.Next()
		})
	}
}

func (middleware metricMiddleware) RequestCounterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the client info from header or IP
		client := c.Request.Header.Get("X-Client-ID")
		if client == "" {
			client = c.ClientIP() // Use IP if no client ID is provided
		}

		// Get the current request path
		path := c.FullPath()

		// Skip counting for specific endpoints
		if path == "/health" || path == "/metrics" || path == "/" || path == "" {
			// Process the request without incrementing the counter
			c.Next()
			return
		}

		// Increment the request counter
		middleware.RequestCounter.With(prometheus.Labels{
			"client":   client,
			"endpoint": path,
			"method":   c.Request.Method,
		}).Inc()

		// Process the request
		c.Next()
	}
}

func (middleware *metricMiddleware) Skip(handlers []SkipHandler) {
	for _, handler := range handlers {
		middleware.SkipHandler[handler.Path] = true
	}
}

func NewMetricsMiddleware(serviceName string) GinMiddleware {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_counter",
			Help: "Total number of HTTP requests, labeled by endpoint, method, client, auth_domain_code",
		},
		[]string{"client", "endpoint", "method", "auth_domain_code"},
	)

	prometheus.MustRegister(requestCounter)

	return &metricMiddleware{
		ServiceName:    serviceName,
		SkipHandler:    make(map[string]bool, 0),
		RequestCounter: requestCounter,
	}
}

func Init(router *gin.Engine, handlers []SkipHandler, serviceName string) {
	mm := NewMetricsMiddleware(serviceName)

	// Setup skip middleware red metrics
	mm.Skip(handlers)

	// use red metrics gin
	router.Use(mm.Use())

	router.Use(mm.RequestCounterMiddleware())

	// setup router /metrics for prometheus
	router.GET("/metrics", gin.WrapH(
		promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{DisableCompression: true},
		),
	))
}
