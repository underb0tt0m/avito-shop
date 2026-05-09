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
	RequestDuration *prometheus.HistogramVec
	AuthAttempts    *prometheus.CounterVec
	RegisteredUsers prometheus.Counter
	Transactions    *prometheus.CounterVec
	Purchases       *prometheus.CounterVec
	Errors          *prometheus.CounterVec
	RequestSize     *prometheus.CounterVec
	ResponseSize    *prometheus.CounterVec
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "echo_request_duration_seconds",
			Buckets: []float64{0.01, 0.025, 0.05, 0.075, 0.1, 0.25},
		},
			[]string{"instance", "code"},
		),
		AuthAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "auth_attempts_total",
		},
			[]string{"code"},
		),
		RegisteredUsers: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "users_registered_total",
		}),
		Transactions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "transactions_total",
		},
			[]string{"code"},
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
			[]string{"code"},
		),
		RequestSize: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "echo_request_size_bytes_sum",
			Help: "Total transferred request data",
		},
			[]string{"instance", "code"},
		),
		ResponseSize: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "echo_response_size_bytes_sum",
			Help: "Total transferred response data",
		},
			[]string{"instance", "code"},
		),
	}

	reg.MustRegister(m.RequestDuration)
	reg.MustRegister(m.AuthAttempts)
	reg.MustRegister(m.RegisteredUsers)
	reg.MustRegister(m.Transactions)
	reg.MustRegister(m.Purchases)
	reg.MustRegister(m.Errors)
	reg.MustRegister(m.RequestSize)
	reg.MustRegister(m.ResponseSize)

	return m
}
