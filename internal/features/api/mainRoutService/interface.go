package mainRoutService

import (
	"avito-shop/internal/features/api/mainRoutTransport/mainRoutDTO"
	"context"
)

type Service interface {
	GetUserInfo(ctx context.Context, username string) (*mainRoutDTO.InfoResponse, error)
}
