package postgres

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"avito-shop/internal/storage/views"
	"avito-shop/internal/tools"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testLogger logging.Logger
	testPool   *pgxpool.Pool
	testCtx    context.Context
	testUsers  []struct {
		ID             int
		Name           string
		Balance        int
		HashedPassword string
	}
	testItems []struct {
		ID    int
		Name  string
		Price int
	}
	testInventories []struct {
		ID       int
		UserID   int
		ItemID   int
		Quantity int
	}
	testTransations []struct {
		FromUserID int
		ToUserID   int
		Amount     int
	}
)

func TestMain(m *testing.M) {
	if err := config.Init("../../../cmd/config.yaml"); err != nil {
		panic(errors.Join(
			errors.New("failed to load config: "),
			err))
	}
	testCtx = context.Background()

	testNetwork, err := network.New(testCtx)
	dbNetworkAlias := "db"
	if err != nil {
		panic(errors.Join(
			errors.New("failed to create network: "),
			err))
	}
	defer testNetwork.Remove(testCtx)

	dbImg := config.App.Storage.Type + ":" + config.App.Storage.Version
	dbEnvs := map[string]string{
		"POSTGRES_USER":     "postgres",
		"POSTGRES_PASSWORD": "postgres",
		"POSTGRES_DB":       "avito_shop",
	}

	postgresReq := testcontainers.ContainerRequest{
		Image:    dbImg,
		Networks: []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {dbNetworkAlias},
		},
		Env:          dbEnvs,
		ExposedPorts: []string{config.App.Storage.Connection.Port},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(config.App.Storage.Connection.Port),
		),
	}
	postgresC, err := testcontainers.GenericContainer(testCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		panic(errors.Join(
			errors.New("failed to run db container: "),
			err))
	}

	_, err = testcontainers.GenericContainer(
		testCtx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Networks: []string{testNetwork.Name},
				Image:    "migrate/migrate:v4.19.1",
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      "./migrations",
						ContainerFilePath: "/migrations",
						FileMode:          0755,
					},
				},
				Cmd: []string{
					"-path",
					"/migrations",
					"-database",
					fmt.Sprintf(
						"postgresql://postgres:postgres@%s:%s/avito_shop?sslmode=disable",
						dbNetworkAlias,
						config.App.Storage.Connection.Port,
					),
					"up",
				},
				WaitingFor: wait.ForAll(
					wait.ForLog("1/u init_schema"),
				),
			},
			Started: true,
		},
	)
	if err != nil {
		panic(errors.Join(
			errors.New("failed to run migrations container: "),
			err))
	}

	testUsers = []struct {
		ID             int
		Name           string
		Balance        int
		HashedPassword string
	}{
		{
			ID:             1,
			Name:           "Timur",
			Balance:        100,
			HashedPassword: "Timur",
		},
		{
			ID:             2,
			Name:           "Artem",
			Balance:        100,
			HashedPassword: "Artem",
		},
		{
			ID:             3,
			Name:           "Ruslan",
			Balance:        100,
			HashedPassword: "Ruslan",
		},
		{
			ID:             4,
			Name:           "Kate",
			Balance:        100,
			HashedPassword: "Kate",
		},
		{
			ID:             5,
			Name:           "Vlad",
			Balance:        0,
			HashedPassword: "Vlad",
		},
	}
	testItems = []struct {
		ID    int
		Name  string
		Price int
	}{
		{
			ID:    1,
			Name:  "Book",
			Price: 10,
		},
		{
			ID:    2,
			Name:  "Pencil",
			Price: 10,
		},
		{
			ID:    3,
			Name:  "notebook",
			Price: 10,
		},
		{
			ID:    4,
			Name:  "hamster",
			Price: 10,
		},
		{
			ID:    5,
			Name:  "pineapple",
			Price: 10,
		},
	}

	testTransations = []struct {
		FromUserID int
		ToUserID   int
		Amount     int
	}{
		{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     10,
		},

		{
			FromUserID: 2,
			ToUserID:   1,
			Amount:     10,
		},

		{
			FromUserID: 1,
			ToUserID:   3,
			Amount:     50,
		},

		{
			FromUserID: 2,
			ToUserID:   1,
			Amount:     15,
		},
	}
	testInventories = []struct {
		ID       int
		UserID   int
		ItemID   int
		Quantity int
	}{
		{
			ID:       1,
			UserID:   1,
			ItemID:   1,
			Quantity: 1,
		},

		{
			ID:       2,
			UserID:   1,
			ItemID:   2,
			Quantity: 3,
		},

		{
			ID:       3,
			UserID:   2,
			ItemID:   1,
			Quantity: 1,
		},

		{
			ID:       4,
			UserID:   2,
			ItemID:   2,
			Quantity: 2,
		},

		{
			ID:       5,
			UserID:   3,
			ItemID:   3,
			Quantity: 8,
		},
	}

	mappedPort, _ := postgresC.MappedPort(testCtx, config.App.Storage.Connection.Port)
	config.App.Storage.Connection.Port = mappedPort.Port()
	config.App.Storage.Connection.User = "postgres"
	config.App.Storage.Connection.Password = "postgres"
	testPool, err = tools.CreatePool(testCtx)
	if err != nil {
		panic(errors.Join(
			errors.New("failed to create pool: "),
			err))
	}

	testLogger = mocks.NewLogger(nil)

	code := m.Run()
	os.Exit(code)
}

