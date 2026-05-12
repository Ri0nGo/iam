package dto

type CreateRoleRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

type BindUserRolesRequest struct {
	RoleCodes []string `json:"role_codes" binding:"required"`
}
