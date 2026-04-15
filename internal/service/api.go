package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/storage"
	"context"

	"go.uber.org/zap"
)

type API interface {
	GetUserInfo(ctx context.Context, username string) (*dto.InfoResponse, error)
}
type api struct {
	Storage storage.API
	Logger  *zap.Logger
}

func NewApi(s storage.API, l *zap.Logger) API {
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
			zap.Error(err),
			zap.String("username", username),
		)
		return nil, err
	}
	s.Logger.Debug("received data from mainRoutRepository",
		zap.Int("inventory_count", len(userInventories)),
		zap.Int("transactions_count", len(userTransactions)),
		zap.String("username", username),
	)

	userBalance := userInventories[0].Balance
	s.Logger.Debug("user balance extracted",
		zap.Int("balance", userBalance),
		zap.String("username", username),
	)

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
	s.Logger.Debug("inventory mapped",
		zap.Int("items_count", len(userInventory)),
		zap.String("username", username),
	)

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
	s.Logger.Debug("transactions mapped",
		zap.Int("received_count", len(receivedTransactions)),
		zap.Int("sent_count", len(sentTransactions)),
		zap.String("username", username),
	)

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

	s.Logger.Debug("preparing response",
		zap.Int("coins", dtoUser.Coins),
		zap.Int("inventory_items", len(dtoUser.Inventory)),
		zap.Int("received_transactions", len(dtoUser.CoinHistory.Received)),
		zap.Int("sent_transactions", len(dtoUser.CoinHistory.Sent)),
		zap.String("username", username),
	)

	return &dtoUser, nil
}
