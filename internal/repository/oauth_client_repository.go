package repository

import (
	"context"
	"strings"

	"iam/internal/model"

	"gorm.io/gorm"
)

type OAuthClientRepository interface {
	GetByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error)
	GetByID(ctx context.Context, id uint64) (*model.OAuthClient, error)
	List(ctx context.Context, keyword string, status int) ([]model.OAuthClient, error)
	Create(ctx context.Context, client *model.OAuthClient) error
	Update(ctx context.Context, client *model.OAuthClient) error
	Delete(ctx context.Context, id uint64) error
}

type oAuthClientRepository struct{ db *gorm.DB }

func NewOAuthClientRepository(db *gorm.DB) OAuthClientRepository {
	return &oAuthClientRepository{db: db}
}

func (r *oAuthClientRepository) GetByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error) {
	var client model.OAuthClient
	err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *oAuthClientRepository) Create(ctx context.Context, client *model.OAuthClient) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *oAuthClientRepository) GetByID(ctx context.Context, id uint64) (*model.OAuthClient, error) {
	var client model.OAuthClient
	err := r.db.WithContext(ctx).First(&client, id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *oAuthClientRepository) List(ctx context.Context, keyword string, status int) ([]model.OAuthClient, error) {
	var clients []model.OAuthClient
	q := r.db.WithContext(ctx).Model(&model.OAuthClient{})
	if status != 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR client_id LIKE ? OR redirect_uri LIKE ?", kw, kw, kw, kw)
	}
	err := q.Order("id DESC").Find(&clients).Error
	return clients, err
}

func (r *oAuthClientRepository) Update(ctx context.Context, client *model.OAuthClient) error {
	return r.db.WithContext(ctx).Save(client).Error
}

func (r *oAuthClientRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.OAuthClient{}, id).Error
}
