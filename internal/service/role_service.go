package service

import (
	"context"
	"fmt"

	"iam/internal/dto"
	"iam/internal/model"
	"iam/internal/repository"
)

type RoleService interface {
	Create(ctx context.Context, req dto.CreateRoleRequest) (*model.Role, error)
	List(ctx context.Context) ([]model.Role, error)
	BindUserRoles(ctx context.Context, userID uint64, roleCodes []string) error
	GetUserRoles(ctx context.Context, userID uint64) ([]model.Role, error)
}

type roleService struct {
	roles repository.RoleRepository
	users repository.UserRepository
}

func NewRoleService(roles repository.RoleRepository, users repository.UserRepository) RoleService {
	return &roleService{roles: roles, users: users}
}

func (s *roleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{Code: req.Code, Name: req.Name, Status: 1, Remark: req.Remark}
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) List(ctx context.Context) ([]model.Role, error) { return s.roles.List(ctx) }

func (s *roleService) BindUserRoles(ctx context.Context, userID uint64, roleCodes []string) error {
	roles, err := s.roles.GetByCodes(ctx, roleCodes)
	if err != nil {
		return err
	}
	if len(roles) != len(roleCodes) {
		return fmt.Errorf("some roles not found")
	}
	return s.users.SetRoles(ctx, userID, roles)
}

func (s *roleService) GetUserRoles(ctx context.Context, userID uint64) ([]model.Role, error) {
	return s.users.GetRoles(ctx, userID)
}
