package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"avito-shop/internal/features/api/mainRoutRepository/mainRootViews"
)

type ShopStorage struct {
	Conn   *pgx.Conn
	Logger *zap.Logger
}

var ErrNotFound = fmt.Errorf("not found")

// вернуть домен
func (s ShopStorage) GetUserInfo(ctx context.Context, username string) ([]mainRootViews.UserInventory, []mainRootViews.UserTransaction, error) {
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

	// https://github.com/masterminds/squirrel

	rows, err := s.Conn.Query(ctx, userInfoStmt, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		//s.Logger.Error(
		//	"failed to query user inventory",
		//	zap.Error(err),
		//	zap.String("username", username),
		//)
		return nil, nil, fmt.Errorf("failed to query user inventory: %w", err)
	}

	var (
		balance, quantity *int
		itemName          *string
		userInventories   []mainRootViews.UserInventory
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
				zap.Error(err),
				zap.String("username", username),
			)
			return nil, nil, err
		}
		if itemName != nil {
			userInventories = append(userInventories, mainRootViews.UserInventory{
				Balance:  *balance,
				ItemName: *itemName,
				Quantity: *quantity,
			})
		} else {
			userInventories = append(userInventories, mainRootViews.UserInventory{
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
		userTransactions []mainRootViews.UserTransaction
	)

	if err != nil {
		s.Logger.Error(
			"failed to query user transactions",
			zap.Error(err),
			zap.String("username", username),
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
				zap.Error(err),
				zap.String("username", username),
			)
			return nil, nil, err
		}
		userTransactions = append(userTransactions, mainRootViews.UserTransaction{
			FromUser: fromUser,
			ToUser:   toUser,
			Amount:   amount,
		})
	}
	return userInventories, userTransactions, nil
}
