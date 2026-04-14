package mainRootViews

type UserInventory struct {
	Balance  int
	ItemName string
	Quantity int
}

type UserTransaction struct {
	FromUser string
	ToUser   string
	Amount   int
}
