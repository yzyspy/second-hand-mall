package models

import (
	"gorm.io/gorm"
	"time"
)

// SysUser maps to the sys_user table in the second-hand mall database.
type SysUser struct {
	ID           uint           `gorm:"primarykey;autoIncrement"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Username     string         `gorm:"column:username"`
	Password     string         `gorm:"column:password"`
	Phone        string         `gorm:"column:phone"`
	WxUserid     string         `gorm:"column:wx_userid"`
	WxOpenid     string         `gorm:"column:wx_openid"`
	Avatar       string         `gorm:"column:avatar"`
	Sex          string         `gorm:"column:sex"`
	Email        string         `gorm:"column:email"`
	Remarks      string         `gorm:"column:remarks"`
	RoleID       int            `gorm:"column:role_id"`
	WxSessionKey string         `gorm:"column:wx_session_key"`
	WxUnionid    string         `gorm:"column:wx_unionid"`
	NickName     string         `gorm:"column:nick_name"`
}

func (SysUser) TableName() string { return "sys_user" }

// MaskPhone replaces the 4th–7th characters (zero-indexed) with '*'.
// "13812345678" → "138****5678"
func MaskPhone(phone string) string {
	if len(phone) < 8 {
		return phone
	}
	runes := []rune(phone)
	for i := 3; i < 7 && i < len(runes); i++ {
		runes[i] = '*'
	}
	return string(runes)
}
