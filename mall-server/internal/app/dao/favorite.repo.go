package dao

import (
	"errors"

	"gorm.io/gorm"
)

// ToggleFavorite 切换收藏状态。返回切换后是否已收藏。
func ToggleFavorite(db *gorm.DB, userID, productID uint) (bool, error) {
	var fav UserFavorite
	err := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&fav).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.Create(&UserFavorite{UserID: userID, ProductID: productID}).Error; createErr != nil {
			return false, createErr
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if delErr := db.Delete(&fav).Error; delErr != nil {
		return false, delErr
	}
	return false, nil
}

// IsFavorited 查询当前用户是否收藏了某商品。
func IsFavorited(db *gorm.DB, userID, productID uint) (bool, error) {
	var count int64
	err := db.Model(&UserFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// GetFavoriteList 获取用户收藏的在售商品列表（分页），按收藏时间倒序。
func GetFavoriteList(db *gorm.DB, userID uint, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&UserFavorite{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("JOIN product ON product.id = user_favorite.product_id AND product.status = 0").
		Joins("LEFT JOIN sys_user ON sys_user.id = product.user_id").
		Where("user_favorite.user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Order("user_favorite.created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
