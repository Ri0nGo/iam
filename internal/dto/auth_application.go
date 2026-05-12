package dto

type CreateAuthApplicationRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	SecretKey    string `json:"secret_key" binding:"required"`
	ResponseType string `json:"response_type" binding:"required,oneof=code"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	Status       int    `json:"status" binding:"required,oneof=1 2"`
	Remark       string `json:"remark"`
}

type UpdateAuthApplicationRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	SecretKey    string `json:"secret_key" binding:"required"`
	ResponseType string `json:"response_type" binding:"required,oneof=code"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	Status       int    `json:"status" binding:"required,oneof=1 2"`
	Remark       string `json:"remark"`
}

type AuthApplicationListQuery struct {
	Keyword string `form:"keyword"`
	Status  int    `form:"status"`
}

type AuthApplicationResponse struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	SecretKey    string `json:"secret_key"`
	ResponseType string `json:"response_type"`
	RedirectURI  string `json:"redirect_uri"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
}
