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
	GrantType   string `json:"grant_type" form:"grant_type" binding:"required"`
	ClientID    string `json:"client_id" form:"client_id" binding:"required"`
	Secret      string `json:"secret" form:"secret" binding:"required"`
	Code        string `json:"code" form:"code" binding:"required"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope,omitempty"`
}

type RefreshTokenRequest struct {
	ClientID     string `json:"client_id" form:"client_id" binding:"required"`
	GrantType    string `json:"grant_type" form:"grant_type" binding:"required"`
	RefreshToken string `json:"refresh_token" form:"refresh_token" binding:"required"`
}

type CheckTokenQuery struct {
	AccessToken string `form:"access_token" binding:"required"`
	OpenID      string `form:"openid" binding:"required"`
}

type OAuthUserInfo struct {
	OpenID      string   `json:"openid"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Status      int      `json:"status"`
	Roles       []string `json:"roles"`
}

type OAuthCodePayload struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	Scope       string `json:"scope"`
	State       string `json:"state"`
}

type OAuthRefreshPayload struct {
	ClientID string `json:"client_id"`
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	OpenID   string `json:"openid"`
	Scope    string `json:"scope"`
}
