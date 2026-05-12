package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"iam/internal/dto"
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
	token := ""
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else if cookie, err := c.Cookie("iam_access_token"); err == nil {
		token = cookie
	}
	if token == "" {
		return 0, false
	}

	claims, err := h.auth.ParseToken(token)
	if err != nil {
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
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	redirect := fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, c.Request.URL.RequestURI())
	return loginURL + "?redirect=" + url.QueryEscape(redirect)
}

func (h *OAuthHandler) Token(c *gin.Context) {
	var req dto.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": err.Error()})
		return
	}
	data, err := h.service.Token(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *OAuthHandler) UserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		resp.Fail(c, 401, "missing bearer token")
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	data, err := h.service.UserInfo(c.Request.Context(), token)
	if err != nil {
		resp.Fail(c, 401, err.Error())
		return
	}
	resp.OK(c, data)
}
