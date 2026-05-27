package service_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/internal/app/service"
	"mall-server/pkg/jwtx"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newChatTestSetup(t *testing.T) (*gin.Engine, *models.ServiceContext) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&dao.SysUser{}, &dao.Product{}, &dao.Conversation{}, &dao.Message{})

	// seed two users
	db.Create(&dao.SysUser{UserName: "buyer", NickName: "买家", Avatar: "http://a.png"})
	db.Create(&dao.SysUser{UserName: "seller", NickName: "卖家", Avatar: "http://b.png"})
	// seed a product owned by seller (ID=2)
	db.Create(&dao.Product{Title: "旧手机", Images: "http://img.png", UserId: 2})

	svc := &models.ServiceContext{DB: db}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		// inject buyer (ID=1) as current user for all test requests
		c.Set("user_id", uint(1))
		c.Next()
	})
	r.POST("/api/chat/send", service.ChatSend(svc))
	r.GET("/api/chat/unread-count", service.ChatUnreadCount(svc))
	r.GET("/api/chat/conversations", service.ChatConversations(svc))
	r.GET("/api/chat/messages", service.ChatMessages(svc))
	r.PUT("/api/chat/read/:conv_id", service.ChatMarkRead(svc))
	return r, svc
}

func authHeader(t *testing.T, userID uint) string {
	t.Helper()
	token, err := jwtx.GenerateToken(userID, "testuser")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return "Bearer " + token
}

func TestChatSend_CreatesConversationAndMessage(t *testing.T) {
	r, _ := newChatTestSetup(t)
	body, _ := json.Marshal(map[string]interface{}{
		"product_id":  1,
		"receiver_id": 2,
		"content":     "还在吗？",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Fatalf("expected code=0: %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["conversation_id"].(float64) == 0 {
		t.Fatal("expected non-zero conversation_id")
	}
}

func TestChatUnreadCount(t *testing.T) {
	r, svc := newChatTestSetup(t)
	// create a message from seller(2) to buyer(1)
	conv, _ := dao.FindOrCreateConversation(svc.DB, 1, 1, 2)
	dao.SaveMessage(svc.DB, conv.ID, 2, "你好")

	req := httptest.NewRequest(http.MethodGet, "/api/chat/unread-count", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["count"].(float64) != 1 {
		t.Fatalf("expected count=1, got %v", data["count"])
	}
}

func TestChatMarkRead_ClearsUnread(t *testing.T) {
	r, svc := newChatTestSetup(t)
	conv, _ := dao.FindOrCreateConversation(svc.DB, 1, 1, 2)
	dao.SaveMessage(svc.DB, conv.ID, 2, "你好")

	url := fmt.Sprintf("/api/chat/read/%d", conv.ID)
	req := httptest.NewRequest(http.MethodPut, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	count, _ := dao.UnreadCount(svc.DB, 1)
	if count != 0 {
		t.Fatalf("expected 0 unread after mark-read, got %d", count)
	}
}
