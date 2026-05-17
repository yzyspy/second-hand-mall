package dao

import "time"

type UserFavorite struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_product"`
	ProductID uint      `gorm:"not null;uniqueIndex:idx_user_product"`
	CreatedAt time.Time
}

func (UserFavorite) TableName() string {
	return "user_favorite"
}
