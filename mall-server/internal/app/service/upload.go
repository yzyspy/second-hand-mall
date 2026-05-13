package service

import (
	"fmt"
	"log"
	"mall-server/internal/app/config"
	"mall-server/pkg/logger"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// UploadTokenRequest 上传凭证请求
type UploadTokenRequest struct {
	Key string `json:"key" form:"key"` // 可选，指定文件名
}

// GetUploadToken 获取七牛云上传凭证
// 前端使用 uptoken 通过 wx.uploadFile 上传文件
// 参考文档: https://developer.qiniu.com/kodo/1312/upload
func GetUploadToken(c *gin.Context) {
	qiniuConf := config.C.Qiniu

	log.Println("GetUploadToken ", qiniuConf)
	logger.Infof("GetUploadToken %v", qiniuConf)

	if qiniuConf.AccessKey == "" || qiniuConf.SecretKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "七牛云配置缺失，请检查服务端配置",
		})
		return
	}

	// 读取前端传入的 key 参数
	var req UploadTokenRequest
	c.ShouldBind(&req)
	uploadKey := req.Key
	if uploadKey == "" {
		uploadKey = c.Query("key")
	}

	now := time.Now()

	// 生成完整的上传路径（服务端生成完整 key，确保唯一性）
	cosPathPrefix := fmt.Sprintf("upload/%s/", now.Format("20060102"))
	if uploadKey == "" {
		// 生成完整文件名：upload/20260423/123456_abc123.jpg
		uploadKey = fmt.Sprintf("%s%d_%s.jpg", cosPathPrefix, randInt(), randomString(6))
	}

	// 创建七牛云鉴权对象
	mac := qbox.NewMac(qiniuConf.AccessKey, qiniuConf.SecretKey)

	// 上传策略 - 有效期1小时
	putPolicy := storage.PutPolicy{
		Scope: qiniuConf.Bucket,
		// 可选：限制上传文件名
		// Scope: fmt.Sprintf("%s:%s", qiniuConf.Bucket, uploadKey),
	}
	putPolicy.Expires = 3600 // 1小时有效期

	// 生成上传凭证
	upToken := putPolicy.UploadToken(mac)

	log.Printf("GetUploadToken success, key: %s", uploadKey)

	// 返回数据
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"uploadKey": uploadKey,
			"upToken":   upToken,
			"domain":    qiniuConf.Domain,
			// 七牛云上传地址，华南区域
			"uploadUrl": "https://upload-z2.qiniup.com",
		},
	})
}

// randInt 生成一个随机数用于路径去重
func randInt() int {
	return rand.Intn(999999)
}

// randomString 生成指定长度的随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
