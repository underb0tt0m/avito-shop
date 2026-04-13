package repository

import (
	"avito-shop/internal/features/api/repository/views"
	"context"
)

type Storage interface {
	GetUserInfo(ctx context.Context, username string) ([]views.UserInventory, []views.UserTransaction, error)
}
