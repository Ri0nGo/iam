package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"iam/internal/dto"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type OAuthHandler struct {
	service  service.OAuthService
	auth     service.AuthService
	redis    *redis.Client
	loginURL string
}

func NewOAuthHandler(oauthService service.OAuthService, authService service.AuthService, redisClient *redis.Client, loginURL string) *OAuthHandler {
	return &OAuthHandler{service: oauthService, auth: authService, redis: redisClient, loginURL: loginURL}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	var query dto.AuthorizeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	userID, ok := h.currentUserID(c)
	if !ok {
		c.Redirect(http.StatusFound, h.loginRedirectURL(c))
		return
	}

	data, err := h.service.Authorize(c.Request.Context(), query, userID)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	c.Redirect(http.StatusFound, data.RedirectTo)
}

func (h *OAuthHandler) currentUserID(c *gin.Context) (uint64, bool) {
	token, err := c.Cookie("iam_access_token")
	if err != nil {
		token = ""
	}
	if token == "" {
		return 0, false
	}

	claims, err := h.auth.ParseToken(token)
	if err != nil {
		return 0, false
	}
	if claims.TokenUse != jwtpkg.TokenUseConsole {
		return 0, false
	}
	if ok, _ := h.redis.Exists(c.Request.Context(), "iam:token:blacklist:"+claims.ID).Result(); ok > 0 {
		return 0, false
	}
	userID, err := service.ParseUserID(claims.Subject)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func (h *OAuthHandler) loginRedirectURL(c *gin.Context) string {
	loginURL := h.loginURL
	if loginURL == "" {
		loginURL = "/login"
	}
	redirect := h.authorizeRedirectURL(c, loginURL)
	return loginURL + "?redirect=" + url.QueryEscape(redirect)
}

func (h *OAuthHandler) authorizeRedirectURL(c *gin.Context, loginURL string) string {
	if u, err := url.Parse(loginURL); err == nil && u.IsAbs() {
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, c.Request.URL.RequestURI())
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, c.Request.URL.RequestURI())
}

func (h *OAuthHandler) Token(c *gin.Context) {
	var req dto.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.Token(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.OK(c, data)
}

func (h *OAuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.OK(c, data)
}

func (h *OAuthHandler) CheckToken(c *gin.Context) {
	var query dto.CheckTokenQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.CheckToken(c.Request.Context(), query.AccessToken, query.OpenID); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.OK(c, gin.H{"success": true})
}

func (h *OAuthHandler) UserInfo(c *gin.Context) {
	token := strings.TrimSpace(c.Query("access_token"))
	if token == "" {
		resp.Fail(c, 401, "missing access_token")
		return
	}
	data, err := h.service.UserInfo(c.Request.Context(), token, c.Query("openid"))
	if err != nil {
		resp.Fail(c, 401, err.Error())
		return
	}
	resp.OK(c, data)
}
