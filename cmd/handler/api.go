package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"
	"avito-shop/internal/tools"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Main(s service.API, r chi.Router, logger logging.Logger) {
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		logger.Debug("JWT token is valid")
		username := user.UserName

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug("calling mainRoutService GetUserInfo method")
		dtoUser, err := s.GetUserInfo(ctx, username)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		response, err := json.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			logger.Error(
				"failed to marshal user info response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
		if _, err = w.Write(response); err != nil {
			logger.Error(
				"failed to write info response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
	})

	r.Post("/sendCoin", func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			logger.Error(
				"failed to read request body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		transaction := dto.SendCoinRequest{}
		if err = json.Unmarshal(
			requestBody,
			&transaction,
		); err != nil {
			logger.Error(
				"failed to unmarshal request body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.SendCoins(
			ctx,
			user.UserName,
			transaction,
		); err != nil {
			tools.WriteError(w, err)
			return
		}
	})

	r.Post("/buy/{itemID}", func(w http.ResponseWriter, r *http.Request) {
		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		strItemID := chi.URLParam(r, "itemID")
		if strItemID == "" {
			logger.Warn(
				"attempt to buy item with empty param {item}",
				domain.ErrBadRequest,
			)
			tools.WriteError(w, domain.ErrBadRequest)
			return
		}

		itemID, err := strconv.Atoi(strItemID)
		if err != nil {
			logger.Warn(
				"attempt to buy item with invalid id",
				domain.ErrBadRequest,
			)
			tools.WriteError(w, domain.ErrBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.BuyItem(ctx, itemID, user.UserName); err != nil {
			tools.WriteError(w, err)
			return
		}
	})
}
