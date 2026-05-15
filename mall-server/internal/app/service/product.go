package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

func SearchProducts(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchProductRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "参数错误",
			})
			return
		}

		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 50 {
			req.PageSize = 10
		}

		results, total, err := dao.SearchProducts(svc.DB, req.Keyword, req.Sort, req.Status, req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "搜索失败",
			})
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

// GetProductDetail 获取商品详情
func GetProductDetail(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "参数错误",
			})
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "参数错误",
			})
			return
		}

		detail, err := dao.GetProductByID(svc.DB, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "商品不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": detail,
		})
	}
}

// PublishProduct 发布商品
func PublishProduct(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "未登录",
			})
			return
		}

		var req PublishProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "参数错误: " + err.Error(),
			})
			return
		}

		imagesStr := strings.Join(req.Images, ",")

		product := dao.Product{
			Title:       req.Title,
			Description: req.Description,
			Price:       req.Price,
			Images:      imagesStr,
			Location:    req.Location,
			Status:      0,
			UserId:      userID.(uint),
			BuyUid:      0,
		}

		if err := svc.DB.Create(&product).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "发布失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "发布成功",
			"data": gin.H{
				"id": product.ID,
			},
		})
	}
}

// GetMyProducts 获取当前用户发布的商品列表
func GetMyProducts(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req GetMyProductsRequest
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

		results, total, err := dao.GetMyProducts(svc.DB, userID.(uint), req.Page, req.PageSize)
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

// UpdateProduct 更新商品内容
func UpdateProduct(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误: " + err.Error()})
			return
		}

		// title 取 description 前 50 字，与 publish 保持一致
		title := req.Description
		runes := []rune(title)
		if len(runes) > 50 {
			title = string(runes[:50])
		}

		updates := map[string]interface{}{
			"title":       title,
			"description": req.Description,
			"price":       req.Price,
			"location":    req.Location,
			"images":      strings.Join(req.Images, ","),
		}

		if err := dao.UpdateProduct(svc.DB, req.ID, userID.(uint), updates); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
	}
}

// ChangeProductStatus 变更商品状态
func ChangeProductStatus(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req ChangeProductStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误: " + err.Error()})
			return
		}
		if req.Status != 1 && req.Status != 2 {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "状态值无效，只允许 1(已售) 或 2(下架)"})
			return
		}

		if err := dao.ChangeProductStatus(svc.DB, req.ID, userID.(uint), req.Status); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}
