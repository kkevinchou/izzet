package globals

import "github.com/kkevinchou/kitolib/metrics"

var (
	clientMetricsRegistry *metrics.MetricsRegistry
)

func SetClientMetricsRegistry(metricsRegistry *metrics.MetricsRegistry) {
	clientMetricsRegistry = metricsRegistry
}

func GetClientMetricsRegistry() *metrics.MetricsRegistry {
	return clientMetricsRegistry
}
