package service

import (
	"avito-shop/internal/core/domains"
	"avito-shop/internal/features/api/transport/dto"
)

type ServiceImpl struct{}

func (s ServiceImpl) GetUserInfo(username string) (*dto.InfoResponse, error) {
	// 1. Проверить и преобразовать токен в имя пользователя
	// 2. Вызвать метод репозитория для получения данных пользователя
	// 3. Передать информацию обработчику
	user := domains.User{}
	dtoUser := dto.InfoResponse{
		Coins:       user.Coins,
		Inventory:   user.Inventory,
		CoinHistory: user.CoinHistory,
	}
	//TODO реализовать маппер
	return &dtoUser, nil
}
