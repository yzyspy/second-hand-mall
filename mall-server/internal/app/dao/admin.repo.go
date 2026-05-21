package dao

import (
	"gorm.io/gorm"
)

// GetAdminByUsername 根据用户名查询管理员
func GetAdminByUsername(db *gorm.DB, username string) (*AdminUser, error) {
	var admin AdminUser
	err := db.Where("username = ?", username).First(&admin).Error
	return &admin, err
}
