package metrics

import (
	"github.com/gracew/widget-proxy/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	CREATE = "create"
	READ = "read"
	LIST = "list"
	DELETE = "delete"
)
var (
	objectives = map[float64]float64{0.5: 0.05, 0.75: .025, 0.9: 0.01, 0.95: .005, 0.99: 0.001}
	customLogicLabels = []string{"method", "when"}

	RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.APIName,
		Name: "http_requests_total",
	},[]string{"method"})
	RequestSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: config.APIName,
		Name: "http_request_duration_seconds",
		Objectives: objectives,
	}, []string{"method"})

	CustomLogicSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: config.APIName,
		Name: "custom_logic_duration_seconds",
		Objectives: objectives,
	}, customLogicLabels)
	CustomLogicErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.APIName,
		Name: "custom_logic_errors_total",
	}, customLogicLabels)

	DatabaseSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: config.APIName,
		Name: "database_access_duration_seconds",
		Objectives: objectives,
	}, []string{"method"})
	DatabaseErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.APIName,
		Name: "database_access_errors_total",
	}, []string{"method"})
)
