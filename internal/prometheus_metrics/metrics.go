package prometheus_metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type contextKey string

const (
	ReqStartTimeContextKey   contextKey = "request_start_time"
	StatusAuthSuccess        string     = "success"
	StatusAuthFailed         string     = "failure"
	StatusTransactionSuccess string     = "success"
	StatusTransactionFailed  string     = "failure"
)

type Metrics struct {
	RequestCounter  prometheus.Counter
	RequestDuration prometheus.Histogram
	AuthAttempts    *prometheus.CounterVec
	RegisteredUsers prometheus.Counter
	Transactions    *prometheus.CounterVec
	Purchases       *prometheus.CounterVec
	Errors          *prometheus.CounterVec
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "Current count of requests since last app start",
		}),
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Buckets: []float64{0.01, 0.025, 0.05, 0.075, 0.1, 0.25},
		}),
		AuthAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "auth_attempts_total",
		},
			[]string{"Status"},
		),
		RegisteredUsers: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "users_registered_total",
		}),
		Transactions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "transactions_total",
		},
			[]string{"Status"},
		),
		Purchases: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "purchases_total",
		},
			[]string{"ID"},
		),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "errors_user_total",
			Help: "Уrrors caused by the client",
		},
			[]string{"Status"},
		),
	}

	reg.MustRegister(m.RequestCounter)
	reg.MustRegister(m.RequestDuration)
	reg.MustRegister(m.AuthAttempts)
	reg.MustRegister(m.RegisteredUsers)
	reg.MustRegister(m.Transactions)
	reg.MustRegister(m.Purchases)
	reg.MustRegister(m.Errors)

	return m
}
