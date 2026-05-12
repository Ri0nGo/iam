package handler

import (
	"strconv"

	"iam/internal/dto"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthApplicationHandler struct {
	service service.AuthApplicationService
}

func NewAuthApplicationHandler(service service.AuthApplicationService) *AuthApplicationHandler {
	return &AuthApplicationHandler{service: service}
}

func (h *AuthApplicationHandler) Create(c *gin.Context) {
	var req dto.CreateAuthApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	data, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, data)
}

func (h *AuthApplicationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	var req dto.UpdateAuthApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	data, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, data)
}

func (h *AuthApplicationHandler) Delete(c *gin.Context) {
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

func (h *AuthApplicationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	data, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c, 404, err.Error())
		return
	}
	resp.OK(c, data)
}

func (h *AuthApplicationHandler) List(c *gin.Context) {
	var query dto.AuthApplicationListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	data, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, data)
}
