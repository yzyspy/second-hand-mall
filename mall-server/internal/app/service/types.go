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

// MyProductItem 我的商品列表项
type MyProductItem struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Images     string  `json:"images"`
	Status     int     `json:"status"`
	CreateTime string  `json:"create_time"`
}

// GetMyProductsRequest 获取我的商品列表请求
type GetMyProductsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	ID          uint     `json:"id" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Price       float64  `json:"price" binding:"required"`
	Location    string   `json:"location" binding:"required"`
	Images      []string `json:"images" binding:"required"`
}

// ChangeProductStatusRequest 变更商品状态请求
type ChangeProductStatusRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int  `json:"status" binding:"required"`
}
