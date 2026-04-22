package dao

import "gorm.io/gorm"

// GetUserByUserName 根据用户名查询用户
func GetUserByUserName(tx *gorm.DB, userName string) (*SysUser, error) {
	u := new(SysUser)
	err := tx.Where("username = ?", userName).First(u).Error
	return u, err
}

func (u *SysUser) Save(tx *gorm.DB) error {
	tx = tx.Save(u)
	return tx.Error
}
