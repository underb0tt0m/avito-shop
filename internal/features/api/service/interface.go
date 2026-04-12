package service

import (
	"avito-shop/internal/features/api/transport/dto"
)

type Service interface {
	GetUserInfo(username string) (*dto.InfoResponse, error)
}
