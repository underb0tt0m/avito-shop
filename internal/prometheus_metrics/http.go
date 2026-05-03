package prometheus_metrics

import (
	"errors"
	"net/http"
	"time"

	"avito-shop/internal/logging"
)

func Middlware(m *Metrics, logger logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.RequestCounter.Inc()
			next.ServeHTTP(w, r)
			startRaw := r.Context().Value(ReqStartTimeContextKey)
			startTime, ok := startRaw.(time.Time)
			if !ok {
				logger.Errorf(
					errors.New("failed to get request start time from context"),
					"failed to get request start time from context",
				)
				return
			}
			duration := time.Since(startTime).Seconds()
			m.RequestDuration.Observe(duration)
		})
	}
}
