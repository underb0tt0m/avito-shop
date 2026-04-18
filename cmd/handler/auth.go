package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Auth(s service.Auth, r chi.Router, logger logging.Logger) {
	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			fmt.Sprintf(
				"request received, method: %v, pattern: %v, remoteAddr: %v",
				r.Method,
				r.Pattern,
				r.RemoteAddr,
			),
		)

		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			logger.Error(
				"failed to read auth request body",
				err,
			)
			WriteError(w, err)
			return
		}

		user := dto.AuthRequest{}
		if err = json.Unmarshal(
			requestBody,
			&user,
		); err != nil {
			logger.Error(
				"failed to unmarshal auth request body",
				err,
			)
			WriteError(w, err)
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
			WriteError(w, err)
			return
		}

		responseBody, err := json.Marshal(token)
		if err != nil {
			logger.Error(
				"failed to marshal auth response body",
				err,
			)
			WriteError(w, err)
			return
		}

		if _, err = w.Write(responseBody); err != nil {
			logger.Error(
				"failed to write auth response",
				err,
			)
			WriteError(w, err)
			return
		}

		logger.Debug(
			fmt.Sprintf(
				"request has been processed, status: %v, processing time: %v",
				http.StatusOK,
				time.Since(start),
			),
		)
	})
}
