package models

import (
	"gorm.io/gorm"
	"time"
)

// Product maps to the product table in the second-hand mall database.
type Product struct {
	ID           uint           `gorm:"primarykey;autoIncrement"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Title        string         `gorm:"column:title"`
	Description  string         `gorm:"column:description"`
	Price        float64        `gorm:"column:price"`
	Images       string         `gorm:"column:images"`
	Location     string         `gorm:"column:location"`
	Status       int            `gorm:"column:status"`
	UserID       uint           `gorm:"column:user_id"`
	BuyUID       uint           `gorm:"column:buy_uid"`
	ContactType  string         `gorm:"column:contact_type"`
	ContactValue string         `gorm:"column:contact_value"`
}

func (Product) TableName() string { return "product" }

// UserFavorite maps to the user_favorite table.
type UserFavorite struct {
	ID        uint      `gorm:"primarykey;autoIncrement"`
	UserID    uint      `gorm:"column:user_id"`
	ProductID uint      `gorm:"column:product_id"`
	CreatedAt time.Time
}

func (UserFavorite) TableName() string { return "user_favorite" }

// StatusLabel returns a human-readable label for a product status code.
func StatusLabel(status int) string {
	switch status {
	case 0:
		return "available"
	case 1:
		return "sold"
	case 2:
		return "removed"
	default:
		return "unknown"
	}
}
