package dao

import (
	"gorm.io/gorm"
)

type ProductSearchResult struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Images     string  `json:"images"`
	Location   string  `json:"location"`
	Status     int     `json:"status"`
	Seller     string  `json:"seller"`
	CreateTime string  `json:"create_time"`
}

func SearchProducts(db *gorm.DB, keyword string, sort string, status *int, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).Select("product.id, product.title, product.price, product.images, product.location, product.status, sys_user.nick_name as seller, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id")

	if keyword != "" {
		query = query.Where("product.title LIKE ?", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("product.status = ?", *status)
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
