package authRepository

type Storage interface {
	GetHashedUserPassword(username string, password []byte) ([]byte, bool, error)
}
