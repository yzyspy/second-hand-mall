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

// PublishProductRequest 发布商品请求
type PublishProductRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Price       float64  `json:"price" binding:"required"`
	Location    string   `json:"location" binding:"required"`
	Category    string   `json:"category"`
	Images      []string `json:"images" binding:"required"`
}

// ProductDetailRequest 商品详情请求
type ProductDetailRequest struct {
	ID uint `form:"id" binding:"required"`
}
