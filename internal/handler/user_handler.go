package handler

import (
	"strconv"

	"iam/internal/dto"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{ service service.UserService }

func NewUserHandler(service service.UserService) *UserHandler { return &UserHandler{service: service} }

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	user, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, user)
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c, 404, err.Error())
		return
	}
	resp.OK(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	users, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, users)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	if err := h.service.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, gin.H{"success": true})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, gin.H{"success": true})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), id, req.Password); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, gin.H{"success": true})
}
