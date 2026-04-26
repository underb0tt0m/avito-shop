package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"
	"avito-shop/internal/tools"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Main struct {
	service      service.API
	logger       logging.Logger
	jsonCodec    tools.JSONCodec
	queryTimeout time.Duration
	tokenMaker   tools.TokenMaker
}

func NewMain(
	service service.API,
	logger logging.Logger,
	jsonCodec tools.JSONCodec,
	queryTimeout time.Duration,
	tokenMaker tools.TokenMaker,
) Main {
	return Main{
		service:      service,
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
		tokenMaker:   tokenMaker,
	}
}

func (h Main) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(api_middleware.Auth(h.logger, h.tokenMaker))

		h.GetInfo(r)
		h.SendCoin(r)
		h.BuyItem(r)
	})
}

func (h Main) GetInfo(r chi.Router) {
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		h.logger.Debug("JWT token is valid")
		username := user.UserName

		ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
		defer cancel()
		h.logger.Debug("calling mainRoutService GetUserInfo method")
		dtoUser, err := h.service.GetUserInfo(ctx, username)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		response, err := h.jsonCodec.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			h.logger.Error(
				"failed to marshal user info response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
		if _, err = w.Write(response); err != nil {
			h.logger.Error(
				"failed to write info response",
				err,
			)
			tools.WriteError(w, err)
			return
		}
	})
}

func (h Main) SendCoin(r chi.Router) {
	r.Post("/sendCoin", func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			h.logger.Error(
				"failed to read request body",
				err,
			)
			tools.WriteError(w, err)
			return
		}

		transaction := dto.SendCoinRequest{}
		if err = h.jsonCodec.Unmarshal(
			requestBody,
			&transaction,
		); err != nil {
			h.logger.Error(
				"failed to unmarshal request body",
				err,
			)
			tools.WriteError(w, domain.ErrUnprocessableEntity)
			return
		}

		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
		defer cancel()
		if err = h.service.SendCoins(
			ctx,
			user.UserName,
			transaction,
		); err != nil {
			tools.WriteError(w, err)
			return
		}
	})
}

func (h Main) BuyItem(r chi.Router) {
	r.Post("/buy/{itemID}", func(w http.ResponseWriter, r *http.Request) {
		user, err := tools.GetUserFromContext(r)
		if err != nil {
			tools.WriteError(w, err)
			return
		}

		strItemID := chi.URLParam(r, "itemID")
		if strItemID == "" {
			h.logger.Warn(
				"attempt to buy item with empty param {item}",
				domain.ErrBadRequest,
			)
			tools.WriteError(w, domain.ErrBadRequest)
			return
		}

		itemID, err := strconv.Atoi(strItemID)
		if err != nil {
			h.logger.Warn(
				"attempt to buy item with invalid id",
				domain.ErrBadRequest,
			)
			tools.WriteError(w, domain.ErrBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
		defer cancel()
		if err = h.service.BuyItem(ctx, itemID, user.UserName); err != nil {
			tools.WriteError(w, err)
			return
		}
	})
}
