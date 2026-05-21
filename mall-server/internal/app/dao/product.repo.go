package dao

import (
	"fmt"

	"gorm.io/gorm"
)

type ProductSearchResult struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Images     string  `json:"images"`
	Location   string  `json:"location"`
	Status     int     `json:"status"`
	Category   string  `json:"category"`
	Seller     string  `json:"seller"`
	Avatar     string  `json:"avatar"`
	BuyUid     uint    `json:"buy_uid"`
	CreateTime string  `json:"create_time"`
}

type ProductDetail struct {
	ID           uint    `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Images       string  `json:"images"`
	Location     string  `json:"location"`
	Status       int     `json:"status"`
	BuyUid       uint    `json:"buy_uid"`
	Category     string  `json:"category"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Seller       string  `json:"seller"`
	Avatar       string  `json:"avatar"`
	CreateTime   string  `json:"create_time"`
	ContactType  string  `json:"contact_type"`
	ContactValue string  `json:"contact_value"`
	IsFavorited  bool    `json:"is_favorited"`
}

func SearchProducts(db *gorm.DB, keyword, sort string, status *int,
	category, province, city, district string,
	page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.category, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id")

	if keyword != "" {
		query = query.Where("product.title LIKE ?", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("product.status = ?", *status)
	}
	if category != "" {
		query = query.Where("product.category = ?", category)
	}
	if province != "" {
		query = query.Where("product.province = ?", province)
	}
	if city != "" {
		query = query.Where("product.city = ?", city)
	}
	if district != "" {
		query = query.Where("product.district = ?", district)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch sort {
	case "time_asc":
		query = query.Order("product.created_at ASC")
	default:
		query = query.Order("product.created_at DESC")
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func GetProductByID(db *gorm.DB, id uint) (*ProductDetail, error) {
	var detail ProductDetail
	err := db.Model(&Product{}).
		Select("product.id, product.title, product.description, product.price, product.images, product.location, product.status, product.buy_uid, product.category, product.province, product.city, product.district, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time, product.contact_type, product.contact_value").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id").
		Where("product.id = ?", id).
		First(&detail).Error

	if err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	return &detail, nil
}

// GetMyProducts 获取指定用户发布的商品列表
func GetMyProducts(db *gorm.DB, userID uint, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.category, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id").
		Where("product.user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Order("product.created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// UpdateProduct 更新商品内容，仅允许在售（status=0）且属于本人的商品
func UpdateProduct(db *gorm.DB, id uint, userID uint, updates map[string]interface{}) error {
	result := db.Model(&Product{}).
		Where("id = ? AND user_id = ? AND status = 0", id, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在或无权限编辑")
	}
	return nil
}

// ChangeProductStatus 变更商品状态，须本人校验
func ChangeProductStatus(db *gorm.DB, id uint, userID uint, status int) error {
	result := db.Model(&Product{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在或无权限操作")
	}
	return nil
}
