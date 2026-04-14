package mainRoutRepository

import (
	"avito-shop/internal/features/api/mainRoutRepository/mainRootViews"
	"context"
)

type Storage interface {
	GetUserInfo(ctx context.Context, username string) ([]mainRootViews.UserInventory, []mainRootViews.UserTransaction, error)
}
