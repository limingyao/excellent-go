package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	KVCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kv_requests_total",
			Help: "number of kv requests",
		},
		// method: set/get ...
		[]string{"code", "id", "addr", "method", "state"},
	)
	KVDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "kv_request_duration_milliseconds",
			Help:       "kv request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			MaxAge:     time.Minute,
		},
		// method: set/get ...
		[]string{"code", "id", "addr", "method"},
	)
)

const (
	KVSetMethod = "set"
	KVGetMethod = "get"
)

func init() {
	prometheus.MustRegister(KVCounter)
	prometheus.MustRegister(KVDuration)
}
