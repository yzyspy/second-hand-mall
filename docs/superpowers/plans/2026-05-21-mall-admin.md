# 二手商城管理后台 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `mall-admin-web` Web 管理后台（Vue3+Vite+TS+Element Plus）及 `mall-server` 对应的 `/admin` 前缀 API，支持管理员对用户（列表/详情/封禁/解封）和商品（列表/详情/强制下架）进行管理。

**Architecture:** 后端新增独立的 `admin_user` 表和 Admin JWT（7天有效期，claims 含 `is_admin: true`），`AdminAuthMiddleware` 独立验证；`sys_user` 新增 `is_banned` 字段，两个登录入口均检查该字段；第一个管理员账号通过 CLI 命令 `./mall-server admin create-admin` 创建（bcrypt 哈希密码）。前端为独立 `mall-admin-web/` 项目，Vite proxy 将 `/admin/*` 转发到后端 10088 端口。

**Tech Stack:** Go/GORM/Gin/bcrypt（后端），Vue3/Vite/TypeScript/Element Plus/Pinia/Axios（前端），SQLite（数据库）

---

## 工作目录

所有后端步骤在 `mall-server/` 下执行；前端步骤在项目根目录执行（`mall-admin-web/` 将被创建在此）。

---

### Task 1: 后端 — AdminUser 实体 + is_banned + AutoMigrate

**Files:**
- Create: `mall-server/internal/app/dao/admin.entity.go`
- Modify: `mall-server/internal/app/dao/user.entity.go`
- Modify: `mall-server/internal/app/models/init.go`
- Test: `mall-server/internal/app/dao/admin_test.go`

- [ ] **Step 1: 新增 `admin.entity.go`**

```go
package dao

import "gorm.io/gorm"

type AdminUser struct {
	gorm.Model
	Username     string `gorm:"column:username;type:varchar(50);not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"       json:"-"`
}

func (AdminUser) TableName() string {
	return "admin_user"
}
```

- [ ] **Step 2: 在 `user.entity.go` 的 SysUser 中追加 IsBanned 字段**

在 `NickName` 字段之后追加一行：

```go
IsBanned bool `gorm:"column:is_banned;type:boolean;not null;default:false" json:"is_banned" comment:"是否封禁"`
```

SysUser 完整字段列表变为（只展示新行位置）：
```go
NickName  string `gorm:"column:nick_name;type:varchar(50);not null;default:''" json:"nick_name" comment:"昵称"`
IsBanned  bool   `gorm:"column:is_banned;type:boolean;not null;default:false"  json:"is_banned" comment:"是否封禁"`
```

- [ ] **Step 3: 在 `models/init.go` 的 `NewDB()` 末尾追加 AdminUser AutoMigrate**

在收藏表 AutoMigrate 之后添加：

```go
// 自动迁移管理员表
if err := con.AutoMigrate(&dao.AdminUser{}); err != nil {
    panic(fmt.Sprintf("db auto migrate error: %v", err))
}
```

- [ ] **Step 4: 写 `dao/admin_test.go` 并验证 AutoMigrate 可正常建表**

```go
package dao_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
)

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&dao.AdminUser{}, &dao.SysUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAdminUserCreate(t *testing.T) {
	db := newTestDB(t)
	admin := dao.AdminUser{Username: "admin", PasswordHash: "hash"}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if admin.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestAdminUserUsernameUnique(t *testing.T) {
	db := newTestDB(t)
	db.Create(&dao.AdminUser{Username: "dup", PasswordHash: "h1"})
	err := db.Create(&dao.AdminUser{Username: "dup", PasswordHash: "h2"}).Error
	if err == nil {
		t.Fatal("expected unique constraint violation")
	}
}

func TestSysUserIsBanned(t *testing.T) {
	db := newTestDB(t)
	user := dao.SysUser{UserName: "testuser", Password: "pw"}
	db.Create(&user)
	db.Model(&user).Update("is_banned", true)

	var result dao.SysUser
	db.First(&result, user.ID)
	if !result.IsBanned {
		t.Fatal("expected is_banned=true")
	}
}
```

- [ ] **Step 5: 运行测试**

```bash
cd mall-server && go test ./internal/app/dao/... -run TestAdminUser -v
cd mall-server && go test ./internal/app/dao/... -run TestSysUserIsBanned -v
```

Expected: 3 tests PASS

- [ ] **Step 6: Commit**

```bash
git add mall-server/internal/app/dao/admin.entity.go \
        mall-server/internal/app/dao/user.entity.go \
        mall-server/internal/app/dao/admin_test.go \
        mall-server/internal/app/models/init.go
git commit -m "feat: add admin_user table and is_banned field to sys_user"
```

---

### Task 2: 后端 — Admin JWT + AdminAuthMiddleware

**Files:**
- Modify: `mall-server/pkg/jwtx/jwtx.go`
- Modify: `mall-server/internal/app/router/auth.go`
- Test: `mall-server/pkg/jwtx/jwtx_test.go`（如不存在则新建）

- [ ] **Step 1: 在 `pkg/jwtx/jwtx.go` 末尾追加 Admin JWT 函数**

