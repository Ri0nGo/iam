package dto

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	ExpiresIn   int64       `json:"expires_in"`
	User        CurrentUser `json:"user"`
}

type CurrentUser struct {
	ID          uint64     `json:"id"`
	Username    string     `json:"username"`
	OpenID      *string    `json:"openid"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Mobile      *string    `json:"mobile"`
	Email       *string    `json:"email"`
	Status      int        `json:"status"`
	Remark      string     `json:"remark"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Roles       []string   `json:"roles"`
}
