package storage

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/storage/views"
	"context"
)

type API interface {
	GetUserInfo(ctx context.Context, username string) ([]views.UserInventory, []views.UserTransaction, error)
	SendCoins(ctx context.Context, fromUser string, transaction domain.SentTransaction) error
}
