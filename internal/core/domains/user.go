package domains

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
