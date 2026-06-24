package metrics

import (
	"time"
)

func MeasureDBQuery(operation string, registry *MetricsRegistry, fn func() error) error {
	if registry == nil {
		return fn()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	registry.DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
	return err
}