```go
// AdminClaims 管理员 JWT Claims
type AdminClaims struct {
	AdminID  uint   `json:"admin_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenerateAdminToken 生成管理员 JWT（7天有效期）
func GenerateAdminToken(adminID uint, username string) (string, error) {
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		IsAdmin:  true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mall-server-admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseAdminToken 解析并验证管理员 JWT
func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid || !claims.IsAdmin {
		return nil, errors.New("invalid admin token")
	}
	return claims, nil
}
```

- [ ] **Step 2: 写 JWT 测试（在 `pkg/jwtx/jwtx_test.go` 中追加）**

如文件不存在则新建，如已存在则追加到末尾：

```go
package jwtx_test

import (
	"testing"
	"mall-server/pkg/jwtx"
)

func TestGenerateAndParseAdminToken(t *testing.T) {
	token, err := jwtx.GenerateAdminToken(1, "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := jwtx.ParseAdminToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.AdminID != 1 || claims.Username != "admin" || !claims.IsAdmin {
		t.Fatalf("wrong claims: %+v", claims)
	}
}

func TestParseAdminToken_RejectUserToken(t *testing.T) {
	// 普通用户 token 不能被 ParseAdminToken 接受
	userToken, _ := jwtx.GenerateToken(1, "user")
	_, err := jwtx.ParseAdminToken(userToken)
	if err == nil {
		t.Fatal("expected error for user token")
	}
}
```

- [ ] **Step 3: 运行 JWT 测试**

```bash
cd mall-server && go test ./pkg/jwtx/... -v
```

Expected: PASS

- [ ] **Step 4: 在 `router/auth.go` 末尾追加 AdminAuthMiddleware**

```go
// AdminAuthMiddleware 管理员 JWT 鉴权中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "未登录"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "token 格式错误"})
			return
		}
		claims, err := jwtx.ParseAdminToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "token 无效或已过期"})
			return
		}
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_name", claims.Username)
		c.Next()
	}
}
```

- [ ] **Step 5: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误

- [ ] **Step 6: Commit**

```bash
git add mall-server/pkg/jwtx/jwtx.go \
        mall-server/pkg/jwtx/jwtx_test.go \
        mall-server/internal/app/router/auth.go
git commit -m "feat: add admin JWT generation and AdminAuthMiddleware"
```

---

### Task 3: 后端 — Admin 登录接口

**Files:**
- Create: `mall-server/internal/app/service/admin_types.go`
- Create: `mall-server/internal/app/service/admin_auth.go`
- Create: `mall-server/internal/app/dao/admin.repo.go`

- [ ] **Step 1: 新增 `dao/admin.repo.go`**

```go
package dao

import (
	"gorm.io/gorm"
)

// GetAdminByUsername 根据用户名查询管理员
func GetAdminByUsername(db *gorm.DB, username string) (*AdminUser, error) {
	var admin AdminUser
	err := db.Where("username = ?", username).First(&admin).Error
	return &admin, err
}
```

- [ ] **Step 2: 新增 `service/admin_types.go`**

```go
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
```

- [ ] **Step 3: 新增 `service/admin_auth.go`**

```go
package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/pkg/jwtx"
)

// AdminLogin POST /admin/login
func AdminLogin(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		admin, err := dao.GetAdminByUsername(svc.DB, req.Username)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户名或密码错误"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户名或密码错误"})
			return
		}

		token, err := jwtx.GenerateAdminToken(admin.ID, admin.Username)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "生成 token 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "登录成功",
			"data": gin.H{"token": token, "username": admin.Username},
		})
	}
}
```

- [ ] **Step 4: 添加 bcrypt 依赖**

```bash
cd mall-server && go get golang.org/x/crypto/bcrypt
```

- [ ] **Step 5: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误

- [ ] **Step 6: Commit**

```bash
git add mall-server/internal/app/dao/admin.repo.go \
        mall-server/internal/app/service/admin_types.go \
        mall-server/internal/app/service/admin_auth.go \
        mall-server/go.mod mall-server/go.sum
git commit -m "feat: add admin login handler with bcrypt password verification"
```

---

### Task 4: 后端 — Admin 用户管理接口

**Files:**
- Modify: `mall-server/internal/app/dao/admin.repo.go`
- Create: `mall-server/internal/app/service/admin_user.go`

- [ ] **Step 1: 在 `dao/admin.repo.go` 追加用户管理查询函数**

```go
package dao

import (
	"gorm.io/gorm"
)

// GetAdminByUsername 根据用户名查询管理员
func GetAdminByUsername(db *gorm.DB, username string) (*AdminUser, error) {
	var admin AdminUser
	err := db.Where("username = ?", username).First(&admin).Error
	return &admin, err
}

