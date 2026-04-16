package dao

import "gorm.io/gorm"

// 增加字段注释
type SysUser struct {
	gorm.Model `json:"gorm_._model"`
	UserName   string `gorm:"column:username;type:varchar(50);not null;default:''" json:"user_name" comment:"用户名"`
	Password   string `gorm:"column:password;type:varchar(100);not null;default:''" json:"password" comment:"密码"`
	Phone      string `gorm:"column:phone;type:varchar(20);not null;default:''" json:"phone" comment:"手机号"`
	WxUserid   string `gorm:"column:wx_userid;type:varchar(50);not null;default:''" json:"wx_userid" comment:"微信用户ID"`
	WxOpenid   string `gorm:"column:wx_openid;type:varchar(50);not null;default:''" json:"wx_openid" comment:"微信开放ID"`
	Avatar     string `gorm:"column:avatar;type:varchar(255);not null;default:''" json:"avatar" comment:"头像"`
	Sex        string `gorm:"column:sex;type:varchar(20);not null;default:''" json:"sex" comment:"性别"`
	Email      string `gorm:"column:email;type:varchar(100);not null;default:''" json:"email" comment:"邮箱"`
	Remarks    string `gorm:"column:remarks;type:varchar(255);not null;default:''" json:"remarks" comment:"备注"`
	RoleId     int    `gorm:"column:role_id;type:int;not null;default:0" json:"role_id" comment:"角色ID"`
}

// 设置表名称
func (SysUser) TableName() string {
	return "sys_user"
}

func (u *SysUser) Save(tx *gorm.DB) error {
	tx = tx.Save(u)
	return tx.Error
}

//func GetUserByUserNameAndPsw(userName string, password string) (*SysUser, error) {
//	//u := SysUser{}
//	u := new(SysUser)
//	err := DB.Where("username = ? AND password = ?", userName, password).First(u).Error
//	return u, err
//}
