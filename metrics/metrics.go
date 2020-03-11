package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	apiName = os.Getenv("API_NAME")
	objectives = map[float64]float64{0.5: 0.05, 0.75: .025, 0.9: 0.01, 0.95: .005, 0.99: 0.001}
	customLogicLabels = []string{"method", "when"}

	RequestSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: apiName,
		Name: "http_request_duration_ms",
		Objectives: objectives,
	}, []string{"method"})

	CustomLogicSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: apiName,
		Name: "custom_logic_duration_ms",
		Objectives: objectives,
	}, customLogicLabels)
	CustomLogicErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: apiName,
		Name: "custom_logic_errors",
	}, customLogicLabels)

	DatabaseSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: apiName,
		Name: "database_access_duration_ms",
		Objectives: objectives,
	}, []string{"method"})
	DatabaseErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: apiName,
		Name: "database_access_errors",
	}, []string{"method"})
)