// AdminListUsersResult 用户列表查询结果
type AdminListUsersResult struct {
	ID           uint   `json:"id"`
	UserName     string `json:"user_name"`
	NickName     string `json:"nick_name"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	IsBanned     bool   `json:"is_banned"`
	ProductCount int64  `json:"product_count"`
	CreatedAt    string `json:"created_at"`
}

// AdminListUsers 分页查询用户列表
func AdminListUsers(db *gorm.DB, keyword string, isBanned *bool, page, pageSize int) ([]AdminListUsersResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&SysUser{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR nick_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if isBanned != nil {
		query = query.Where("is_banned = ?", *isBanned)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []SysUser
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	results := make([]AdminListUsersResult, 0, len(users))
	for _, u := range users {
		var count int64
		db.Model(&Product{}).Where("user_id = ? AND deleted_at IS NULL", u.ID).Count(&count)
		results = append(results, AdminListUsersResult{
			ID:           u.ID,
			UserName:     u.UserName,
			NickName:     u.NickName,
			Phone:        u.Phone,
			Avatar:       u.Avatar,
			IsBanned:     u.IsBanned,
			ProductCount: count,
			CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return results, total, nil
}

// AdminGetUserDetail 获取用户详情（含商品数、收藏数）
func AdminGetUserDetail(db *gorm.DB, id uint) (*AdminListUsersResult, error) {
	var u SysUser
	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}
	var productCount, favoriteCount int64
	db.Model(&Product{}).Where("user_id = ? AND deleted_at IS NULL", id).Count(&productCount)
	db.Model(&UserFavorite{}).Where("user_id = ?", id).Count(&favoriteCount)

	return &AdminListUsersResult{
		ID:           u.ID,
		UserName:     u.UserName,
		NickName:     u.NickName,
		Phone:        u.Phone,
		Avatar:       u.Avatar,
		IsBanned:     u.IsBanned,
		ProductCount: productCount,
		CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminSetUserBanned 设置用户封禁状态
func AdminSetUserBanned(db *gorm.DB, id uint, banned bool) error {
	return db.Model(&SysUser{}).Where("id = ?", id).Update("is_banned", banned).Error
}
```

Note: replace the previous `admin.repo.go` (which only had `GetAdminByUsername`) with the above complete version.

- [ ] **Step 2: 新增 `service/admin_user.go`**

```go
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// AdminListUsers GET /admin/users
func AdminListUsers(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminUserListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 50 {
			req.PageSize = 10
		}

		results, total, err := dao.AdminListUsers(svc.DB, req.Keyword, req.IsBanned, req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0, "msg": "success",
			"data": gin.H{"list": results, "total": total, "page": req.Page, "page_size": req.PageSize},
		})
	}
}

// AdminGetUserDetail GET /admin/users/:id
func AdminGetUserDetail(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		detail, err := dao.AdminGetUserDetail(svc.DB, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": detail})
	}
}

// AdminBanUser POST /admin/users/:id/ban
func AdminBanUser(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if err := dao.AdminSetUserBanned(svc.DB, uint(id), true); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "封禁成功"})
	}
}

// AdminUnbanUser POST /admin/users/:id/unban
func AdminUnbanUser(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if err := dao.AdminSetUserBanned(svc.DB, uint(id), false); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "解封成功"})
	}
}
```

- [ ] **Step 3: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add mall-server/internal/app/dao/admin.repo.go \
        mall-server/internal/app/service/admin_user.go
git commit -m "feat: add admin user management handlers (list/detail/ban/unban)"
```

---

### Task 5: 后端 — Admin 商品管理接口

**Files:**
- Create: `mall-server/internal/app/service/admin_product.go`

- [ ] **Step 1: 新增 `service/admin_product.go`**

```go
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// AdminListProducts GET /admin/products
func AdminListProducts(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminProductListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 50 {
			req.PageSize = 10
		}

		results, total, err := dao.SearchProducts(
			svc.DB,
			req.Keyword, "time_desc", req.Status,
			req.Category, req.Province, req.City, req.District,
			req.Page, req.PageSize,
		)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0, "msg": "success",
			"data": gin.H{"list": results, "total": total, "page": req.Page, "page_size": req.PageSize},
		})
	}
}

// AdminGetProductDetail GET /admin/products/:id
func AdminGetProductDetail(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		detail, err := dao.GetProductByID(svc.DB, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "商品不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": detail})
	}
}

// AdminDelistProduct POST /admin/products/:id/delist
func AdminDelistProduct(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}
		// 直接设置 status=2，不校验所有者（管理员操作）
		if err := svc.DB.Model(&dao.Product{}).Where("id = ?", id).Update("status", 2).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已下架"})
	}
}
```

- [ ] **Step 2: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add mall-server/internal/app/service/admin_product.go
git commit -m "feat: add admin product management handlers (list/detail/delist)"
```

---

### Task 6: 后端 — 路由注册 + CORS + is_banned 登录检查 + create-admin CLI

**Files:**
- Modify: `mall-server/internal/app/router/router.go`
- Modify: `mall-server/internal/app/service/login.go`
- Modify: `mall-server/main.go`

- [ ] **Step 1: 修改 `router.go` 的 CORSMiddleware，允许 5174 端口**

将 `CORSMiddleware` 中的 Allow-Origin 行替换为：

```go
origin := c.Request.Header.Get("Origin")
allowed := map[string]bool{
    "http://localhost:5173": true,
    "http://localhost:5174": true,
}
if allowed[origin] {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
} else {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
}
```

