package repository

import (
	"context"

	"iam/internal/model"

	"gorm.io/gorm"
)

type AuthIdentityRepository interface {
	Create(ctx context.Context, identity *model.AuthIdentity) error
	GetByIdentity(ctx context.Context, identityType string, identifier string) (*model.AuthIdentity, error)
	UpdateCredential(ctx context.Context, userID uint64, identityType string, credential string) error
}

type authIdentityRepository struct{ db *gorm.DB }

func NewAuthIdentityRepository(db *gorm.DB) AuthIdentityRepository {
	return &authIdentityRepository{db: db}
}

func (r *authIdentityRepository) Create(ctx context.Context, identity *model.AuthIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

func (r *authIdentityRepository) GetByIdentity(ctx context.Context, identityType string, identifier string) (*model.AuthIdentity, error) {
	var identity model.AuthIdentity
	err := r.db.WithContext(ctx).Where("identity_type = ? AND identifier = ?", identityType, identifier).First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *authIdentityRepository) UpdateCredential(ctx context.Context, userID uint64, identityType string, credential string) error {
	return r.db.WithContext(ctx).Model(&model.AuthIdentity{}).
		Where("user_id = ? AND identity_type = ?", userID, identityType).
		Update("credential", credential).Error
}
