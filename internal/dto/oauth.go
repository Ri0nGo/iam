package dto

type AuthorizeQuery struct {
	ResponseType string `form:"response_type" binding:"required"`
	ClientID     string `form:"client_id" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	Scope        string `form:"scope"`
	State        string `form:"state"`
}

type AuthorizeResponse struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	RedirectTo string `json:"redirect_to"`
	ExpiresIn  int    `json:"expires_in"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type" binding:"required"`
	ClientID     string `json:"client_id" form:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" form:"client_secret" binding:"required"`
	Code         string `json:"code" form:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

type OAuthCodePayload struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	Scope       string `json:"scope"`
	State       string `json:"state"`
}
