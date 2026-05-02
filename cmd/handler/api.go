package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"avito-shop/cmd/dto"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/domain"
	jsoncodec "avito-shop/internal/jsoncodec"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
	"avito-shop/internal/service"

	"github.com/go-chi/chi/v5"
)

type Main struct {
	service      service.API
	logger       logging.Logger
	jsonCodec    jsoncodec.JSONCodec
	queryTimeout time.Duration
	tokenMaker   jwtmanager.TokenMaker
}

//nolint:revive
func NewMain(
	service service.API,
	logger logging.Logger,
	jsonCodec jsoncodec.JSONCodec,
	queryTimeout time.Duration,
	tokenMaker jwtmanager.TokenMaker,
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
	r.Get("/info", h.handleInfo)
}

func (h Main) handleInfo(w http.ResponseWriter, r *http.Request) {
	user, err := domain.GetUserFromContext(r)
	if err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}

	h.logger.Debugf("JWT token is valid")
	username := user.UserName

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	h.logger.Debugf("calling mainRoutService GetUserInfo method")
	dtoUser, err := h.service.GetUserInfo(ctx, username)
	if err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}

	response, err := h.jsonCodec.MarshalIndent(dtoUser, "", "	")
	if err != nil {
		h.logger.Errorf(
			err,
			"failed to marshal user info response",
		)
		domain.WriteError(w, err, h.logger)
		return
	}
	if _, err = w.Write(response); err != nil {
		h.logger.Errorf(
			err,
			"failed to write info response",
		)
		domain.WriteError(w, err, h.logger)
		return
	}
}

func (h Main) SendCoin(r chi.Router) {
	r.Post("/sendCoin", h.handleSendCoin)
}

func (h Main) handleSendCoin(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	defer func() {
		if err = r.Body.Close(); err != nil {
			h.logger.Errorf(
				err,
				"failed to close request body",
			)
		}
	}()
	if err != nil {
		h.logger.Errorf(
			err,
			"failed to read request body",
		)
		domain.WriteError(w, err, h.logger)
		return
	}

	transaction := dto.SendCoinRequest{}
	if err = h.jsonCodec.Unmarshal(
		requestBody,
		&transaction,
	); err != nil {
		h.logger.Errorf(
			err,
			"failed to unmarshal request body",
		)
		domain.WriteError(w, domain.ErrUnprocessableEntity, h.logger)
		return
	}

	user, err := domain.GetUserFromContext(r)
	if err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	if err = h.service.SendCoins(
		ctx,
		user.UserName,
		transaction,
	); err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}
}

func (h Main) BuyItem(r chi.Router) {
	r.Post("/buy/{itemID}", h.handleBuyItem)
}

func (h Main) handleBuyItem(w http.ResponseWriter, r *http.Request) {
	user, err := domain.GetUserFromContext(r)
	if err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}

	strItemID := chi.URLParam(r, "itemID")
	if strItemID == "" {
		h.logger.Warnf(
			domain.ErrBadRequest,
			"attempt to buy item with empty param {item}",
		)
		domain.WriteError(w, domain.ErrBadRequest, h.logger)
		return
	}

	itemID, err := strconv.Atoi(strItemID)
	if err != nil {
		h.logger.Warnf(
			domain.ErrBadRequest,
			"attempt to buy item with invalid id",
		)
		domain.WriteError(w, domain.ErrBadRequest, h.logger)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	if err = h.service.BuyItem(ctx, itemID, user.UserName); err != nil {
		domain.WriteError(w, err, h.logger)
		return
	}
}
