package mocks

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"
)

type mockStorageAPI struct{}

func NewStorageAPI() storage.API {
	return mockStorageAPI{}
}

func (s mockStorageAPI) GetUserInfo(ctx context.Context, username string) ([]views.UserInventory, []views.UserTransaction, error) {
	inventory := []views.UserInventory{{0, "test", 0}}
	transactions := []views.UserTransaction{{"test", "test", 0}}
	return inventory, transactions, nil
}

func (s mockStorageAPI) SendCoins(ctx context.Context, fromUser string, transaction domain.SentTransaction) error {
	return nil
}

func (s mockStorageAPI) BuyItem(ctx context.Context, itemID int, user string) error {
	return nil
}
