package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"context"
)

type API interface {
	GetUserInfo(ctx context.Context, username string) (*dto.InfoResponse, error)
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
	userInventories, userTransactions, err := s.Storage.GetUserInfo(ctx, username)
	if err != nil {
		s.Logger.Error(
			"failed to get user info from mainRoutRepository",
			err,
		)
		return nil, err
	}
	s.Logger.Debug("received data from mainRoutRepository")

	userBalance := userInventories[0].Balance
	s.Logger.Debug("user balance extracted")

	var userInventory []domain.Item
	for _, item := range userInventories {
		if item.ItemName == "" {
			break
		}
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
