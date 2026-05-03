package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"avito-shop/internal/prometheus_metrics"
)

type Metrics struct {
	metrics *prometheus_metrics.Metrics
	reg     *prometheus.Registry
}

func NewMetrics(
	metrics *prometheus_metrics.Metrics,
	reg *prometheus.Registry,
) Metrics {
	return Metrics{
		metrics: metrics,
		reg:     reg,
	}
}

func (h Metrics) RegisterRoutes(r chi.Router) {
	r.Handle("/metrics", promhttp.HandlerFor(h.reg, promhttp.HandlerOpts{
		Registry: h.reg,
	}))
}