完整 `CORSMiddleware` 变为：

```go
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := map[string]bool{
			"http://localhost:5173": true,
			"http://localhost:5174": true,
		}
		if allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: 在 `router.go` 的 `App` 函数中注册 admin 路由**

在 `return r` 之前追加：

```go
// ========== Admin 接口（独立鉴权）==========
r.POST("/admin/login", service.AdminLogin(svc))

adminGroup := r.Group("/admin")
adminGroup.Use(AdminAuthMiddleware())
{
    adminGroup.GET("/users", service.AdminListUsers(svc))
    adminGroup.GET("/users/:id", service.AdminGetUserDetail(svc))
    adminGroup.POST("/users/:id/ban", service.AdminBanUser(svc))
    adminGroup.POST("/users/:id/unban", service.AdminUnbanUser(svc))

    adminGroup.GET("/products", service.AdminListProducts(svc))
    adminGroup.GET("/products/:id", service.AdminGetProductDetail(svc))
    adminGroup.POST("/products/:id/delist", service.AdminDelistProduct(svc))
}
```

- [ ] **Step 3: 在 `service/login.go` 的 `LoginPsw` 中添加 is_banned 检查**

在密码验证通过之后、生成 token 之前插入：

```go
// 检查用户是否被封禁
if user.IsBanned {
    c.JSON(http.StatusForbidden, gin.H{
        "code": -1,
        "msg":  "账号已被封禁，请联系管理员",
    })
    return
}
```

- [ ] **Step 4: 在 `service/login.go` 的 `WxLogin` 中添加 is_banned 检查**

找到 `WxLogin` 中查询到用户之后生成 token 之前的位置，插入同样的检查：

```go
// 检查用户是否被封禁
if user.IsBanned {
    c.JSON(http.StatusForbidden, gin.H{
        "code": -1,
        "msg":  "账号已被封禁，请联系管理员",
    })
    return
}
```

- [ ] **Step 5: 在 `main.go` 中追加 `admin` CLI 命令**

在 `newWebCmd` 函数之后添加：

```go
func newAdminCmd(ctx context.Context) *cli.Command {
	return &cli.Command{
		Name:  "admin",
		Usage: "Admin management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "create-admin",
				Usage: "Create a new admin user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "Config file", Required: true},
					&cli.StringFlag{Name: "username", Aliases: []string{"u"}, Usage: "Admin username", Required: true},
					&cli.StringFlag{Name: "password", Aliases: []string{"p"}, Usage: "Admin password", Required: true},
				},
				Action: func(c *cli.Context) error {
					if err := app.Run(ctx, app.SetConfigFile(c.String("config"))); err != nil {
						return err
					}
					svc := models.NewServiceContext(ctx)
					hash, err := bcrypt.GenerateFromPassword([]byte(c.String("password")), bcrypt.DefaultCost)
					if err != nil {
						return fmt.Errorf("hash password: %w", err)
					}
					admin := dao.AdminUser{
						Username:     c.String("username"),
						PasswordHash: string(hash),
					}
					if err := svc.DB.Create(&admin).Error; err != nil {
						return fmt.Errorf("create admin: %w", err)
					}
					log.Printf("管理员 %q 创建成功 (ID=%d)\n", admin.Username, admin.ID)
					return nil
				},
			},
		},
	}
}
```

在 `main()` 中将 `newAdminCmd` 加入 Commands 列表：

```go
app.Commands = []*cli.Command{
    newWebCmd(ctx),
    newAdminCmd(ctx),
}
```

在 `main.go` 顶部 import 中追加：

```go
"fmt"
"golang.org/x/crypto/bcrypt"
"mall-server/internal/app/dao"
```

- [ ] **Step 6: 编译并验证 create-admin 命令存在**

```bash
cd mall-server && go build -o mall-server-bin . && ./mall-server-bin admin create-admin --help
```

Expected: 输出显示 `--username`, `--password`, `--config` flags

- [ ] **Step 7: 运行全量测试**

```bash
cd mall-server && go test ./...
```

Expected: 所有测试 PASS

- [ ] **Step 8: Commit**

```bash
git add mall-server/internal/app/router/router.go \
        mall-server/internal/app/service/login.go \
        mall-server/main.go
git commit -m "feat: register admin routes, fix CORS, add is_banned login check, add create-admin CLI"
```

---

### Task 7: 前端脚手架 + Axios 封装 + Pinia Auth Store

**Files:**
- Create: `mall-admin-web/` (整个目录)
- Create: `mall-admin-web/src/utils/request.ts`
- Create: `mall-admin-web/src/stores/auth.ts`

- [ ] **Step 1: 初始化 Vue3+Vite+TS 项目**

在项目根目录执行：

```bash
cd /path/to/second-hand-mall
npm create vite@latest mall-admin-web -- --template vue-ts
cd mall-admin-web
npm install
```

- [ ] **Step 2: 安装依赖**

```bash
cd mall-admin-web
npm install element-plus @element-plus/icons-vue
npm install pinia vue-router axios
npm install -D unplugin-vue-components unplugin-auto-import
```

- [ ] **Step 3: 替换 `vite.config.ts`**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({ resolvers: [ElementPlusResolver()] }),
    Components({ resolvers: [ElementPlusResolver()] }),
  ],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5174,
    proxy: {
      '/admin': {
        target: 'http://localhost:10088',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 4: 新建 `src/utils/request.ts`**

```typescript
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const request = axios.create({ timeout: 10000 })

