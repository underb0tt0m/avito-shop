package dto

// DTO успешного ответа на GET /api/info
type Response interface{}

type InfoResponse struct {
	Coins       int     `json:"coins"` // Количество доступных монет
	Inventory   []Item  `json:"inventory"`
	CoinHistory History `json:"coinHistory"`
}

type Item struct {
	ObjType  string `json:"type"`     // Тип предмета
	Quantity int    `json:"quantity"` // Количество предметов
}

type History struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

type ReceivedTransaction struct {
	FromUser string `json:"fromUser"` // Имя пользователя, который отправил монеты
	Amount   int    `json:"amount"`   // Количество полученных монет
}

type SentTransaction struct {
	ToUser string `json:"toUser"` // Имя пользователя, которому отправлены монеты
	Amount int    `json:"amount"` // Количество полученных монет
}

// DTO успешного ответа на POST /api/login
type AuthResponse struct {
	Token string `json:"token"`
}

// DTO ответа с ошибкой на все запросы
type ErrorResponse struct {
	Errors string `json:"errors"` // Сообщение об ошибке, описывающее проблему
}
