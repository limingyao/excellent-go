package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPHandleCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_handles_total",
			Help: "number of http handles",
		},
		// method: post/get ...
		[]string{"addr", "method", "path"},
	)
	HTTPHandleDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_handle_duration_milliseconds",
			Help:       "http handle duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		// method: post/get ...
		[]string{"addr", "method", "path"},
	)
	HTTPRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "number of http requests",
		},
		// method: post/get ...
		[]string{"code", "id", "addr", "method", "path", "state"},
	)
	HTTPRequestDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_requests_duration_milliseconds",
			Help:       "http request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		// method: post/get ...
		[]string{"code", "id", "addr", "method", "path"},
	)
)

const (
	HTTPPostMethod   = "post"
	HTTPGetMethod    = "get"
	HTTPPutMethod    = "put"
	HTTPPatchMethod  = "patch"
	HTTPDeleteMethod = "delete"
)

func init() {
	prometheus.MustRegister(HTTPHandleCounter)
	prometheus.MustRegister(HTTPHandleDuration)
	prometheus.MustRegister(HTTPRequestCounter)
	prometheus.MustRegister(HTTPRequestDuration)
}
