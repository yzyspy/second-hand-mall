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

// AdminListUsersResult 用户列表查询结果
type AdminListUsersResult struct {
	ID           uint   `json:"id"`
	UserName     string `json:"user_name"`
	NickName     string `json:"nick_name"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	IsBanned     bool   `json:"is_banned"`
	ProductCount int64  `json:"product_count"`
	CreatedAt    string `json:"created_at"`
}

// AdminListUsers 分页查询用户列表
func AdminListUsers(db *gorm.DB, keyword string, isBanned *bool, page, pageSize int) ([]AdminListUsersResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&SysUser{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR nick_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if isBanned != nil {
		query = query.Where("is_banned = ?", *isBanned)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []SysUser
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	results := make([]AdminListUsersResult, 0, len(users))
	for _, u := range users {
		var count int64
		db.Model(&Product{}).Where("user_id = ? AND deleted_at IS NULL", u.ID).Count(&count)
		results = append(results, AdminListUsersResult{
			ID:           u.ID,
			UserName:     u.UserName,
			NickName:     u.NickName,
			Phone:        u.Phone,
			Avatar:       u.Avatar,
			IsBanned:     u.IsBanned,
			ProductCount: count,
			CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return results, total, nil
}

// AdminGetUserDetail 获取用户详情（含商品数、收藏数）
func AdminGetUserDetail(db *gorm.DB, id uint) (*AdminListUsersResult, error) {
	var u SysUser
	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}
	var productCount, favoriteCount int64
	db.Model(&Product{}).Where("user_id = ? AND deleted_at IS NULL", id).Count(&productCount)
	db.Model(&UserFavorite{}).Where("user_id = ?", id).Count(&favoriteCount)

	return &AdminListUsersResult{
		ID:           u.ID,
		UserName:     u.UserName,
		NickName:     u.NickName,
		Phone:        u.Phone,
		Avatar:       u.Avatar,
		IsBanned:     u.IsBanned,
		ProductCount: productCount,
		CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminSetUserBanned 设置用户封禁状态
func AdminSetUserBanned(db *gorm.DB, id uint, banned bool) error {
	return db.Model(&SysUser{}).Where("id = ?", id).Update("is_banned", banned).Error
}
