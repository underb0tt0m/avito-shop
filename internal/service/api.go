package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=api.go -destination=../mocks/service_api.go -package=mocks -mock_names=API=MockServiceAPI
type API interface {
	GetUserInfo(ctx context.Context, username string) (*dto.InfoResponse, error)
	SendCoins(ctx context.Context, fromUser string, toUser dto.SendCoinRequest) error
	BuyItem(ctx context.Context, itemID int, user string) error
}
type api struct {
	Storage storage.API
	Logger  logging.Logger
}

func NewApi(s storage.API, l logging.Logger) API {
	return api{
		Storage: s,
		Logger:  l,
	}
}

func (s api) GetUserInfo(ctx context.Context, username string) (*dto.InfoResponse, error) {
	userBalance, userInventories, userTransactions, err := s.Storage.GetUserInfo(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.Logger.Error(
				"user is missing from the database",
				err,
			)
			return nil, domain.ErrNotFound
		}
		s.Logger.Error(
			"failed to get user info from mainRoutRepository",
			err,
		)
		return nil, err
	}
	s.Logger.Debug("received data from mainRoutRepository")

	var userInventory []domain.Item
	for _, item := range userInventories {
		userInventory = append(userInventory, domain.Item{
			ObjType:  item.ItemName,
			Quantity: item.Quantity,
		})
	}
	s.Logger.Debug("inventory mapped")

	var receivedTransactions []domain.ReceivedTransaction
	var sentTransactions []domain.SentTransaction
	for _, transaction := range userTransactions {
		if transaction.FromUser == username {
			sentTransactions = append(sentTransactions, domain.SentTransaction{
				ToUser: transaction.ToUser,
				Amount: transaction.Amount,
			})
		} else {
			receivedTransactions = append(receivedTransactions, domain.ReceivedTransaction{
				FromUser: transaction.FromUser,
				Amount:   transaction.Amount,
			})
		}
	}
	s.Logger.Debug("transactions mapped")

	userDomain := domain.User{
		Coins:     userBalance,
		Inventory: userInventory,
		CoinHistory: domain.History{
			Received: receivedTransactions,
			Sent:     sentTransactions,
		},
	}

	dtoInventory := make([]dto.Item, len(userDomain.Inventory))
	for idx := range userDomain.Inventory {
		dtoInventory[idx] = dto.Item{
			ObjType:  userDomain.Inventory[idx].ObjType,
			Quantity: userDomain.Inventory[idx].Quantity,
		}
	}
	dtoReceived := make([]dto.ReceivedTransaction, len(userDomain.CoinHistory.Received))
	for idx := range userDomain.CoinHistory.Received {
		dtoReceived[idx] = dto.ReceivedTransaction{
			FromUser: userDomain.CoinHistory.Received[idx].FromUser,
			Amount:   userDomain.CoinHistory.Received[idx].Amount,
		}
	}
	dtoSent := make([]dto.SentTransaction, len(userDomain.CoinHistory.Sent))
	for idx := range userDomain.CoinHistory.Sent {
		dtoSent[idx] = dto.SentTransaction{
			ToUser: userDomain.CoinHistory.Sent[idx].ToUser,
			Amount: userDomain.CoinHistory.Sent[idx].Amount,
		}
	}

	dtoUser := dto.InfoResponse{
		Coins:     userDomain.Coins,
		Inventory: dtoInventory,
		CoinHistory: dto.History{
			Received: dtoReceived,
			Sent:     dtoSent,
		},
	}

	s.Logger.Debug("preparing response")

	return &dtoUser, nil
}

func (s api) SendCoins(ctx context.Context, fromUser string, toUser dto.SendCoinRequest) error {
	if toUser.Amount <= 0 {
		s.Logger.Warn(
			"attempt to send unnatural amount of coins",
			domain.ErrBadRequest)
		return domain.ErrBadRequest
	}

	transaction := domain.SentTransaction{
		ToUser: toUser.ToUser,
		Amount: toUser.Amount,
	}
	if err := s.Storage.SendCoins(
		ctx,
		fromUser,
		transaction,
	); err != nil {
		return err
	}

	return nil
}

func (s api) BuyItem(ctx context.Context, itemID int, user string) error {
	return s.Storage.BuyItem(ctx, itemID, user)
}