func TestGetUserInfo(t *testing.T) {
	tests := []struct {
		name                 string
		username             string
		expectedBalance      int
		expectedInventory    []views.UserInventory
		expectedTransactions []views.UserTransaction
		wantErr              bool
		wantSpecificErr      error
	}{
		{
			"succesful",
			"Timur",
			100,
			[]views.UserInventory{
				{
					ItemName: "Book",
					Quantity: 1,
				},
				{
					ItemName: "Pencil",
					Quantity: 3,
				},
			},
			[]views.UserTransaction{
				{
					FromUser: "Timur",
					ToUser:   "Artem",
					Amount:   10,
				},
				{
					FromUser: "Artem",
					ToUser:   "Timur",
					Amount:   10,
				},
				{
					FromUser: "Timur",
					ToUser:   "Ruslan",
					Amount:   50,
				},
				{
					FromUser: "Artem",
					ToUser:   "Timur",
					Amount:   15,
				},
			},
			false,
			nil,
		},

		{
			"error_no_user",
			"NoName",
			0,
			nil,
			nil,
			true,
			pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := truncateTables(); err != nil {
				t.Fatalf("failed to truncate tables: %v", err)
			}

			testStorage := NewStorageAPI(testPool, testLogger)
			queryCtx, cancel := context.WithTimeout(testCtx, config.App.Storage.QueryTimeout)
			defer cancel()
			userBalance, userInventories, userTransactions, err := testStorage.GetUserInfo(queryCtx, tt.username)

			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected error %+v", err)
				}
				if tt.wantSpecificErr != nil && !errors.Is(err, tt.wantSpecificErr) {
					t.Fatalf("unexpected error %+v, want %+v", err, tt.wantSpecificErr)
				}
			}

			if tt.wantErr == true {
				if err == nil {
					t.Fatalf("unhandled error %+v", tt.wantErr)
				}
			}

			if userBalance != tt.expectedBalance {
				t.Fatalf("unexpected balance %+v, want %+v", userBalance, tt.expectedBalance)
			}
			if !reflect.DeepEqual(userInventories, tt.expectedInventory) {
				t.Fatalf("unexpected inventory %+v, want %+v", userInventories, tt.expectedInventory)
			}
			if !reflect.DeepEqual(userTransactions, tt.expectedTransactions) {
				t.Fatalf("unexpected transactions %+v, want %+v", userTransactions, tt.expectedTransactions)
			}
		})
	}
}

func TestSendCoins(t *testing.T) {
	tests := []struct {
		name        string
		fromUser    string
		transaction domain.SentTransaction
		err         error
	}{
		{
			"succesful",
			"Timur",
			domain.SentTransaction{
				ToUser: "Artem",
				Amount: 10,
			},
			nil,
		},

		{
			"error_sender_not_found",
			"NoName",
			domain.SentTransaction{
				ToUser: "Artem",
				Amount: 10,
			},
			domain.ErrNotFound,
		},

		{
			"error_insufficient_balance",
			"Timur",
			domain.SentTransaction{
				ToUser: "Artem",
				Amount: 1000,
			},
			domain.ErrInsufficientFunds,
		},

		{
			"error_recipient_not_found",
			"Timur",
			domain.SentTransaction{
				ToUser: "NoName",
				Amount: 10,
			},
			domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := truncateTables(); err != nil {
				t.Fatalf("failed to truncate tables: %v", err)
			}

			beforeFromBalance, checkBalanceErr := checkBalance(tt.fromUser)
			beforeToBalance, checkBalanceErr := checkBalance(tt.transaction.ToUser)

			testStorage := NewStorageAPI(testPool, testLogger)
			queryCtx, cancel := context.WithTimeout(testCtx, config.App.Storage.QueryTimeout)
			defer cancel()
			err := testStorage.SendCoins(queryCtx, tt.fromUser, tt.transaction)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("unexpected error %+v", err)
				}
				if !errors.Is(err, tt.err) {
					t.Fatalf("unexpected error %+v, want %+v", err, tt.err)
				}
			}

			if tt.err != nil {
				if err == nil {
					t.Fatalf("unhandled error %+v", tt.err)
				}
			}

			if checkBalanceErr == nil && tt.err == nil {
				afterFromBalance, _ := checkBalance(tt.fromUser)
				afterToBalance, _ := checkBalance(tt.transaction.ToUser)
				if afterFromBalance != beforeFromBalance-tt.transaction.Amount {
					t.Fatalf("expected balance %d, got %d", beforeFromBalance-tt.transaction.Amount, afterFromBalance)
				}
				if afterToBalance != beforeToBalance+tt.transaction.Amount {
					t.Fatalf("expected balance %d, got %d", beforeToBalance+tt.transaction.Amount, afterToBalance)
				}
			}
		})
	}
}

