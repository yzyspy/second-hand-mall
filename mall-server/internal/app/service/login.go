package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/pkg/logger"
	"net/http"
	"time"
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

	// 验证密码
	if user.Password != request.Password {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "密码错误",
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
		},
	})
}

func SaveUser(ctx context.Context, ginCtx *gin.Context, svc *models.ServiceContext) string {
	request := new(LoginPswRequest)
	// 绑定 POST 请求中的 JSON 数据到 request 对象
	if err := ginCtx.ShouldBindJSON(request); err != nil {
		return fmt.Sprintf("bind request error: %v", err)
	}
	//fmt.Printf("SaveUser request:%s, %s\n", request.Username, request.Password)
	log.Printf("SaveUser request:%s, %s\n", request.Username, request.Password)
	logger.Errorf("保存用户成功：%s\n", request.Username)
	user := newUser(request.Username, request.Password)
	err2 := user.Save(svc.DB)

	if err2 != nil {
		panic(fmt.Sprintf("db save error: %v", err2))
	}

	return "save user success " + request.Username
}

// MockSysUser 返回一个填充了模拟数据的 SysUser 指针
func newUser(username, password string) *dao.SysUser {
	return &dao.SysUser{
		Model: gorm.Model{
			//		ID:        2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{Valid: false}, // 未删除
		},
		UserName: username,
		Password: password, // 模拟密码，实际应加密
		Phone:    "13800138000",
		WxUserid: "wx_user_123",
		WxOpenid: "openid_123",
		Avatar:   "https://example.com/avatar.jpg",
		Sex:      "male",
		Email:    "test@example.com",
		Remarks:  "这是一个用于测试的模拟用户",
		RoleId:   1,
	}
}
