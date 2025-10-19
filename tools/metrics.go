package tools

import (
	"log/slog"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	module = "METRICS"
)

func RunMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	slog.Info("started metrics server", "tag", module)
	slog.Error(http.ListenAndServe(":9100", nil).Error(), "tag", module)
}