package model

type AuthIdentity struct {
	Base
	UserID       uint64 `gorm:"index;not null" json:"user_id"`
	IdentityType string `gorm:"size:32;uniqueIndex:uk_identity;not null" json:"identity_type"`
	Identifier   string `gorm:"size:128;uniqueIndex:uk_identity;not null" json:"identifier"`
	Credential   string `gorm:"size:255;not null;default:''" json:"-"`
	Extra        string `gorm:"type:text" json:"extra"`
}
