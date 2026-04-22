package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"mall-server/pkg/logger"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"mall-server/internal/app/config"
)

// CosSignatureResponse COS 签名响应
type CosSignatureResponse struct {
	TmpSecretId  string `json:"tmpSecretId"`
	TmpSecretKey string `json:"tmpSecretKey"`
	SessionToken string `json:"sessionToken"`
	StartTime    int64  `json:"startTime"`
	ExpiredTime  int64  `json:"expiredTime"`
	Signature    string `json:"signature"`
	KeyTime      string `json:"keyTime"`
}

// CosSignatureRequest COS 签名请求
type CosSignatureRequest struct {
	Key string `json:"key" form:"key"`
}

// GetCosSignature 生成腾讯云 COS 前端直传签名
// COS 签名算法参考: https://cloud.tencent.com/document/product/436/14690
// 前端直传使用 POST Object 接口，需要生成 policy 签名
func GetCosSignature(c *gin.Context) {
	cosConf := config.C.Cos

	log.Println("GetCosSignature 1 ", cosConf)
	logger.Infof("GetCosSignature 2 %v", cosConf)

	if cosConf.SecretId == "" || cosConf.SecretKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "COS 配置缺失，请检查服务端配置",
		})
		return
	}

	// 读取前端传入的 key 参数（支持 JSON body 和 form-data）
	var req CosSignatureRequest
	c.ShouldBind(&req)
	uploadKey := req.Key
	if uploadKey == "" {
		uploadKey = c.Query("key")
	}

	now := time.Now()
	expireSeconds := cosConf.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = 600 // 默认 10 分钟
	}

	startTime := now.Unix()
	expiredTime := now.Add(time.Duration(expireSeconds) * time.Second).Unix()
	keyTime := fmt.Sprintf("%d;%d", startTime, expiredTime)

	// 生成建议的上传路径前缀（前端上传的key必须以此开头）
	cosPathPrefix := fmt.Sprintf("upload/%s/", now.Format("20060102"))

	// 如果前端传了 key，使用该 key 作为 starts-with 的匹配前缀
	// 否则使用默认的日期目录前缀
	startsWithPrefix := cosPathPrefix
	cosPath := cosPathPrefix + fmt.Sprintf("%d_", randInt())

	if uploadKey != "" {
		startsWithPrefix = uploadKey
		//cosPath = uploadKey
	}
	log.Printf("cosPath: %s", cosPath)
	log.Printf("uploadKey: %s", uploadKey)

	// 1. 生成 SignKey = HMAC-SHA1(SecretKey, KeyTime)
	signKey := hmacSha1(cosConf.SecretKey, keyTime)

	// 2. 生成 policy（POST Object 所需的策略）
	// expiration 使用 ISO8601 格式
	// starts-with 使用宽松的前缀匹配，允许前端在该目录下上传任意文件
	expiration := now.Add(time.Duration(expireSeconds) * time.Second).UTC().Format("2006-01-02T15:04:05.000Z")
	policy := fmt.Sprintf(`{"expiration":"%s","conditions":[{"bucket":"%s"},["starts-with","$key","%s"],["eq","$q-ak","%s"],["eq","$q-sign-time","%s"],["eq","$q-sign-algorithm","sha1"]]}`,
		expiration, cosConf.Bucket, startsWithPrefix, cosConf.SecretId, keyTime)

	// 3. policy 需要 base64 编码
	policyBase64 := base64.StdEncoding.EncodeToString([]byte(policy))

	// 4. 生成 StringToSign = SHA1(policy_base64)
	stringToSign := sha1Hex(policyBase64)

	// 5. 生成 Signature = HMAC-SHA1(SignKey, StringToSign)
	signature := hmacSha1(signKey, stringToSign)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"cosHost":          fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosConf.Bucket, cosConf.Region),
			"secretId":         cosConf.SecretId,
			"keyTime":          keyTime,
			"signature":        signature,
			"q-sign-algorithm": "sha1",
			"policy":           policyBase64,
			"startTime":        startTime,
			"expiredTime":      expiredTime,
			"sessionToken":     "",
			"cosPathPrefix":    cosPathPrefix,
			"cosPath":          cosPath,
		},
	})
}

// GetCosSignatureV2 使用腾讯云 STS 临时密钥生成 COS 上传签名
// 前端使用返回的 tmpSecretId + tmpSecretKey + sessionToken 直接上传文件到 COS
// 参考文档: https://cloud.tencent.com/document/product/436/14048
func GetCosSignatureV2(c *gin.Context) {
	cosConf := config.C.Cos

	log.Println("GetCosSignatureV2 ", cosConf)
	logger.Infof("GetCosSignatureV2 %v", cosConf)

	if cosConf.SecretId == "" || cosConf.SecretKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "COS 配置缺失，请检查服务端配置",
		})
		return
	}

	// 读取前端传入的 key 参数（支持 JSON body 和 form-data）
	var req CosSignatureRequest
	c.ShouldBind(&req)
	uploadKey := req.Key
	if uploadKey == "" {
		uploadKey = c.Query("key")
	}

	// 生成建议的上传路径
	now := time.Now()
	cosPathPrefix := fmt.Sprintf("upload/%s/", now.Format("20060102"))
	cosPath := cosPathPrefix + fmt.Sprintf("%d_", randInt())
	if uploadKey != "" {
		cosPath = uploadKey
	}

	expireSeconds := cosConf.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = 600 // 默认 10 分钟
	}

	// 创建 STS 客户端
	client := sts.NewClient(
		cosConf.SecretId,
		cosConf.SecretKey,
		nil,
	)

	// 构建 CAM 策略，限制只能上传指定路径
	// Resource 格式: qcs::cos:<region>:uid/<appid>:<bucket>/<path>
	resourcePrefix := fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s*",
		cosConf.Region, cosConf.AppId, cosConf.Bucket, cosPathPrefix)
	if uploadKey != "" {
		resourcePrefix = fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s",
			cosConf.Region, cosConf.AppId, cosConf.Bucket, uploadKey)
	}

	opt := &sts.CredentialOptions{
		DurationSeconds: int64(expireSeconds),
		Region:          cosConf.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action: []string{
						"name/cos:PostObject",
						"name/cos:PutObject",
					},
					Effect:   "allow",
					Resource: []string{resourcePrefix},
				},
			},
		},
	}

	// 请求临时密钥
	res, err := client.GetCredential(opt)
	if err != nil {
		logger.Errorf("获取 STS 临时密钥失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取临时密钥失败: " + err.Error(),
		})
		return
	}

	log.Printf("GetCosSignatureV2 success, cosPath: %s", cosPath)
	log.Printf("uploadKey: %s", uploadKey)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"cosHost":      fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosConf.Bucket, cosConf.Region),
			"tmpSecretId":  res.Credentials.TmpSecretID,
			"tmpSecretKey": res.Credentials.TmpSecretKey,
			"sessionToken": res.Credentials.SessionToken,
			//"expiredTime":  res.Credentials.ExpiredTime,
			//"startTime":    res.Credentials.StartTime,
			"region":        cosConf.Region,
			"bucket":        cosConf.Bucket,
			"cosPath":       cosPath,
			"cosPathPrefix": cosPathPrefix,
		},
	})
}

// hmacSha1 计算 HMAC-SHA1
func hmacSha1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// sha1Hex 计算 SHA1 的十六进制字符串
func sha1Hex(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// randInt 生成一个随机数用于路径去重
func randInt() int {
	return rand.Intn(999999)
}
