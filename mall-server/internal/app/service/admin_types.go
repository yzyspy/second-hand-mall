package service

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminUserListRequest 管理员查询用户列表请求
type AdminUserListRequest struct {
	Keyword  string `form:"keyword"`
	IsBanned *bool  `form:"is_banned"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AdminProductListRequest 管理员查询商品列表请求
type AdminProductListRequest struct {
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Province string `form:"province"`
	City     string `form:"city"`
	District string `form:"district"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