request.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

request.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      router.push('/login')
    }
    return Promise.reject(err)
  }
)

export default request
```

- [ ] **Step 5: 新建 `src/stores/auth.ts`**

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('admin_token') ?? '')
  const username = ref<string>(localStorage.getItem('admin_username') ?? '')

  function setAuth(t: string, u: string) {
    token.value = t
    username.value = u
    localStorage.setItem('admin_token', t)
    localStorage.setItem('admin_username', u)
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_username')
  }

  return { token, username, setAuth, logout }
})
```

- [ ] **Step 6: 新建 `src/api/auth.ts`**

```typescript
import request from '@/utils/request'

export function adminLogin(username: string, password: string) {
  return request.post<any, { code: number; msg: string; data: { token: string; username: string } }>(
    '/admin/login',
    { username, password }
  )
}
```

- [ ] **Step 7: 新建 `src/api/users.ts`**

```typescript
import request from '@/utils/request'

export interface AdminUser {
  id: number
  user_name: string
  nick_name: string
  phone: string
  avatar: string
  is_banned: boolean
  product_count: number
  created_at: string
}

export interface UserListParams {
  keyword?: string
  is_banned?: boolean
  page: number
  page_size: number
}

export function getUserList(params: UserListParams) {
  return request.get<any, { code: number; data: { list: AdminUser[]; total: number } }>(
    '/admin/users', { params }
  )
}

export function getUserDetail(id: number) {
  return request.get<any, { code: number; data: AdminUser & { favorite_count?: number } }>(
    `/admin/users/${id}`
  )
}

export function banUser(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/users/${id}/ban`)
}

export function unbanUser(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/users/${id}/unban`)
}
```

- [ ] **Step 8: 新建 `src/api/products.ts`**

```typescript
import request from '@/utils/request'

export interface AdminProduct {
  id: number
  title: string
  price: number
  images: string
  location: string
  status: number
  category: string
  seller: string
  create_time: string
}

export interface ProductListParams {
  keyword?: string
  category?: string
  province?: string
  status?: number
  page: number
  page_size: number
}

export function getProductList(params: ProductListParams) {
  return request.get<any, { code: number; data: { list: AdminProduct[]; total: number } }>(
    '/admin/products', { params }
  )
}

export function getProductDetail(id: number) {
  return request.get<any, { code: number; data: AdminProduct }>(`/admin/products/${id}`)
}

export function delistProduct(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/products/${id}/delist`)
}
```

- [ ] **Step 9: 更新 `src/main.ts`**

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import 'element-plus/dist/index.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

- [ ] **Step 10: 验证项目能启动**

```bash
cd mall-admin-web && npm run dev
```

Expected: 控制台输出 `Local: http://localhost:5174/`，无编译错误

- [ ] **Step 11: Commit**

```bash
git add mall-admin-web/
git commit -m "feat: scaffold mall-admin-web with Vue3+Vite+TS+Element Plus, request util and auth store"
```

---

### Task 8: 前端 — 登录页

**Files:**
- Create: `mall-admin-web/src/pages/Login.vue`
- Create: `mall-admin-web/src/router/index.ts`
- Modify: `mall-admin-web/src/App.vue`

- [ ] **Step 1: 新建 `src/router/index.ts`（仅 Login 路由，后续任务补充其他路由）**

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/pages/Login.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/login' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path !== '/login' && !auth.token) {
    return '/login'
  }
})

export default router
```

- [ ] **Step 2: 更新 `src/App.vue`**

```vue
<template>
  <router-view />
</template>
```

- [ ] **Step 3: 新建 `src/pages/Login.vue`**

```vue
<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">二手商城管理后台</h2>
      <el-form :model="form" :rules="rules" ref="formRef" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码"
            prefix-icon="Lock" show-password @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { adminLogin } from '@/api/auth'

