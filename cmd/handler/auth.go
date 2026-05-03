package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/logging"
	"avito-shop/internal/prometheus_metrics"
	"avito-shop/internal/service"

	"github.com/go-chi/chi/v5"
)

type Auth struct {
	service      service.Auth
	logger       logging.Logger
	jsonCodec    jsoncodec.JSONCodec
	queryTimeout time.Duration
	metrics      *prometheus_metrics.Metrics
}

//nolint:revive
func NewAuth(
	authService service.Auth,
	logger logging.Logger,
	jsonCodec jsoncodec.JSONCodec,
	queryTimeout time.Duration,
	metrics *prometheus_metrics.Metrics,
) Auth {
	return Auth{
		service:      authService,
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
		metrics:      metrics,
	}
}

func (h Auth) RegisterRoutes(r chi.Router) {
	h.Auth(r)
}

func (h Auth) Auth(r chi.Router) {
	r.Post("/auth", h.handleAuth)
}

func (h Auth) handleAuth(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	defer func() {
		if err = r.Body.Close(); err != nil {
			h.logger.Errorf(
				err,
				"failed to close auth request body",
			)
		}
	}()
	if err != nil {
		h.logger.Errorf(
			err,
			"failed to read auth request body",
		)
		domain.WriteError(w, err, h.logger, h.metrics)
		return
	}

	user := dto.AuthRequest{}
	if err = h.jsonCodec.Unmarshal(
		requestBody,
		&user,
	); err != nil {
		h.logger.Errorf(
			err,
			"failed to unmarshal auth request body",
		)
		domain.WriteError(w, domain.ErrUnprocessableEntity, h.logger, h.metrics)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	h.logger.Debugf(
		"calling AuthService Auth method, username: %v",
		user.Name,
	)
	token, err := h.service.Auth(ctx, user)
	if err != nil {
		h.logger.Warnf(
			err,
			"authentication denied, username: %v",
			user.Name,
		)
		domain.WriteError(w, err, h.logger, h.metrics)
		return
	}

	responseBody, err := h.jsonCodec.Marshal(token)
	if err != nil {
		h.logger.Errorf(
			err,
			"failed to marshal auth response body",
		)
		domain.WriteError(w, err, h.logger, h.metrics)
		return
	}

	if _, err = w.Write(responseBody); err != nil {
		h.logger.Errorf(
			err,
			"failed to write auth response",
		)
		domain.WriteError(w, err, h.logger, h.metrics)
		return
	}
}
