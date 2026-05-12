package handler

import (
	"strings"

	"iam/internal/dto"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{ service service.AuthService }

func NewAuthHandler(service service.AuthService) *AuthHandler { return &AuthHandler{service: service} }

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	data, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	c.SetCookie("iam_access_token", data.AccessToken, int(data.ExpiresIn), "/", "", false, true)
	resp.OK(c, data)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token == "" {
		token, _ = c.Cookie("iam_access_token")
	}
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	c.SetCookie("iam_access_token", "", -1, "/", "", false, true)
	resp.OK(c, gin.H{"success": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint64("userID")
	data, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c, 404, err.Error())
		return
	}
	resp.OK(c, data)
}
