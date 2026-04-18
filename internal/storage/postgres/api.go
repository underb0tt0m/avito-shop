package postgres

import (
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"

	"github.com/jackc/pgx/v5"
)

type storageAPI struct {
	Conn   *pgx.Conn
	Logger logging.Logger
}

func NewStorageAPI(conn *pgx.Conn, logger logging.Logger) storage.API {
	return storageAPI{
		Conn:   conn,
		Logger: logger,
	}
}

func (s storageAPI) GetUserInfo(ctx context.Context, username string) ([]views.UserInventory, []views.UserTransaction, error) {
	userInfoStmt := `
SELECT
	a.balance,
	c.name AS item_name,
	b.quantity
FROM users a
LEFT JOIN user_inventories b ON a.id=b.user_id
LEFT JOIN items c ON c.id=b.item_id
WHERE a.name=$1
;`
	rows, err := s.Conn.Query(ctx, userInfoStmt, username)
	if err != nil {
		s.Logger.Error(
			"failed to query user inventory",
			err,
		)
		return nil, nil, err
	}

	var (
		balance, quantity *int
		itemName          *string
		userInventories   []views.UserInventory
	)
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&balance,
			&itemName,
			&quantity,
		); err != nil {
			s.Logger.Error(
				"failed to scan user inventory row",
				err,
			)
			return nil, nil, err
		}
		if itemName != nil {
			userInventories = append(userInventories, views.UserInventory{
				Balance:  *balance,
				ItemName: *itemName,
				Quantity: *quantity,
			})
		} else {
			userInventories = append(userInventories, views.UserInventory{
				Balance: *balance,
			})
		}
	}

	userTransactionsInfoStmt := `
SELECT
	b.name AS from_username,
	c.name AS to_username,
	a.amount
FROM
	transactions a
JOIN users b ON a.from_user_id=b.id
JOIN users c ON a.to_user_id=c.id
WHERE b.name=$1 OR c.name=$1
;`
	rows, err = s.Conn.Query(ctx, userTransactionsInfoStmt, username)

	var (
		fromUser         string
		toUser           string
		amount           int
		userTransactions []views.UserTransaction
	)

	if err != nil {
		s.Logger.Error(
			"failed to query user transactions",
			err,
		)
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(
			&fromUser,
			&toUser,
			&amount,
		); err != nil {
			s.Logger.Error(
				"failed to scan user transaction row",
				err,
			)
			return nil, nil, err
		}
		userTransactions = append(userTransactions, views.UserTransaction{
			FromUser: fromUser,
			ToUser:   toUser,
			Amount:   amount,
		})
	}
	return userInventories, userTransactions, nil
}
