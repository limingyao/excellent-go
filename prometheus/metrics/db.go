package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	DBCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_requests_total",
			Help: "number of db requests",
		},
		// method: insert/delete/select/update ...
		[]string{"code", "id", "addr", "method", "table", "state"},
	)
	DBDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "db_request_duration_milliseconds",
			Help:       "db request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		// method: insert/delete/select/update ...
		[]string{"code", "id", "addr", "method", "table"},
	)
)

const (
	DBInsertMethod = "insert"
	DBDeleteMethod = "delete"
	DBSelectMethod = "select"
	DBUpdateMethod = "update"
)

func init() {
	prometheus.MustRegister(DBCounter)
	prometheus.MustRegister(DBDuration)
}
