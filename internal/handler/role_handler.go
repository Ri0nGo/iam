package handler

import (
	"strconv"

	"iam/internal/dto"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct{ service service.RoleService }

func NewRoleHandler(service service.RoleService) *RoleHandler { return &RoleHandler{service: service} }

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	role, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.service.List(c.Request.Context())
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, roles)
}

func (h *RoleHandler) BindUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	var req dto.BindUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	if err := h.service.BindUserRoles(c.Request.Context(), userID, req.RoleCodes); err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, gin.H{"success": true})
}

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, 400, "invalid id")
		return
	}
	roles, err := h.service.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c, 400, err.Error())
		return
	}
	resp.OK(c, roles)
}
