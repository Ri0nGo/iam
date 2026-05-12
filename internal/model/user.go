package model

type User struct {
	Base
	Username    string  `gorm:"size:64;uniqueIndex;not null" json:"username"`
	OpenID      *string `gorm:"size:128;uniqueIndex" json:"openid"`
	DisplayName string  `gorm:"size:128;not null" json:"display_name"`
	AvatarURL   string  `gorm:"size:255;not null;default:''" json:"avatar_url"`
	Mobile      *string `gorm:"size:32;uniqueIndex" json:"mobile"`
	Email       *string `gorm:"size:128;uniqueIndex" json:"email"`
	Status      int     `gorm:"not null;default:1" json:"status"`
	Remark      string  `gorm:"size:255;not null;default:''" json:"remark"`
	Roles       []Role  `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}
