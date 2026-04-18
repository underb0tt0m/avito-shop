package init

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateMainTables(ctx context.Context, conn *pgx.Conn) {
	stmt := `
CREATE TABLE users(
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(20) NOT NULL UNIQUE,
	password_hash BYTEA NOT NULL,
	balance INT NOT NULL CHECK ( balance >= 0 ) DEFAULT 1000
);

CREATE TABLE items(
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    price INT NOT NULL CHECK ( price > 0 )
);

CREATE TABLE user_inventories(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK ( quantity > 0 ) DEFAULT 1,
    UNIQUE (user_id, item_id)
);

CREATE TABLE transactions(
    id BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE NO ACTION,
    to_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE NO ACTION,
    amount INT NOT NULL CHECK ( amount > 0 )
);

INSERT INTO items(name, price)
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500)
`
	conn.Exec(ctx, stmt)
}
