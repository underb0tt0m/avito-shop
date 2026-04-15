package storage

import (
	"avito-shop/internal/storage/views"
	"context"
)

type API interface {
	GetUserInfo(ctx context.Context, username string) ([]views.UserInventory, []views.UserTransaction, error)
}
