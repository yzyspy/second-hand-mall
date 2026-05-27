package dao

import (
	"gorm.io/gorm"
	"time"
)

// FindOrCreateConversation returns existing conversation or creates one.
// productID+buyerID is the unique key.
func FindOrCreateConversation(db *gorm.DB, productID, buyerID, sellerID uint) (*Conversation, error) {
	conv := &Conversation{}
	err := db.Where(Conversation{ProductID: productID, BuyerID: buyerID}).
		Attrs(Conversation{SellerID: sellerID, LastAt: time.Now()}).
		FirstOrCreate(conv).Error
	return conv, err
}

// SaveMessage inserts a message and updates the parent conversation's preview.
func SaveMessage(db *gorm.DB, conversationID, senderID uint, content string) (*Message, error) {
	msg := &Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		IsRead:         false,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, err
	}
	preview := content
	if len([]rune(preview)) > 30 {
		preview = string([]rune(preview)[:30]) + "…"
	}
	if err := db.Model(&Conversation{}).Where("id = ?", conversationID).
		Updates(map[string]interface{}{"last_message": preview, "last_at": time.Now()}).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// UnreadCount returns total unread messages across all conversations for userID.
func UnreadCount(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&Message{}).
		Joins("JOIN chat_conversation ON chat_conversation.id = chat_message.conversation_id").
		Where("chat_message.is_read = ? AND chat_message.sender_id != ? AND (chat_conversation.buyer_id = ? OR chat_conversation.seller_id = ?)",
			false, userID, userID, userID).
		Count(&count).Error
	return count, err
}

// ListConversations returns conversations for userID with per-conversation unread count.
type ConversationRow struct {
	ConvID      uint
	ProductID   uint
	BuyerID     uint
	SellerID    uint
	LastMessage string
	LastAt      time.Time
	UnreadCount int64
}

func ListConversations(db *gorm.DB, userID uint) ([]ConversationRow, error) {
	var rows []ConversationRow
	err := db.Model(&Conversation{}).
		Select(`chat_conversation.id as conv_id, chat_conversation.product_id,
			chat_conversation.buyer_id, chat_conversation.seller_id,
			chat_conversation.last_message, chat_conversation.last_at,
			COUNT(CASE WHEN chat_message.is_read = false AND chat_message.sender_id != ? THEN 1 END) as unread_count`, userID).
		Joins("LEFT JOIN chat_message ON chat_message.conversation_id = chat_conversation.id AND chat_message.deleted_at IS NULL").
		Where("chat_conversation.buyer_id = ? OR chat_conversation.seller_id = ?", userID, userID).
		Where("chat_conversation.deleted_at IS NULL").
		Group("chat_conversation.id").
		Order("chat_conversation.last_at DESC").
		Scan(&rows).Error
	return rows, err
}

// ListMessages returns messages in a conversation with id > lastID.
// Pass lastID=0 to get the most recent 50 messages (returned in ascending order).
func ListMessages(db *gorm.DB, conversationID, lastID uint) ([]Message, error) {
	var msgs []Message
	var err error
	if lastID > 0 {
		err = db.Where("conversation_id = ? AND id > ?", conversationID, lastID).
			Order("id ASC").Find(&msgs).Error
	} else {
		var recent []Message
		err = db.Where("conversation_id = ?", conversationID).
			Order("id DESC").Limit(50).Find(&recent).Error
		if err == nil {
			for i, j := 0, len(recent)-1; i < j; i, j = i+1, j-1 {
				recent[i], recent[j] = recent[j], recent[i]
			}
			msgs = recent
		}
	}
	return msgs, err
}

// MarkRead marks all messages in conversationID as read where sender != userID.
func MarkRead(db *gorm.DB, conversationID, userID uint) error {
	return db.Model(&Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conversationID, userID, false).
		Update("is_read", true).Error
}