const router = useRouter()
const auth = useAuthStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  await formRef.value?.validate()
  loading.value = true
  try {
    const res = await adminLogin(form.username, form.password)
    if (res.code === 0) {
      auth.setAuth(res.data.token, res.data.username)
      router.push('/users')
    } else {
      ElMessage.error(res.msg || '登录失败')
    }
  } catch {
    ElMessage.error('网络错误，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f1f5f9;
}
.login-card {
  width: 380px;
}
.title {
  text-align: center;
  margin-bottom: 24px;
  font-size: 20px;
  color: #1d2d44;
}
.login-btn {
  width: 100%;
}
</style>
```

- [ ] **Step 4: 验证登录页渲染**

```bash
cd mall-admin-web && npm run dev
```

访问 http://localhost:5174/login，应显示登录表单，无控制台错误

- [ ] **Step 5: Commit**

```bash
git add mall-admin-web/src/router/index.ts \
        mall-admin-web/src/App.vue \
        mall-admin-web/src/pages/Login.vue
git commit -m "feat: add admin login page with form validation"
```

---

### Task 9: 前端 — AdminLayout + 路由完善

**Files:**
- Create: `mall-admin-web/src/layouts/AdminLayout.vue`
- Modify: `mall-admin-web/src/router/index.ts`

- [ ] **Step 1: 新建 `src/layouts/AdminLayout.vue`**

```vue
<template>
  <div class="admin-layout">
    <!-- 左侧导航 -->
    <aside class="sidebar">
      <div class="logo">⚙ 二手商城</div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#1d2d44"
        text-color="#94a3b8"
        active-text-color="#fff"
      >
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><Box /></el-icon>
          <span>商品管理</span>
        </el-menu-item>
      </el-menu>
      <div class="logout" @click="handleLogout">
        <el-icon><SwitchButton /></el-icon> 退出登录
      </div>
    </aside>

    <!-- 右侧主区域 -->
    <div class="main">
      <header class="header">
        <span class="breadcrumb">{{ currentTitle }}</span>
        <span class="admin-name">{{ auth.username }}</span>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => {
  if (route.path.startsWith('/users')) return '/users'
  if (route.path.startsWith('/products')) return '/products'
  return '/users'
})

const titleMap: Record<string, string> = {
  '/users': '用户管理 / 用户列表',
  '/products': '商品管理 / 商品列表',
}

