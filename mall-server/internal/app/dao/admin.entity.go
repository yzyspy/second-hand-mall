package dao

import "gorm.io/gorm"

type AdminUser struct {
	gorm.Model
	Username     string `gorm:"column:username;type:varchar(50);not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"       json:"-"`
}

func (AdminUser) TableName() string {
	return "admin_user"
}
