package views

type UserInventory struct {
	ItemName string
	Quantity int
}

type UserTransaction struct {
	FromUser string
	ToUser   string
	Amount   int
}