func TestBuyItem(t *testing.T) {
	tests := []struct {
		name   string
		itemID int
		user   string
		err    error
	}{
		{
			"succesful",
			1,
			"Timur",
			nil,
		},

		{
			"error_item_not_found",
			1000,
			"Timur",
			domain.ErrNotFound,
		},

		{
			"error_insufficient_balance",
			1,
			"Vlad",
			domain.ErrInsufficientFunds,
		},

		{
			"user_not_found",
			1,
			"Noname",
			domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := truncateTables(); err != nil {
				t.Fatalf("failed to truncate tables: %v", err)
			}

			testStorage := NewStorageAPI(testPool, testLogger)
			queryCtx, cancel := context.WithTimeout(testCtx, config.App.Storage.QueryTimeout)
			defer cancel()
			err := testStorage.BuyItem(queryCtx, tt.itemID, tt.user)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("unexpected error %+v", err)
				}
				if !errors.Is(err, tt.err) {
					t.Fatalf("unexpected error %+v, want %+v", err, tt.err)
				}
			}

			if tt.err != nil {
				if err == nil {
					t.Fatalf("unhandled error %+v", tt.err)
				}
			}
		})
	}
}

func truncateTables() error {
	tx, err := testPool.Begin(testCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(testCtx)

	if _, err = tx.Exec(
		testCtx,
		`
TRUNCATE TABLE users, items, transactions, user_inventories CASCADE;
`); err != nil {
		return err
	}

	for _, user := range testUsers {
		var hash bytes.Buffer
		if err = json.NewEncoder(&hash).Encode(user.HashedPassword); err != nil {
			return errors.Join(
				errors.New("failed to encode user test password: "),
				err)
		}
		if _, err = tx.Exec(
			testCtx,
			`
INSERT INTO users (id, name, balance, password_hash) 
VALUES ($1, $2, $3, $4);
`, user.ID, user.Name, user.Balance, user.HashedPassword); err != nil {
			return errors.Join(
				errors.New("failed to insert test user: "),
				err)
		}
	}

	for _, item := range testItems {
		if _, err = tx.Exec(
			testCtx,
			`
INSERT INTO items (id, name, price) 
VALUES ($1, $2, $3);
`, item.ID, item.Name, item.Price); err != nil {
			return errors.Join(
				errors.New("failed to insert test item: "),
				err)
		}
	}

	for _, transaction := range testTransations {
		if _, err = tx.Exec(
			testCtx,
			`
INSERT INTO transactions (from_user_id, to_user_id, amount) 
VALUES ($1, $2, $3);
`, transaction.FromUserID, transaction.ToUserID, transaction.Amount); err != nil {
			return errors.Join(
				errors.New("failed to insert test transaction: "),
				err)
		}
	}

	for _, inventory := range testInventories {
		if _, err = tx.Exec(
			testCtx,
			`
INSERT INTO user_inventories (id, user_id, item_id, quantity) 
VALUES ($1, $2, $3, $4);
`, inventory.ID, inventory.UserID, inventory.ItemID, inventory.Quantity); err != nil {
			return errors.Join(
				errors.New("failed to insert test inventory: "),
				err)
		}
	}

	return tx.Commit(testCtx)
}

func checkBalance(user string) (int, error) {
	var balance int
	if err := testPool.QueryRow(
		testCtx,
		`
SELECT balance 
FROM users 
WHERE name = $1;
`,
		user).Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}
