package dto

// Response - DTO ответа на GET /api/info
type Response any

// InfoResponse - DTO успешного ответа на GET /api/info
type InfoResponse struct {
	Coins       int     `json:"coins"` // Количество доступных монет
	Inventory   []Item  `json:"inventory"`
	CoinHistory History `json:"coinHistory"`
}

// Item представляет предмет в инвентаре пользователя.
type Item struct {
	ObjType  string `json:"type"`     // Тип предмета
	Quantity int    `json:"quantity"` // Количество предметов
}

// History представляет историю транзакций пользователя.
type History struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

// ReceivedTransaction представляет транзакцию получения монет от другого пользователя.
type ReceivedTransaction struct {
	FromUser string `json:"fromUser"` // Имя пользователя, который отправил монеты
	Amount   int    `json:"amount"`   // Количество полученных монет
}

// SentTransaction представляет транзакцию отправки монет другому пользователю.
type SentTransaction struct {
	ToUser string `json:"toUser"` // Имя пользователя, которому отправлены монеты
	Amount int    `json:"amount"` // Количество полученных монет
}

// AuthResponse - DTO успешного ответа на POST /api/login
type AuthResponse struct {
	Token string `json:"token"`
}

// ErrorResponse - DTO ответа с ошибкой на все запросы
type ErrorResponse struct {
	Errors string `json:"errors"` // Сообщение об ошибке, описывающее проблему
}
