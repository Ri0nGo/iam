package model

type Role struct {
	Base
	Code   string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name   string `gorm:"size:128;not null" json:"name"`
	Status int    `gorm:"not null;default:1" json:"status"`
	Remark string `gorm:"size:255;not null;default:''" json:"remark"`
}
