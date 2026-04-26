package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"
	"avito-shop/internal/tools"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Auth struct {
	service      service.Auth
	logger       logging.Logger
	jsonCodec    tools.JSONCodec
	queryTimeout time.Duration
}

func NewAuth(
	service service.Auth,
	logger logging.Logger,
	jsonCodec tools.JSONCodec,
	queryTimeout time.Duration,
) Auth {
	return Auth{
		service:      service,
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
	}
}

func (h Auth) RegisterRoutes(r chi.Router) {
	h.Auth(r)
}

func (h Auth) Auth(r chi.Router) {
	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			h.logger.Error(
				"failed to read auth request body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		user := dto.AuthRequest{}
		if err = h.jsonCodec.Unmarshal(
			requestBody,
			&user,
		); err != nil {
			h.logger.Error(
				"failed to unmarshal auth request body",
				err,
			)
			tools.WriteError(w, domain.ErrUnprocessableEntity)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
		defer cancel()
		h.logger.Debug(
			fmt.Sprintf(
				"calling AuthService Auth method, username: %v",
				user.Name,
			),
		)
		token, err := h.service.Auth(ctx, user)
		if err != nil {
			h.logger.Warn(
				fmt.Sprintf(
					"authentication denied, username: %v",
					user.Name,
				),
				err,
			)
			tools.WriteError(w, err)
			return
		}

		responseBody, err := h.jsonCodec.Marshal(token)
		if err != nil {
			h.logger.Error(
				"failed to marshal auth response body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		if _, err = w.Write(responseBody); err != nil {
			h.logger.Error(
				"failed to write auth response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
	})
}
