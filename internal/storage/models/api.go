package models

type Item struct {
	Id    int
	Name  string
	Price int
}

type Transaction struct {
	Id         int
	FromUserID int
	ToUserId   int
	Amount     int
}

type User struct {
	Id      int
	Name    string
	Balance int
}

type UserInventory struct {
	Id       int
	UserId   int
	ItemID   int
	Quantity int
}
