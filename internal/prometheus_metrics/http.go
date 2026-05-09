package prometheus_metrics

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"avito-shop/internal/logging"
)

type ResponseWriter struct {
	http.ResponseWriter
	Status   int
	BodySize int
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.BodySize += len(data)
	return w.ResponseWriter.Write(data)
}

func Middlware(m *Metrics, logger logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rWriter := &ResponseWriter{
				ResponseWriter: w,
				Status:         http.StatusOK,
				BodySize:       0,
			}
			next.ServeHTTP(rWriter, r)
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
			strStatus := strconv.Itoa(rWriter.Status)
			responseLength := rWriter.BodySize

			m.RequestDuration.WithLabelValues("app:8080", strStatus).Observe(duration)

			m.RequestSize.WithLabelValues("app:8080", strStatus).Add(float64(r.ContentLength))
			m.ResponseSize.WithLabelValues("app:8080", strStatus).Add(float64(responseLength))
		})
	}
}
