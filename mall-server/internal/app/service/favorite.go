package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// ToggleFavoriteHandler POST /api/favorite/toggle
func ToggleFavoriteHandler(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req FavoriteToggleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		isFav, err := dao.ToggleFavorite(svc.DB, userID.(uint), req.ProductID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{"is_favorited": isFav},
		})
	}
}

// GetFavoriteListHandler GET /api/favorite/list
func GetFavoriteListHandler(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req FavoriteListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 50 {
			req.PageSize = 10
		}

		results, total, err := dao.GetFavoriteList(svc.DB, userID.(uint), req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{
				"list":      results,
				"total":     total,
				"page":      req.Page,
				"page_size": req.PageSize,
			},
		})
	}
}
