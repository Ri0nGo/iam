package repository

import (
	"context"

	"iam/internal/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	List(ctx context.Context) ([]model.Role, error)
	GetByCodes(ctx context.Context, codes []string) ([]model.Role, error)
}

type roleRepository struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository { return &roleRepository{db: db} }

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Order("id DESC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByCodes(ctx context.Context, codes []string) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&roles).Error
	return roles, err
}
