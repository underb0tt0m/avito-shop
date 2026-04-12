package models

type Transaction struct {
	Id         int
	FromUserID int
	ToUserId   int
	Amount     int
}
