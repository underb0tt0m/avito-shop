package dto

// DTO запроса на POST /api/login
type UserData struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}
