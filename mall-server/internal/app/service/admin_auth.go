package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/pkg/jwtx"
)

// AdminLogin POST /admin/login
func AdminLogin(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		admin, err := dao.GetAdminByUsername(svc.DB, req.Username)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户名或密码错误"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户名或密码错误"})
			return
		}

		token, err := jwtx.GenerateAdminToken(admin.ID, admin.Username)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "生成 token 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "登录成功",
			"data": gin.H{"token": token, "username": admin.Username},
		})
	}
}
