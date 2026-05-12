package service

import (
	"context"
	"fmt"

	"iam/internal/dto"
	"iam/internal/model"
	"iam/internal/pkg/password"
	"iam/internal/pkg/random"
	"iam/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, req dto.CreateUserRequest) (*model.User, error)
	Get(ctx context.Context, id uint64) (*model.User, error)
	List(ctx context.Context, query dto.UserListQuery) ([]model.User, error)
	UpdateStatus(ctx context.Context, id uint64, status int) error
	Delete(ctx context.Context, id uint64) error
	ResetPassword(ctx context.Context, id uint64, newPassword string) error
	GetRoles(ctx context.Context, id uint64) ([]model.Role, error)
}

type userService struct {
	users        repository.UserRepository
	roles        repository.RoleRepository
	identities   repository.AuthIdentityRepository
	passwordCost int
}

func NewUserService(users repository.UserRepository, roles repository.RoleRepository, identities repository.AuthIdentityRepository, passwordCost int) UserService {
	return &userService{users: users, roles: roles, identities: identities, passwordCost: passwordCost}
}

func (s *userService) Create(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	openid, err := generateOpenID()
	if err != nil {
		return nil, err
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	user := &model.User{
		Username:    req.Username,
		OpenID:      &openid,
		DisplayName: req.DisplayName,
		Email:       optionalString(req.Email),
		Mobile:      optionalString(req.Mobile),
		Status:      status,
		Remark:      req.Remark,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	hash, err := password.Hash(req.Password, s.passwordCost)
	if err != nil {
		return nil, err
	}
	if err := s.identities.Create(ctx, &model.AuthIdentity{UserID: user.ID, IdentityType: "password", Identifier: req.Username, Credential: hash}); err != nil {
		return nil, err
	}
	if len(req.RoleCodes) > 0 {
		roles, err := s.roles.GetByCodes(ctx, req.RoleCodes)
		if err != nil {
			return nil, err
		}
		if len(roles) != len(req.RoleCodes) {
			return nil, fmt.Errorf("some roles not found")
		}
		if err := s.users.SetRoles(ctx, user.ID, roles); err != nil {
			return nil, err
		}
	}
	return s.users.GetByID(ctx, user.ID)
}

func (s *userService) Get(ctx context.Context, id uint64) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}
func (s *userService) List(ctx context.Context, query dto.UserListQuery) ([]model.User, error) {
	return s.users.List(ctx, query.Keyword, query.Status)
}
func (s *userService) UpdateStatus(ctx context.Context, id uint64, status int) error {
	return s.users.UpdateStatus(ctx, id, status)
}
func (s *userService) Delete(ctx context.Context, id uint64) error {
	return s.users.Delete(ctx, id)
}
func (s *userService) ResetPassword(ctx context.Context, id uint64, newPassword string) error {
	hash, err := password.Hash(newPassword, s.passwordCost)
	if err != nil {
		return err
	}
	return s.identities.UpdateCredential(ctx, id, "password", hash)
}
func (s *userService) GetRoles(ctx context.Context, id uint64) ([]model.Role, error) {
	return s.users.GetRoles(ctx, id)
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func generateOpenID() (string, error) {
	value, err := random.Hex(16)
	if err != nil {
		return "", err
	}
	return "ou_" + value, nil
}
