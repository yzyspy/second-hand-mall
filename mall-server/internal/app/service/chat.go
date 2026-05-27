package service

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// ChatSend creates or reuses a conversation and inserts a message.
func ChatSend(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		buyerID := userID.(uint)

		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		// fetch product to get seller ID
		var product dao.Product
		if err := svc.DB.First(&product, req.ProductID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "商品不存在"})
			return
		}
		sellerID := product.UserId
		var actualBuyerID uint
		if buyerID == sellerID {
			// Current user is the seller; the receiver is the buyer
			actualBuyerID = req.ReceiverID
		} else {
			actualBuyerID = buyerID
		}

		conv, err := dao.FindOrCreateConversation(svc.DB, req.ProductID, actualBuyerID, sellerID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "创建会话失败"})
			return
		}

		msg, err := dao.SaveMessage(svc.DB, conv.ID, buyerID, req.Content)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "发送失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": SendMessageResponse{
				ConversationID: conv.ID,
				MessageID:      msg.ID,
			},
		})
	}
}

// ChatUnreadCount returns total unread message count for the current user.
func ChatUnreadCount(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		count, err := dao.UnreadCount(svc.DB, userID.(uint))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": UnreadCountResponse{Count: count}})
	}
}

// ChatConversations returns all conversations for the current user with join data.
func ChatConversations(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		myID := userID.(uint)

		rows, err := dao.ListConversations(svc.DB, myID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}

		items := make([]ConversationItem, 0, len(rows))
		for _, row := range rows {
			otherID := row.BuyerID
			if row.BuyerID == myID {
				otherID = row.SellerID
			}

			var product dao.Product
			svc.DB.Select("id, title, images").First(&product, row.ProductID)

			var other dao.SysUser
			svc.DB.Select("id, nick_name, avatar").First(&other, otherID)

			cover := ""
			if product.Images != "" {
				cover = strings.SplitN(product.Images, ",", 2)[0]
			}

			items = append(items, ConversationItem{
				ConversationID: row.ConvID,
				Product: ProductBrief{
					ID:    product.ID,
					Title: product.Title,
					Cover: cover,
				},
				OtherUser: UserBrief{
					ID:       other.ID,
					Nickname: other.NickName,
					Avatar:   other.Avatar,
				},
				LastMessage: row.LastMessage,
				LastAt:      row.LastAt.Format(time.RFC3339),
				UnreadCount: row.UnreadCount,
			})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": items})
	}
}

// ChatMessages returns messages for a conversation, filtered by last_id for polling.
func ChatMessages(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("user_id"); !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		var req MessagesQueryRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		// Verify user is a participant in this conversation
		var conv dao.Conversation
		if err := svc.DB.First(&conv, req.ConvID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "会话不存在"})
			return
		}
		myID := c.MustGet("user_id").(uint)
		if conv.BuyerID != myID && conv.SellerID != myID {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权访问"})
			return
		}
		msgs, err := dao.ListMessages(svc.DB, req.ConvID, req.LastID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}
		items := make([]MessageItem, 0, len(msgs))
		for _, m := range msgs {
			items = append(items, MessageItem{
				ID:        m.ID,
				SenderID:  m.SenderID,
				Content:   m.Content,
				CreatedAt: m.CreatedAt.Format(time.RFC3339),
			})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": items})
	}
}

// ChatMarkRead marks all unread messages in a conversation as read for the current user.
func ChatMarkRead(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		convIDStr := c.Param("conv_id")
		convID64, err := strconv.ParseUint(convIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		// Verify user is a participant
		var conv dao.Conversation
		if err := svc.DB.First(&conv, uint(convID64)).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "会话不存在"})
			return
		}
		if conv.BuyerID != userID.(uint) && conv.SellerID != userID.(uint) {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权访问"})
			return
		}
		if err := dao.MarkRead(svc.DB, uint(convID64), userID.(uint)); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "标记失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": gin.H{"ok": true}})
	}
}
