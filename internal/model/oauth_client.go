package model

type OAuthClient struct {
	Base
	ClientID     string `gorm:"size:64;uniqueIndex;not null" json:"client_id"`
	ClientSecret string `gorm:"size:128;not null" json:"-"`
	Name         string `gorm:"size:128;not null" json:"name"`
	Code         string `gorm:"size:64;not null;default:''" json:"code"`
	ResponseType string `gorm:"size:32;not null;default:'code'" json:"response_type"`
	RedirectURI  string `gorm:"size:255;not null" json:"redirect_uri"`
	Status       int    `gorm:"not null;default:1" json:"status"`
	Remark       string `gorm:"size:255;not null;default:''" json:"remark"`
}

func (o *OAuthClient) TableName() string {
	return "oauth_clients"
}
