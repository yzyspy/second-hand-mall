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

func LoginPsw(c *gin.Context) {
	request := new(LoginPswRequest)

	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
	}

	fmt.Printf("LoginPsw request:%s, %s\n", request.Username, request.Password)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "login success",
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
	err2 := MockSysUser().Save(svc.DB)

	if err2 != nil {
		panic(fmt.Sprintf("db save error: %v", err2))
	}

	return "save user success " + request.Username
}

// MockSysUser 返回一个填充了模拟数据的 SysUser 指针
func MockSysUser() *dao.SysUser {
	return &dao.SysUser{
		Model: gorm.Model{
			ID:        2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{Valid: false}, // 未删除
		},
		UserName: "songmeiling",
		Password: "Test@1234", // 模拟密码，实际应加密
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
