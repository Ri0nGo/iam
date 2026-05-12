package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        CurrentUser `json:"user"`
}

type CurrentUser struct {
	ID          uint64   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Status      int      `json:"status"`
	Roles       []string `json:"roles"`
}
