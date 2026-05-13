package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mall-server/internal/app/config"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/pkg/jwtx"
	"mall-server/pkg/logger"
)

// WxLogin 微信小程序登录
func WxLogin(ctx context.Context, c *gin.Context, svc *models.ServiceContext) {
	req := new(WxLoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数错误：code必填",
		})
		return
	}

	// 调用微信 code2Session 接口
	wxResp, err := code2Session(req.Code)
	if err != nil {
		logger.WithContext(c).Errorf("code2Session error: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "微信登录失败",
		})
		return
	}
	logger.WithContext(c).Infof("code2Session1 success: %v", wxResp)
	log.Println("code2Session2 success: %v", wxResp)
	logger.Infof("code2Session3 success: %v", wxResp)

	if wxResp.ErrCode != 0 {
		logger.WithContext(c).Errorf("code2Session errcode=%d errmsg=%s", wxResp.ErrCode, wxResp.ErrMsg)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "微信登录失败",
		})
		return
	}

	// 根据 openId 查找用户
	user, err := dao.GetUserByOpenID(svc.DB, wxResp.OpenID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.WithContext(c).Errorf("query user by openid error: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "查询用户失败",
		})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 新用户，创建记录
		user = &dao.SysUser{
			UserName:     fmt.Sprintf("wx_%s", wxResp.OpenID[:8]),
			WxOpenid:     wxResp.OpenID,
			WxUnionid:    wxResp.UnionID,
			WxSessionKey: wxResp.SessionKey,
			Avatar:       req.Avatar,
			NickName:     req.NickName,
		}
		if err := svc.DB.Create(user).Error; err != nil {
			logger.WithContext(c).Errorf("create user error: %v", err)
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "创建用户失败",
			})
			return
		}
	} else {
		// 老用户，更新信息
		updates := map[string]interface{}{
			"wx_session_key": wxResp.SessionKey,
		}
		if req.NickName != "" {
			updates["nick_name"] = req.NickName
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		if wxResp.UnionID != "" {
			updates["wx_unionid"] = wxResp.UnionID
		}
		if err := svc.DB.Model(user).Updates(updates).Error; err != nil {
			logger.WithContext(c).Errorf("update user error: %v", err)
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "更新用户失败",
			})
			return
		}
	}

	// 生成 JWT token
	token, err := jwtx.GenerateToken(user.ID, user.UserName)
	if err != nil {
		logger.WithContext(c).Errorf("generate token error: %v", err)
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
			"nick_name": user.NickName,
			"avatar":    user.Avatar,
			"token":     token,
		},
	})
}

// code2Session 调用微信 code2Session 接口
func code2Session(code string) (*WxCode2SessionResp, error) {
	cfg := config.C.WxApp
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.AppId, cfg.Secret, code,
	)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request weixin api error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %w", err)
	}

	var wxResp WxCode2SessionResp
	if err := json.Unmarshal(body, &wxResp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	return &wxResp, nil
}