const currentTitle = computed(() => titleMap[route.path] ?? '管理后台')

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
}
.sidebar {
  width: 200px;
  background: #1d2d44;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.logo {
  padding: 16px;
  color: #fff;
  font-size: 15px;
  font-weight: bold;
  border-bottom: 1px solid #2d4060;
}
.logout {
  margin-top: auto;
  padding: 14px 20px;
  color: #64748b;
  cursor: pointer;
  border-top: 1px solid #2d4060;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.logout:hover { color: #94a3b8; }
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f1f5f9;
  overflow: hidden;
}
.header {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  padding: 0 24px;
  justify-content: space-between;
  font-size: 13px;
  color: #64748b;
}
.admin-name { font-weight: 500; color: #334155; }
.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>
```

- [ ] **Step 2: 更新 `src/router/index.ts`，补充所有路由**

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AdminLayout from '@/layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/pages/Login.vue') },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/users',
      children: [
        { path: 'users',        component: () => import('@/pages/users/UserList.vue') },
        { path: 'users/:id',    component: () => import('@/pages/users/UserDetail.vue') },
        { path: 'products',     component: () => import('@/pages/products/ProductList.vue') },
        { path: 'products/:id', component: () => import('@/pages/products/ProductDetail.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/users' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path !== '/login' && !auth.token) {
    return '/login'
  }
})

export default router
```

- [ ] **Step 3: 创建页面占位文件（防止路由报错）**

新建 `src/pages/users/UserList.vue`：
```vue
<template><div>用户列表（待实现）</div></template>
```

新建 `src/pages/users/UserDetail.vue`：
```vue
<template><div>用户详情（待实现）</div></template>
```

新建 `src/pages/products/ProductList.vue`：
```vue
<template><div>商品列表（待实现）</div></template>
```

新建 `src/pages/products/ProductDetail.vue`：
```vue
<template><div>商品详情（待实现）</div></template>
```

- [ ] **Step 4: 验证布局正常渲染**

先用 `create-admin` 命令创建管理员账号（后端需已启动）：
```bash
cd mall-server && ./mall-server-bin admin create-admin -c configs/config.yaml -u admin -p admin123
```

访问 http://localhost:5174/login，登录后应看到左侧导航栏和顶部 Header，点击导航可切换路由。

- [ ] **Step 5: Commit**

```bash
git add mall-admin-web/src/layouts/AdminLayout.vue \
        mall-admin-web/src/router/index.ts \
        mall-admin-web/src/pages/
git commit -m "feat: add AdminLayout with sidebar nav and complete router config"
```

---

### Task 10: 前端 — 用户管理页面

**Files:**
- Modify: `mall-admin-web/src/pages/users/UserList.vue`
- Modify: `mall-admin-web/src/pages/users/UserDetail.vue`

- [ ] **Step 1: 实现 `UserList.vue`**

```vue
<template>
  <el-card>
    <!-- 搜索栏 -->
    <el-form inline class="search-form">
      <el-form-item>
        <el-input v-model="params.keyword" placeholder="搜索用户名/昵称" clearable style="width:200px" />
      </el-form-item>
      <el-form-item>
        <el-select v-model="bannedFilter" placeholder="全部状态" clearable style="width:120px">
          <el-option label="正常" :value="false" />
          <el-option label="已封禁" :value="true" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 表格 -->
    <el-table :data="users" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="user_name" label="用户名" />
      <el-table-column prop="nick_name" label="昵称" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.is_banned ? 'danger' : 'success'">
            {{ row.is_banned ? '已封禁' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="product_count" label="商品数" width="80" />
      <el-table-column prop="created_at" label="注册时间" width="170" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="router.push(`/users/${row.id}`)">详情</el-button>
          <el-button size="small" :type="row.is_banned ? 'success' : 'danger'"
            @click="handleToggleBan(row)">
            {{ row.is_banned ? '解封' : '封禁' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      class="pagination"
      v-model:current-page="params.page"
      v-model:page-size="params.page_size"
      :total="total"
      :page-sizes="[10, 20, 50]"
      layout="total, sizes, prev, pager, next"
      @change="loadUsers"
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, banUser, unbanUser, type AdminUser } from '@/api/users'

const router = useRouter()
const loading = ref(false)
const users = ref<AdminUser[]>([])
const total = ref(0)
const bannedFilter = ref<boolean | undefined>(undefined)

const params = reactive({ keyword: '', page: 1, page_size: 10 })

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUserList({
      ...params,
      is_banned: bannedFilter.value,
    })
    if (res.code === 0) {
      users.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  params.page = 1
  loadUsers()
}

function handleReset() {
  params.keyword = ''
  bannedFilter.value = undefined
  params.page = 1
  loadUsers()
}

async function handleToggleBan(row: AdminUser) {
  const action = row.is_banned ? '解封' : '封禁'
  await ElMessageBox.confirm(`确认${action}用户 "${row.user_name}"？`, '提示', { type: 'warning' })
  try {
    const res = row.is_banned ? await unbanUser(row.id) : await banUser(row.id)
    if (res.code === 0) {
      ElMessage.success(`${action}成功`)
      loadUsers()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; display: flex; }
</style>
```

- [ ] **Step 2: 实现 `UserDetail.vue`**

```vue
<template>
  <el-card v-loading="loading">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <el-button @click="router.back()">← 返回列表</el-button>
        <el-button
          v-if="user"
          :type="user.is_banned ? 'success' : 'danger'"
          @click="handleToggleBan"
        >
          {{ user?.is_banned ? '解封用户' : '封禁用户' }}
        </el-button>
      </div>
    </template>

    <el-descriptions v-if="user" :column="2" border>
      <el-descriptions-item label="ID">{{ user.id }}</el-descriptions-item>
      <el-descriptions-item label="用户名">{{ user.user_name }}</el-descriptions-item>
      <el-descriptions-item label="昵称">{{ user.nick_name || '—' }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ user.phone || '—' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="user.is_banned ? 'danger' : 'success'">
          {{ user.is_banned ? '已封禁' : '正常' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="发布商品数">{{ user.product_count }}</el-descriptions-item>
      <el-descriptions-item label="注册时间">{{ user.created_at }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserDetail, banUser, unbanUser, type AdminUser } from '@/api/users'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const user = ref<AdminUser | null>(null)

async function loadDetail() {
  loading.value = true
  try {
    const res = await getUserDetail(Number(route.params.id))
    if (res.code === 0) user.value = res.data
  } finally {
    loading.value = false
  }
}

async function handleToggleBan() {
  if (!user.value) return
  const action = user.value.is_banned ? '解封' : '封禁'
  await ElMessageBox.confirm(`确认${action}用户 "${user.value.user_name}"？`, '提示', { type: 'warning' })
  try {
    const res = user.value.is_banned
      ? await unbanUser(user.value.id)
      : await banUser(user.value.id)
    if (res.code === 0) {
      ElMessage.success(`${action}成功`)
      loadDetail()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadDetail)
</script>
```

- [ ] **Step 3: 手动验证**

启动后端（`./mall-server-bin web -c configs/config.yaml`）和前端（`npm run dev`），访问 http://localhost:5174/users：
- 表格显示用户列表，按用户名搜索正常
- 点击「封禁」弹出确认框，操作后状态变为「已封禁」
- 点击「详情」跳转详情页，点击返回能回到列表

- [ ] **Step 4: Commit**

```bash
git add mall-admin-web/src/pages/users/
git commit -m "feat: implement user list and user detail pages with ban/unban"
```

---

### Task 11: 前端 — 商品管理页面

**Files:**
- Modify: `mall-admin-web/src/pages/products/ProductList.vue`
- Modify: `mall-admin-web/src/pages/products/ProductDetail.vue`

- [ ] **Step 1: 实现 `ProductList.vue`**

```vue
<template>
  <el-card>
    <!-- 搜索栏 -->
    <el-form inline class="search-form">
      <el-form-item>
        <el-input v-model="params.keyword" placeholder="搜索商品标题" clearable style="width:200px" />
      </el-form-item>
      <el-form-item>
        <el-select v-model="params.category" placeholder="全部分类" clearable style="width:120px">
          <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-select v-model="params.status" placeholder="全部状态" clearable style="width:120px">
          <el-option label="在售" :value="0" />
          <el-option label="已售出" :value="1" />
          <el-option label="已下架" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 表格 -->
    <el-table :data="products" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" show-overflow-tooltip />
      <el-table-column prop="price" label="价格" width="90">
        <template #default="{ row }">¥{{ row.price }}</template>
      </el-table-column>
      <el-table-column prop="category" label="分类" width="100" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 0 ? 'success' : row.status === 1 ? 'warning' : 'info'">
            {{ row.status === 0 ? '在售' : row.status === 1 ? '已售出' : '已下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="location" label="地区" width="120" />
      <el-table-column prop="seller" label="发布者" width="100" />
      <el-table-column prop="create_time" label="发布时间" width="170" />
      <el-table-column label="操作" width="140">
        <template #default="{ row }">
          <el-button size="small" @click="router.push(`/products/${row.id}`)">详情</el-button>
          <el-button size="small" type="danger" :disabled="row.status === 2"
            @click="handleDelist(row)">下架</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      class="pagination"
      v-model:current-page="params.page"
      v-model:page-size="params.page_size"
      :total="total"
      :page-sizes="[10, 20, 50]"
      layout="total, sizes, prev, pager, next"
      @change="loadProducts"
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductList, delistProduct, type AdminProduct } from '@/api/products'

const router = useRouter()
const loading = ref(false)
const products = ref<AdminProduct[]>([])
const total = ref(0)
const categories = ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他']

const params = reactive<{
  keyword: string; category: string | undefined; status: number | undefined; page: number; page_size: number
}>({ keyword: '', category: undefined, status: undefined, page: 1, page_size: 10 })

async function loadProducts() {
  loading.value = true
  try {
    const res = await getProductList(params)
    if (res.code === 0) {
      products.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() { params.page = 1; loadProducts() }
function handleReset() {
  params.keyword = ''; params.category = undefined; params.status = undefined; params.page = 1
  loadProducts()
}

async function handleDelist(row: AdminProduct) {
  await ElMessageBox.confirm(`确认强制下架商品 "${row.title}"？`, '提示', { type: 'warning' })
  try {
    const res = await delistProduct(row.id)
    if (res.code === 0) {
      ElMessage.success('已下架')
      loadProducts()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadProducts)
</script>

<style scoped>
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; display: flex; }
</style>
```

- [ ] **Step 2: 实现 `ProductDetail.vue`**

```vue
<template>
  <el-card v-loading="loading">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <el-button @click="router.back()">← 返回列表</el-button>
        <el-button
          v-if="product && product.status !== 2"
          type="danger"
          @click="handleDelist"
        >强制下架</el-button>
      </div>
    </template>

    <el-descriptions v-if="product" :column="2" border>
      <el-descriptions-item label="ID">{{ product.id }}</el-descriptions-item>
      <el-descriptions-item label="标题">{{ product.title }}</el-descriptions-item>
      <el-descriptions-item label="价格">¥{{ product.price }}</el-descriptions-item>
      <el-descriptions-item label="分类">{{ product.category || '—' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="product.status === 0 ? 'success' : product.status === 1 ? 'warning' : 'info'">
          {{ product.status === 0 ? '在售' : product.status === 1 ? '已售出' : '已下架' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="地区">{{ product.location }}</el-descriptions-item>
      <el-descriptions-item label="发布者">{{ product.seller }}</el-descriptions-item>
      <el-descriptions-item label="发布时间">{{ product.create_time }}</el-descriptions-item>
      <el-descriptions-item label="图片" :span="2">
        <div style="display:flex;gap:8px;flex-wrap:wrap">
          <el-image
            v-for="(img, i) in (product.images || '').split(',').filter(Boolean)"
            :key="i"
            :src="img"
            style="width:100px;height:100px;object-fit:cover;border-radius:4px"
            fit="cover"
          />
        </div>
      </el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductDetail, delistProduct, type AdminProduct } from '@/api/products'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const product = ref<AdminProduct | null>(null)

async function loadDetail() {
  loading.value = true
  try {
    const res = await getProductDetail(Number(route.params.id))
    if (res.code === 0) product.value = res.data
  } finally {
    loading.value = false
  }
}

async function handleDelist() {
  if (!product.value) return
  await ElMessageBox.confirm(`确认强制下架 "${product.value.title}"？`, '提示', { type: 'warning' })
  try {
    const res = await delistProduct(product.value.id)
    if (res.code === 0) {
      ElMessage.success('已下架')
      loadDetail()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadDetail)
</script>
```

- [ ] **Step 3: 手动验证**

访问 http://localhost:5174/products：
- 表格显示商品列表，按标题/分类/状态搜索正常
- 点击「下架」弹出确认框，已下架商品按钮变灰不可点
- 点击「详情」可看到商品图片和完整信息
- 下架按钮在详情页也生效

- [ ] **Step 4: 最终全量验证**

```bash
cd mall-server && go test ./...
```

Expected: 所有测试 PASS

验证完整流程：
1. 后端启动：`./mall-server-bin web -c configs/config.yaml`
2. 创建管理员：`./mall-server-bin admin create-admin -c configs/config.yaml -u admin -p admin123`
3. 前端启动：`cd mall-admin-web && npm run dev`
4. 访问 http://localhost:5174，登录，验证用户管理和商品管理全流程

- [ ] **Step 5: Commit**

```bash
git add mall-admin-web/src/pages/products/
git commit -m "feat: implement product list and product detail pages with force delist"
```
