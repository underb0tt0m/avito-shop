package storage

import (
	"context"

	"avito-shop/internal/features/api/mainRoutRepository/mainRootViews"
)

type Shop interface {
	GetUserInfo(ctx context.Context, username string) ([]mainRootViews.UserInventory, []mainRootViews.UserTransaction, error)
}
