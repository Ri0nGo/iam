package repository

import (
	"context"
	"strings"

	"iam/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, keyword string, status int) ([]model.User, error)
	UpdateStatus(ctx context.Context, id uint64, status int) error
	Delete(ctx context.Context, id uint64) error
	SetRoles(ctx context.Context, userID uint64, roles []model.Role) error
	GetRoles(ctx context.Context, userID uint64) ([]model.Role, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, keyword string, status int) ([]model.User, error) {
	var users []model.User
	q := r.db.WithContext(ctx).Preload("Roles").Model(&model.User{})
	if status != 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		q = q.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ? OR mobile LIKE ?", kw, kw, kw, kw)
	}
	err := q.Order("id DESC").Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uint64, status int) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.AuthIdentity{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, id).Error
	})
}

func (r *userRepository) SetRoles(ctx context.Context, userID uint64, roles []model.Role) error {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&user).Association("Roles").Replace(roles)
}

func (r *userRepository) GetRoles(ctx context.Context, userID uint64) ([]model.Role, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return user.Roles, nil
}
