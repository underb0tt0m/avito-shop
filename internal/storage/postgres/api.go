package postgres

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (s storageAPI) GetUserInfo(ctx context.Context, username string) (int, []views.UserInventory, []views.UserTransaction, error) {
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
		return 0, nil, nil, err
	}

	userBalance := -1
	var (
		balance, quantity *int
		itemName          *string
		userInventories   []views.UserInventory
	)
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(
			&balance,
			&itemName,
			&quantity,
		); err != nil {
			s.Logger.Error(
				"failed to scan user inventory row",
				err,
			)
			return 0, nil, nil, err
		}
		if userBalance == -1 {
			userBalance = *balance
		}
		if itemName != nil {
			userInventories = append(userInventories, views.UserInventory{
				ItemName: *itemName,
				Quantity: *quantity,
			})
		}
	}
	if userBalance == -1 {
		return 0, nil, nil, pgx.ErrNoRows
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
		return 0, nil, nil, err
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
			return 0, nil, nil, err
		}
		userTransactions = append(userTransactions, views.UserTransaction{
			FromUser: fromUser,
			ToUser:   toUser,
			Amount:   amount,
		})
	}
	return userBalance, userInventories, userTransactions, nil
}

func (s storageAPI) SendCoins(ctx context.Context, fromUser string, transaction domain.SentTransaction) error {
	updateStmt1 := `
UPDATE users
SET
  balance = balance - $1
WHERE
  NAME = $2;
`
	updateStmt2 := `
UPDATE users
SET
  balance = balance + $1
WHERE
  NAME = $2;
`
	insertStmt := `
INSERT INTO
	transactions (from_user_id, to_user_id, amount)
SELECT 
    u1.id, 
    u2.id, 
    $1
FROM
    users u1, users u2
WHERE u1.name = $2 AND u2.name = $3;
`
	tx, err := s.Conn.Begin(ctx)
	if err != nil {
		s.Logger.Error(
			"failed to begin transaction",
			err,
		)
		return err
	}
	defer tx.Rollback(ctx)

	cmd1, err := tx.Exec(
		ctx,
		updateStmt1,
		transaction.Amount,
		fromUser,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23514" {
				s.Logger.Warn(
					"attempt to transfer money with insufficient balance",
					err)
				return domain.ErrInsufficientFunds
			}
		}
		s.Logger.Error(
			"failed to update sender balance",
			err,
		)
		return domain.ErrInternalServerError
	}

	if cmd1.RowsAffected() == 0 {
		s.Logger.Error(
			"sender not found",
			domain.ErrNotFound,
		)
		return domain.ErrNotFound // хотя такого быть не должно из логики работы приложения
	}

	cmd2, err := tx.Exec(
		ctx,
		updateStmt2,
		transaction.Amount,
		transaction.ToUser,
	)
	if err != nil {
		s.Logger.Error(
			"failed to update recipient balance",
			err,
		)
		return domain.ErrInternalServerError
	}

	if cmd2.RowsAffected() == 0 {
		s.Logger.Error(
			"recipient not found",
			domain.ErrNotFound,
		)
		return domain.ErrNotFound
	}

	_, err = tx.Exec(
		ctx,
		insertStmt,
		transaction.Amount,
		fromUser,
		transaction.ToUser,
	)
	if err != nil {
		s.Logger.Error(
			"failed to insert transaction",
			err,
		)
		return domain.ErrInternalServerError
	}

	if err = tx.Commit(ctx); err != nil {
		s.Logger.Error(
			"failed to commit transaction",
			err,
		)
		return err
	}

	return nil
}

func (s storageAPI) BuyItem(ctx context.Context, itemID int, user string) error {
	tx, err := s.Conn.Begin(ctx)
	if err != nil {
		s.Logger.Error(
			"failed to begin transaction",
			err,
		)
		return err
	}
	defer tx.Rollback(ctx)

	selectStmt := `
SELECT price
FROM items
WHERE id = $1;
`
	updateStmt := `
UPDATE users
SET balance = balance - $1
WHERE name = $2
RETURNING id;
`
	insertStmt := `
INSERT INTO user_inventories (user_id, item_id, quantity)
VALUES ($1, $2, 1)
ON CONFLICT (user_id, item_id) DO UPDATE
SET quantity = user_inventories.quantity + 1;
`
	var itemPrice int
	if err = tx.QueryRow(
		ctx,
		selectStmt,
		itemID,
	).Scan(&itemPrice); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.Logger.Warn(
				"attempt to buy nonexistent item",
				err,
			)
			return domain.ErrNotFound
		}
		s.Logger.Error(
			"failed to get item cost",
			err,
		)
		return domain.ErrInternalServerError
	}

	var userID int
	if err = tx.QueryRow(
		ctx,
		updateStmt,
		itemPrice,
		user,
	).Scan(&userID); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23514" {
				s.Logger.Warn(
					"attempt to transfer money with insufficient balance",
					err)
				return domain.ErrInsufficientFunds
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			s.Logger.Warn(
				"user doesn't exist",
				err,
			)
			return domain.ErrNotFound
		}
		s.Logger.Error(
			"failed to update user balance",
			err,
		)
		return domain.ErrInternalServerError
	}

	if _, err = tx.Exec(ctx, insertStmt, userID, itemID); err != nil {
		s.Logger.Error(
			"failed to insert row in user_inventories",
			err,
		)
		return domain.ErrInternalServerError
	}

	if err = tx.Commit(ctx); err != nil {
		s.Logger.Error(
			"failed to commit transaction",
			err,
		)
		return domain.ErrInternalServerError
	}

	return nil
}
