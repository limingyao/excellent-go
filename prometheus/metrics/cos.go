package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	COSCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cos_requests_total",
			Help: "number of cos requests",
		},
		// method: put/delete/copy/get ...
		[]string{"code", "id", "addr", "method", "state"},
	)
	COSDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "cos_request_duration_milliseconds",
			Help:       "cos request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		// method: put/delete/copy/get ...
		[]string{"code", "id", "addr", "method"},
	)
)

const (
	COSPutMethod      = "put"
	COSDeleteMethod   = "delete"
	COSCopyMethod     = "copy"
	COSGetMethod      = "get"
	COSDownloadMethod = "download"
)

func init() {
	prometheus.MustRegister(COSCounter)
	prometheus.MustRegister(COSDuration)
}
