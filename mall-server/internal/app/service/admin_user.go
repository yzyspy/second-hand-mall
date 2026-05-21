package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// AdminListUsers GET /admin/users
func AdminListUsers(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminUserListRequest
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

		results, total, err := dao.AdminListUsers(svc.DB, req.Keyword, req.IsBanned, req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0, "msg": "success",
			"data": gin.H{"list": results, "total": total, "page": req.Page, "page_size": req.PageSize},
		})
	}
}

// AdminGetUserDetail GET /admin/users/:id
func AdminGetUserDetail(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		detail, err := dao.AdminGetUserDetail(svc.DB, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": detail})
	}
}

// AdminBanUser POST /admin/users/:id/ban
func AdminBanUser(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if err := dao.AdminSetUserBanned(svc.DB, uint(id), true); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "封禁成功"})
	}
}

// AdminUnbanUser POST /admin/users/:id/unban
func AdminUnbanUser(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if err := dao.AdminSetUserBanned(svc.DB, uint(id), false); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "解封成功"})
	}
}
