package dao

import (
	"gorm.io/gorm"
	"time"
)

type Conversation struct {
	gorm.Model
	ProductID   uint      `gorm:"column:product_id;not null;uniqueIndex:idx_conv_product_buyer"`
	BuyerID     uint      `gorm:"column:buyer_id;not null;uniqueIndex:idx_conv_product_buyer"`
	SellerID    uint      `gorm:"column:seller_id;not null"`
	LastMessage string    `gorm:"column:last_message;type:varchar(200);not null;default:''"`
	LastAt      time.Time `gorm:"column:last_at;not null"`
}

func (Conversation) TableName() string { return "chat_conversation" }

type Message struct {
	gorm.Model
	ConversationID uint   `gorm:"column:conversation_id;not null;index"`
	SenderID       uint   `gorm:"column:sender_id;not null"`
	Content        string `gorm:"column:content;type:varchar(500);not null"`
	IsRead         bool   `gorm:"column:is_read;not null;default:false"`
}

func (Message) TableName() string { return "chat_message" }
