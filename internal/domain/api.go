package domain

import "golang.org/x/crypto/bcrypt"

type User struct {
	Coins       int // Количество доступных монет
	Inventory   []Item
	CoinHistory History
}

type Item struct {
	ObjType  string // Тип предмета
	Quantity int    // Количество предметов
}

type History struct {
	Received []ReceivedTransaction
	Sent     []SentTransaction
}

type ReceivedTransaction struct {
	FromUser string // Имя пользователя, который отправил монеты
	Amount   int    // Количество полученных монет
}

type SentTransaction struct {
	ToUser string // Имя пользователя, которому отправлены монеты
	Amount int    // Количество полученных монет
}

type HashedUserData struct {
	Name     string
	Password []byte
}

func NewHashed(name string, password string) (HashedUserData, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return HashedUserData{}, err
	}

	return HashedUserData{
		Name:     name,
		Password: hashedPassword,
	}, nil
}
