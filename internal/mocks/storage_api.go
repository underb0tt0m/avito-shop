package mocks

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"
)

type mockStorageAPI struct {
	GetUserInfoFunc func(ctx context.Context, username string) (int, []views.UserInventory, []views.UserTransaction, error)
	SendCoinsFunc   func(ctx context.Context, fromUser string, transaction domain.SentTransaction) error
	BuyItemFunc     func(ctx context.Context, itemID int, user string) error
}

func NewStorageAPI(
	GetUserInfoFunc func(ctx context.Context, username string) (int, []views.UserInventory, []views.UserTransaction, error),
	SendCoinsFunc func(ctx context.Context, fromUser string, transaction domain.SentTransaction) error,
	BuyItemFunc func(ctx context.Context, itemID int, user string) error,
) storage.API {
	return mockStorageAPI{
		GetUserInfoFunc: GetUserInfoFunc,
		SendCoinsFunc:   SendCoinsFunc,
		BuyItemFunc:     BuyItemFunc,
	}
}

func (s mockStorageAPI) GetUserInfo(ctx context.Context, username string) (int, []views.UserInventory, []views.UserTransaction, error) {
	if s.GetUserInfoFunc != nil {
		return s.GetUserInfoFunc(ctx, username)
	}
	return 0, nil, nil, nil
}

func (s mockStorageAPI) SendCoins(ctx context.Context, fromUser string, transaction domain.SentTransaction) error {
	if s.SendCoinsFunc != nil {
		return s.SendCoinsFunc(ctx, fromUser, transaction)
	}
	return nil
}

func (s mockStorageAPI) BuyItem(ctx context.Context, itemID int, user string) error {
	if s.BuyItemFunc != nil {
		return s.BuyItemFunc(ctx, itemID, user)
	}
	return nil
}
