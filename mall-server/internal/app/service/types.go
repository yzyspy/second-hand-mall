package service

type LoginPswRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// WxLoginRequest 微信登录请求
type WxLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	NickName string `json:"nick_name"`
	Avatar   string `json:"avatar"`
}

// WxCode2SessionResp 微信code2Session接口响应
type WxCode2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// SearchProductRequest 商品搜索请求
type SearchProductRequest struct {
	Keyword  string `form:"keyword"`
	Sort     string `form:"sort"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
