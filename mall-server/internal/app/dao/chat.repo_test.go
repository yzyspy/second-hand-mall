package dao_test

import (
	"mall-server/internal/app/dao"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&dao.Conversation{}, &dao.Message{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestFindOrCreateConversation_CreatesNew(t *testing.T) {
	db := newTestDB(t)
	conv, err := dao.FindOrCreateConversation(db, 1, 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if conv.ProductID != 1 || conv.BuyerID != 2 || conv.SellerID != 3 {
		t.Fatalf("wrong fields: %+v", conv)
	}
}

func TestFindOrCreateConversation_ReturnsExisting(t *testing.T) {
	db := newTestDB(t)
	first, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	second, err := dao.FindOrCreateConversation(db, 1, 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same ID, got %d and %d", first.ID, second.ID)
	}
}

func TestSaveMessage(t *testing.T) {
	db := newTestDB(t)
	conv, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	msg, err := dao.SaveMessage(db, conv.ID, 2, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID == 0 || msg.Content != "hello" {
		t.Fatalf("wrong message: %+v", msg)
	}
}

func TestUnreadCount(t *testing.T) {
	db := newTestDB(t)
	conv, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	dao.SaveMessage(db, conv.ID, 2, "msg1") // sender=2, receiver=3
	dao.SaveMessage(db, conv.ID, 2, "msg2")

	count, err := dao.UnreadCount(db, 3) // user 3 has 2 unread
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestMarkRead(t *testing.T) {
	db := newTestDB(t)
	conv, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	dao.SaveMessage(db, conv.ID, 2, "msg1")
	dao.MarkRead(db, conv.ID, 3)

	count, _ := dao.UnreadCount(db, 3)
	if count != 0 {
		t.Fatalf("expected 0 after mark-read, got %d", count)
	}
}

func TestListMessages_IncrementalByLastID(t *testing.T) {
	db := newTestDB(t)
	conv, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	m1, _ := dao.SaveMessage(db, conv.ID, 2, "first")
	dao.SaveMessage(db, conv.ID, 3, "second")

	msgs, err := dao.ListMessages(db, conv.ID, m1.ID) // only return id > m1.ID
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "second" {
		t.Fatalf("unexpected messages: %+v", msgs)
	}
}

func TestListConversations(t *testing.T) {
	db := newTestDB(t)
	conv, _ := dao.FindOrCreateConversation(db, 1, 2, 3)
	dao.SaveMessage(db, conv.ID, 2, "hey") // sender=2, unread for user 3

	rows, err := dao.ListConversations(db, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(rows))
	}
	if rows[0].UnreadCount != 1 {
		t.Fatalf("expected 1 unread, got %d", rows[0].UnreadCount)
	}
	if rows[0].LastMessage != "hey" {
		t.Fatalf("expected last_message='hey', got '%s'", rows[0].LastMessage)
	}
}
