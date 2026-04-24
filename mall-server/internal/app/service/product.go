package service

import (
	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"net/http"
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
