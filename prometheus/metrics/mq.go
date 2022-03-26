package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	MQCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mq_requests_total",
			Help: "number of mq requests",
		},
		// method: produce/consume ...
		[]string{"code", "id", "addr", "method", "topic", "state"},
	)
	MQDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "mq_request_duration_milliseconds",
			Help:       "mq request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			MaxAge:     time.Minute,
		},
		// method: produce/consume ...
		[]string{"code", "id", "addr", "method", "topic"},
	)
)

const (
	MQProduceMethod = "produce"
	MQConsumeMethod = "consume"
)

func init() {
	prometheus.MustRegister(MQCounter)
	prometheus.MustRegister(MQDuration)
}
