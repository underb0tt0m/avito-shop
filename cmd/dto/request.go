package dto

// AuthRequest - DTO запроса на POST /api/login
type AuthRequest struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

// SendCoinRequest - DTO запроса на POST /api/sendCoin
type SendCoinRequest struct {
	ToUser string `json:"toUser"` // Имя пользователя, которому нужно отправить монеты.
	Amount int    `json:"amount"` // Количество монет, которые необходимо отправить.
}
