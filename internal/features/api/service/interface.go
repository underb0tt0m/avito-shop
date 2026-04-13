package service

import (
	"avito-shop/internal/features/api/transport/dto"
	"context"
)

type Service interface {
	GetUserInfo(ctx context.Context, username string) (*dto.InfoResponse, error)
}
