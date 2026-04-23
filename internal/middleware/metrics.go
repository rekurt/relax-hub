package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Exposed request metrics.
//
// Label "path" is bound to chi.RouteContext(r.Context()).RoutePattern() —
// the parameterised pattern (e.g. "/api/v1/bathhouses/{id}") rather than
// the concrete URL, otherwise high-cardinality identifiers (UUIDs, slugs)
// would explode the Prometheus series count.
var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests handled by the API, labelled by method, route pattern and response status.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds, labelled by method and route pattern.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being served by the API.",
	})
)

// Metrics instruments Prometheus counters/histograms for every HTTP request.
//
// Must be registered after chi's routing middleware (RequestID, Logging) so
// that RouteContext is populated and a WrapResponseWriter is available to
// observe the status code.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		pattern := chi.RouteContext(r.Context()).RoutePattern()
		if pattern == "" {
			// For unmatched routes (404/405) use a single label to avoid cardinality blow-up.
			pattern = "unmatched"
		}

		status := strconv.Itoa(ww.Status())
		httpRequestsTotal.WithLabelValues(r.Method, pattern, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
	})
}
