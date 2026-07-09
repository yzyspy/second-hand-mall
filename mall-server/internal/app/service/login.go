package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/pkg/jwtx"
	"mall-server/pkg/logger"
	"net/http"
	"strings"
)

func LoginPsw(ctx context.Context, c *gin.Context, svc *models.ServiceContext) {
	request := new(LoginPswRequest)

	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	logger.WithContext(c).Infof("LoginPsw request: %s", request.Username)

	// 根据用户名查询用户
	user, err := dao.GetUserByUserName(svc.DB, request.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "用户不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "查询用户失败",
		})
		return
	}

	// 验证密码：优先 bcrypt；历史数据为明文，比对成功后原地升级为哈希
	if !verifyAndUpgradePassword(svc.DB, user, request.Password) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "密码错误",
		})
		return
	}

	if user.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{
			"code": -1,
			"msg":  "账号已被封禁，请联系管理员",
		})
		return
	}

	// 生成 JWT token
	token, err := jwtx.GenerateToken(user.ID, user.UserName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "登录成功",
		"data": gin.H{
			"user_id":   user.ID,
			"user_name": user.UserName,
			"avatar":    user.Avatar,
			"token":     token,
		},
	})
}

// SaveUser 用户名密码注册，成功后直接返回 token
func SaveUser(ctx context.Context, c *gin.Context, svc *models.ServiceContext) {
	request := new(LoginPswRequest)
	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数错误",
		})
		return
	}

	username := strings.TrimSpace(request.Username)
	if len(username) < 2 || len(username) > 50 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户名长度需在2-50个字符之间",
		})
		return
	}
	if len(request.Password) < 6 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "密码不能少于6位",
		})
		return
	}

	// 用户名查重
	_, err := dao.GetUserByUserName(svc.DB, username)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户名已存在",
		})
		return
	}
	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "查询用户失败",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "注册失败",
		})
		return
	}

	user := &dao.SysUser{
		UserName: username,
		Password: string(hash),
		NickName: username,
		RoleId:   1,
	}
	if err := user.Save(svc.DB); err != nil {
		logger.WithContext(c).Errorf("SaveUser db error: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "注册失败",
		})
		return
	}

	logger.WithContext(c).Infof("SaveUser success: %s", username)

	token, err := jwtx.GenerateToken(user.ID, user.UserName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "注册成功",
		"data": gin.H{
			"user_id":   user.ID,
			"user_name": user.UserName,
			"avatar":    user.Avatar,
			"token":     token,
		},
	})
}

// verifyAndUpgradePassword 校验密码。
// 存量用户密码为明文，为兼容旧数据：bcrypt 比对失败时回退明文比对，
// 明文比对成功则将密码原地升级为 bcrypt 哈希。
func verifyAndUpgradePassword(db *gorm.DB, user *dao.SysUser, password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
		return true
	}

	// 明文旧数据回退比对
	if user.Password == "" || user.Password != password {
		return false
	}

	if hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err == nil {
		if err := db.Model(user).Update("password", string(hash)).Error; err != nil {
			logger.Errorf("upgrade password hash failed for user %d: %v", user.ID, err)
		}
	}
	return true
}
