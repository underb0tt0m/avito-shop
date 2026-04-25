package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"
	"avito-shop/internal/tools"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Auth(s service.Auth, r chi.Router, logger logging.Logger, jsonCodec tools.JSONCodec) {
	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			logger.Error(
				"failed to read auth request body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		user := dto.AuthRequest{}
		if err = jsonCodec.Unmarshal(
			requestBody,
			&user,
		); err != nil {
			logger.Error(
				"failed to unmarshal auth request body",
				err,
			)
			tools.WriteError(w, domain.ErrUnprocessableEntity)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug(
			fmt.Sprintf(
				"calling AuthService Auth method, username: %v",
				user.Name,
			),
		)
		token, err := s.Auth(ctx, user)
		if err != nil {
			logger.Warn(
				fmt.Sprintf(
					"authentication denied, username: %v",
					user.Name,
				),
				err,
			)
			tools.WriteError(w, err)
			return
		}

		responseBody, err := jsonCodec.Marshal(token)
		if err != nil {
			logger.Error(
				"failed to marshal auth response body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		if _, err = w.Write(responseBody); err != nil {
			logger.Error(
				"failed to write auth response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
	})
}
