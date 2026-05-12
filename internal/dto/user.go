package dto

type CreateUserRequest struct {
	Username    string   `json:"username" binding:"required"`
	Password    string   `json:"password" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Email       string   `json:"email"`
	Mobile      string   `json:"mobile"`
	Status      int      `json:"status" binding:"omitempty,oneof=1 2"`
	Remark      string   `json:"remark"`
	RoleCodes   []string `json:"role_codes"`
}

type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type UserListQuery struct {
	Keyword string `form:"keyword"`
	Status  int    `form:"status"`
}
