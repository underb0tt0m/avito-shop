package storage

import "context"

type Auth interface {
	GetHashedUserPassword(ctx context.Context, username string) ([]byte, bool, error)
}
