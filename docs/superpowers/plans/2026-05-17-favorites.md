# 我的收藏功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为二手小程序新增收藏功能：用户在商品详情页收藏商品，在「我的收藏」列表页查看和取消收藏，已售出/下架商品自动从收藏列表消失。

**Architecture:** 新增 `user_favorite` 关联表存储收藏关系；后端提供 toggle/list 两个新接口，详情接口通过软鉴权中间件附加 `is_favorited` 字段；前端新增 myFavorite 页面，改造 detail 页心形按钮接入真实接口。

**Tech Stack:** Go 1.24 / Gin / GORM / SQLite（glebarez/sqlite）/ TypeScript / WeChat Mini Program

---

## 文件变更总览

### 后端（mall-server/）
| 文件 | 类型 | 职责 |
|------|------|------|
| `internal/app/dao/favorite.entity.go` | 新增 | UserFavorite GORM 实体 |
| `internal/app/dao/favorite.repo.go` | 新增 | ToggleFavorite / IsFavorited / GetFavoriteList |
| `internal/app/dao/favorite.repo_test.go` | 新增 | DAO 层单元测试 |
| `internal/app/dao/product.repo.go` | 修改 | ProductDetail 新增 IsFavorited 字段 |
| `internal/app/service/types.go` | 修改 | 新增 FavoriteToggleRequest / FavoriteListRequest |
| `internal/app/service/favorite.go` | 新增 | ToggleFavoriteHandler / GetFavoriteListHandler |
| `internal/app/service/product.go` | 修改 | GetProductDetail 读取 is_favorited |
| `internal/app/router/auth.go` | 修改 | 新增 OptionalAuthMiddleware |
| `internal/app/router/router.go` | 修改 | 注册 favorite 路由，detail 改用 OptionalAuth |
| `internal/app/models/init.go` | 修改 | AutoMigrate 新增 UserFavorite |

### 前端（mall-mini/miniprogram/）
| 文件 | 类型 | 职责 |
|------|------|------|
| `app.json` | 修改 | 注册 myFavorite 页面 |
| `pages/myFavorite/myFavorite.ts` | 新增 | 收藏列表页逻辑 |
| `pages/myFavorite/myFavorite.wxml` | 新增 | 收藏列表页模板 |
| `pages/myFavorite/myFavorite.wxss` | 新增 | 收藏列表页样式 |
| `pages/myFavorite/myFavorite.json` | 新增 | 页面配置 |
| `pages/detail/detail.ts` | 修改 | 接入 is_favorited + 真实 toggle 接口 |

> `pages/my/my.ts` 已有正确的跳转路径，无需修改。

---

## Task 1：UserFavorite 实体 + DB 迁移

**Files:**
- Create: `mall-server/internal/app/dao/favorite.entity.go`
- Modify: `mall-server/internal/app/models/init.go`

- [ ] **Step 1: 创建实体文件**

```go
// mall-server/internal/app/dao/favorite.entity.go
package dao

import "time"

type UserFavorite struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_product"`
	ProductID uint      `gorm:"not null;uniqueIndex:idx_user_product"`
	CreatedAt time.Time
}

func (UserFavorite) TableName() string {
	return "user_favorite"
}
```

- [ ] **Step 2: 在 init.go 里加入 AutoMigrate**

在 `mall-server/internal/app/models/init.go` 的 `NewDB()` 函数末尾（现有两个 AutoMigrate 之后）添加：

```go
// 自动迁移收藏表
if err := con.AutoMigrate(&dao.UserFavorite{}); err != nil {
    panic(fmt.Sprintf("db auto migrate error: %v", err))
}
```

- [ ] **Step 3: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无报错输出。

- [ ] **Step 4: Commit**

```bash
git add mall-server/internal/app/dao/favorite.entity.go \
        mall-server/internal/app/models/init.go
git commit -m "feat: add UserFavorite entity and auto-migrate"
```

---

## Task 2：Favorite DAO 函数 + 测试

**Files:**
- Create: `mall-server/internal/app/dao/favorite.repo.go`
- Create: `mall-server/internal/app/dao/favorite.repo_test.go`

- [ ] **Step 1: 写失败测试**

```go
// mall-server/internal/app/dao/favorite.repo_test.go
package dao

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupFavoriteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&UserFavorite{}, &Product{}, &SysUser{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestToggleFavorite_AddThenRemove(t *testing.T) {
	db := setupFavoriteTestDB(t)

	// 第一次：添加收藏
	isFav, err := ToggleFavorite(db, 1, 10)
	assert.NoError(t, err)
	assert.True(t, isFav)

	// 第二次：取消收藏
	isFav, err = ToggleFavorite(db, 1, 10)
	assert.NoError(t, err)
	assert.False(t, isFav)
}

func TestIsFavorited(t *testing.T) {
	db := setupFavoriteTestDB(t)

	isFav, err := IsFavorited(db, 1, 10)
	assert.NoError(t, err)
	assert.False(t, isFav)

	_, _ = ToggleFavorite(db, 1, 10)

	isFav, err = IsFavorited(db, 1, 10)
	assert.NoError(t, err)
	assert.True(t, isFav)
}

func TestGetFavoriteList_OnlyInSale(t *testing.T) {
	db := setupFavoriteTestDB(t)

	// 插入一个在售商品（status=0）和一个已售商品（status=1）
	p1 := Product{Title: "在售商品", Price: 100, Status: 0, UserId: 99}
	p2 := Product{Title: "已售商品", Price: 200, Status: 1, UserId: 99}
	db.Create(&p1)
	db.Create(&p2)

	// 都收藏
	db.Create(&UserFavorite{UserID: 1, ProductID: p1.ID})
	db.Create(&UserFavorite{UserID: 1, ProductID: p2.ID})

	results, total, err := GetFavoriteList(db, 1, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "在售商品", results[0].Title)
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd mall-server && go test ./internal/app/dao/... -run "TestToggleFavorite|TestIsFavorited|TestGetFavoriteList" -v
```

Expected: FAIL — `ToggleFavorite undefined`（函数未定义）。

- [ ] **Step 3: 实现 DAO 函数**

```go
// mall-server/internal/app/dao/favorite.repo.go
package dao

import (
	"errors"

	"gorm.io/gorm"
)

// ToggleFavorite 切换收藏状态。返回切换后是否已收藏。
func ToggleFavorite(db *gorm.DB, userID, productID uint) (bool, error) {
	var fav UserFavorite
	err := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&fav).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.Create(&UserFavorite{UserID: userID, ProductID: productID}).Error; createErr != nil {
			return false, createErr
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if delErr := db.Delete(&fav).Error; delErr != nil {
		return false, delErr
	}
	return false, nil
}

// IsFavorited 查询当前用户是否收藏了某商品。
func IsFavorited(db *gorm.DB, userID, productID uint) (bool, error) {
	var count int64
	err := db.Model(&UserFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// GetFavoriteList 获取用户收藏的在售商品列表（分页），按收藏时间倒序。
func GetFavoriteList(db *gorm.DB, userID uint, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&UserFavorite{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("JOIN product ON product.id = user_favorite.product_id AND product.status = 0").
		Joins("LEFT JOIN sys_user ON sys_user.id = product.user_id").
		Where("user_favorite.user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Order("user_favorite.created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd mall-server && go test ./internal/app/dao/... -run "TestToggleFavorite|TestIsFavorited|TestGetFavoriteList" -v
```

Expected: 全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add mall-server/internal/app/dao/favorite.repo.go \
        mall-server/internal/app/dao/favorite.repo_test.go
git commit -m "feat: add favorite DAO functions with tests"
```

---

## Task 3：OptionalAuthMiddleware

**Files:**
- Modify: `mall-server/internal/app/router/auth.go`

- [ ] **Step 1: 在 auth.go 末尾追加 OptionalAuthMiddleware**

在 `mall-server/internal/app/router/auth.go` 文件末尾（`AuthMiddleware` 函数之后）添加：

```go
// OptionalAuthMiddleware 软鉴权中间件。
// 有有效 token 时解析并注入 user_id / user_name；无 token 或 token 无效时静默跳过，不返回 401。
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := jwtx.ParseToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_name", claims.UserName)
		c.Next()
	}
}
```

- [ ] **Step 2: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无报错。

- [ ] **Step 3: Commit**

```bash
git add mall-server/internal/app/router/auth.go
git commit -m "feat: add OptionalAuthMiddleware for soft JWT auth"
```

---

## Task 4：类型定义 + 详情接口改造

**Files:**
- Modify: `mall-server/internal/app/service/types.go`
- Modify: `mall-server/internal/app/dao/product.repo.go`
- Modify: `mall-server/internal/app/service/product.go`

- [ ] **Step 1: 在 types.go 末尾添加 Favorite 请求类型**

在 `mall-server/internal/app/service/types.go` 末尾追加：

```go
// FavoriteToggleRequest 收藏/取消收藏请求
type FavoriteToggleRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// FavoriteListRequest 我的收藏列表请求
type FavoriteListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
```

- [ ] **Step 2: ProductDetail 新增 IsFavorited 字段**

在 `mall-server/internal/app/dao/product.repo.go` 的 `ProductDetail` 结构体中，在 `CreateTime` 字段之后追加：

```go
IsFavorited bool `json:"is_favorited"`
```

完整结构体改为：

```go
type ProductDetail struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Images      string  `json:"images"`
	Location    string  `json:"location"`
	Status      int     `json:"status"`
	BuyUid      uint    `json:"buy_uid"`
	Seller      string  `json:"seller"`
	Avatar      string  `json:"avatar"`
	CreateTime  string  `json:"create_time"`
	IsFavorited bool    `json:"is_favorited"`
}
```

- [ ] **Step 3: 改造 GetProductDetail handler 读取 is_favorited**

在 `mall-server/internal/app/service/product.go` 中，找到 `GetProductDetail` handler，在返回成功响应之前（`c.JSON(http.StatusOK, gin.H{...})` 之前）插入：

```go
// 读取当前用户的收藏状态（OptionalAuth 注入，未登录时跳过）
if userIDVal, exists := c.Get("user_id"); exists {
    isFav, _ := dao.IsFavorited(svc.DB, userIDVal.(uint), uint(id))
    detail.IsFavorited = isFav
}
```

改造后的 `GetProductDetail` 完整函数：

```go
func GetProductDetail(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		detail, err := dao.GetProductByID(svc.DB, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "商品不存在"})
			return
		}

		if userIDVal, exists := c.Get("user_id"); exists {
			isFav, _ := dao.IsFavorited(svc.DB, userIDVal.(uint), uint(id))
			detail.IsFavorited = isFav
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": detail,
		})
	}
}
```

- [ ] **Step 4: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无报错。

- [ ] **Step 5: Commit**

```bash
git add mall-server/internal/app/service/types.go \
        mall-server/internal/app/dao/product.repo.go \
        mall-server/internal/app/service/product.go
git commit -m "feat: add favorite request types and is_favorited to product detail"
```

---

## Task 5：Favorite Service Handlers + 路由注册

**Files:**
- Create: `mall-server/internal/app/service/favorite.go`
- Modify: `mall-server/internal/app/router/router.go`

- [ ] **Step 1: 创建 favorite.go**

```go
// mall-server/internal/app/service/favorite.go
package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
)

// ToggleFavoriteHandler POST /api/favorite/toggle
func ToggleFavoriteHandler(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req FavoriteToggleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
			return
		}

		isFav, err := dao.ToggleFavorite(svc.DB, userID.(uint), req.ProductID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{"is_favorited": isFav},
		})
	}
}

// GetFavoriteListHandler GET /api/favorite/list
func GetFavoriteListHandler(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req FavoriteListRequest
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

		results, total, err := dao.GetFavoriteList(svc.DB, userID.(uint), req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{
				"list":      results,
				"total":     total,
				"page":      req.Page,
				"page_size": req.PageSize,
			},
		})
	}
}
```

- [ ] **Step 2: 在 router.go 注册路由，detail 改用 OptionalAuth**

在 `mall-server/internal/app/router/router.go` 中：

1. 将原来的 `r.GET("/api/product/detail", service.GetProductDetail(svc))` 改为：
```go
r.GET("/api/product/detail", OptionalAuthMiddleware(), service.GetProductDetail(svc))
```

2. 在 `auth` 路由组（`auth.Use(AuthMiddleware())` 之后）新增两行：
```go
auth.POST("/api/favorite/toggle", service.ToggleFavoriteHandler(svc))
auth.GET("/api/favorite/list", service.GetFavoriteListHandler(svc))
```

- [ ] **Step 3: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无报错。

- [ ] **Step 4: 运行全部后端测试**

```bash
cd mall-server && go test ./...
```

Expected: PASS（包含 Task 2 的 DAO 测试）。

- [ ] **Step 5: Commit**

```bash
git add mall-server/internal/app/service/favorite.go \
        mall-server/internal/app/router/router.go
git commit -m "feat: add favorite service handlers and routes"
```

---

## Task 6：前端 - 注册页面 + myFavorite 四个文件

**Files:**
- Modify: `mall-mini/miniprogram/app.json`
- Create: `mall-mini/miniprogram/pages/myFavorite/myFavorite.json`
- Create: `mall-mini/miniprogram/pages/myFavorite/myFavorite.ts`
- Create: `mall-mini/miniprogram/pages/myFavorite/myFavorite.wxml`
- Create: `mall-mini/miniprogram/pages/myFavorite/myFavorite.wxss`

- [ ] **Step 1: 在 app.json 的 pages 数组末尾注册新页面**

在 `mall-mini/miniprogram/app.json` 的 `"pages"` 数组中，在 `"pages/productEdit/productEdit"` 之后追加：

```json
"pages/myFavorite/myFavorite"
```

完整 pages 数组：
```json
"pages": [
  "pages/home/home",
  "pages/detail/detail",
  "pages/search/search",
  "pages/publish/publish",
  "pages/my/my",
  "pages/myPublish/myPublish",
  "pages/productEdit/productEdit",
  "pages/myFavorite/myFavorite"
]
```

- [ ] **Step 2: 创建页面配置文件**

```json
// mall-mini/miniprogram/pages/myFavorite/myFavorite.json
{
  "navigationBarTitleText": "我的收藏",
  "enablePullDownRefresh": true,
  "backgroundColor": "#f7f8fa"
}
```

- [ ] **Step 3: 创建 TypeScript 逻辑文件**

```typescript
// mall-mini/miniprogram/pages/myFavorite/myFavorite.ts
import { get, post } from '../../utils/request'

interface FavoriteItem {
  id: number
  title: string
  price: number
  images: string
  location: string
  seller: string
  avatar: string
  firstImage: string
}

interface MyFavoriteData {
  items: FavoriteItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  hasMore: boolean
}

function processItems(list: any[]): FavoriteItem[] {
  return list.map(item => ({
    ...item,
    firstImage: item.images ? item.images.split(',')[0] : ''
  }))
}

Page<MyFavoriteData, WechatMiniprogram.IAnyObject>({
  data: {
    items: [],
    total: 0,
    page: 1,
    pageSize: 10,
    loading: false,
    hasMore: false
  },

  onLoad() {
    this.loadItems()
  },

  onShow() {
    this.setData({ items: [], page: 1, hasMore: false })
    this.loadItems()
  },

  onPullDownRefresh() {
    this.setData({ items: [], page: 1, hasMore: false })
    this.loadItems().finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadItems() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      const res = await get('/api/favorite/list', { page: 1, page_size: this.data.pageSize })
      const { list, total } = res.data
      const items = processItems(list || [])
      this.setData({ items, total, page: 1, hasMore: items.length < total })
    } catch (err) {
      console.error('加载收藏失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadMore() {
    if (this.data.loading) return
    const nextPage = this.data.page + 1
    this.setData({ loading: true })
    try {
      const res = await get('/api/favorite/list', { page: nextPage, page_size: this.data.pageSize })
      const { list } = res.data
      const more = processItems(list || [])
      const all = [...this.data.items, ...more]
      this.setData({ items: all, page: nextPage, hasMore: all.length < this.data.total })
    } catch (err) {
      console.error('加载更多失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/detail/detail?id=${id}` })
  },

  async onUnfavorite(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    try {
      await post('/api/favorite/toggle', { product_id: id })
      const items = this.data.items.filter(item => item.id !== id)
      this.setData({ items, total: this.data.total - 1 })
    } catch (err) {
      console.error('取消收藏失败', err)
    }
  }
})
```

- [ ] **Step 4: 创建 WXML 模板**

```xml
<!-- mall-mini/miniprogram/pages/myFavorite/myFavorite.wxml -->
<view class="container">
  <view wx:if="{{items.length === 0 && !loading}}" class="empty">
    <text class="empty-text">还没有收藏的商品</text>
  </view>

  <view
    wx:for="{{items}}"
    wx:key="id"
    class="product-card"
    bindtap="goToDetail"
    data-id="{{item.id}}"
  >
    <image
      wx:if="{{item.firstImage}}"
      class="product-thumb"
      src="{{item.firstImage}}"
      mode="aspectFill"
    />
    <view wx:else class="product-thumb placeholder" />

    <view class="product-meta">
      <text class="product-title">{{item.title}}</text>
      <text class="product-location">{{item.location}}</text>
      <view class="product-bottom">
        <text class="product-price">¥{{item.price}}</text>
        <text class="seller-name">{{item.seller}}</text>
      </view>
    </view>

    <view
      class="unfavorite-btn"
      catchtap="onUnfavorite"
      data-id="{{item.id}}"
    >❤️</view>
  </view>

  <view wx:if="{{loading}}" class="footer-tip">加载中...</view>
  <view wx:elif="{{!hasMore && items.length > 0}}" class="footer-tip">没有更多了</view>
</view>
```

- [ ] **Step 5: 创建 WXSS 样式**

```css
/* mall-mini/miniprogram/pages/myFavorite/myFavorite.wxss */
.container {
  min-height: 100vh;
  background: #f7f8fa;
  padding: 16rpx;
}

.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 200rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #9aa6b2;
}

.product-card {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 16rpx;
  padding: 20rpx;
  box-shadow: 0 8rpx 24rpx rgba(12, 18, 29, 0.06);
  display: flex;
  align-items: center;
}

.product-thumb {
  width: 140rpx;
  height: 140rpx;
  border-radius: 12rpx;
  flex-shrink: 0;
  background: #f0f4f7;
}

.product-thumb.placeholder {
  background: #eee;
}

.product-meta {
  flex: 1;
  margin-left: 20rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 140rpx;
  overflow: hidden;
}

.product-title {
  font-size: 28rpx;
  color: #111;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.product-location {
  font-size: 22rpx;
  color: #9aa6b2;
}

.product-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.product-price {
  font-size: 32rpx;
  font-weight: 700;
  color: #ff4d4f;
}

.seller-name {
  font-size: 22rpx;
  color: #9aa6b2;
}

.unfavorite-btn {
  font-size: 40rpx;
  padding: 10rpx 0 10rpx 20rpx;
  flex-shrink: 0;
}

.footer-tip {
  text-align: center;
  font-size: 24rpx;
  color: #9aa6b2;
  padding: 24rpx 0;
}
```

- [ ] **Step 6: Commit**

```bash
git add mall-mini/miniprogram/app.json \
        mall-mini/miniprogram/pages/myFavorite/
git commit -m "feat: add myFavorite page"
```

---

## Task 7：前端 - detail 页改造

**Files:**
- Modify: `mall-mini/miniprogram/pages/detail/detail.ts`

- [ ] **Step 1: 改造 detail.ts**

将 `mall-mini/miniprogram/pages/detail/detail.ts` 整体替换为以下内容（主要改动：`Product` 接口移除 `favorites`，`loadProductDetail` 从接口读取 `is_favorited`，`toggleFavorite` 接入真实接口）：

```typescript
// pages/detail/detail.ts
import { get, post } from '../../utils/request'

interface Seller {
  id: string
  name: string
  avatar: string
  rating: number
}

interface Product {
  id: number
  title: string
  description: string
  price: number
  originalPrice?: number
  images: string[]
  category: string
  condition: string
  location: string
  seller: Seller
  createdAt: string
  views: number
}

Page({
  data: {
    productId: '',
    product: {
      id: 0,
      title: '',
      description: '',
      price: 0,
      originalPrice: 0,
      images: [] as string[],
      category: '',
      condition: '九成新',
      location: '',
      seller: { id: '', name: '微信用户', avatar: '', rating: 4.5 },
      createdAt: '',
      views: 0
    },
    isFavorite: false,
    loading: true,
    error: null as string | null
  },

  onLoad(options: { id?: string }) {
    const { id } = options
    if (id) {
      this.setData({ productId: id })
    }
    this.loadProductDetail()
  },

  onPullDownRefresh() {
    this.loadProductDetail()
  },

  onShareAppMessage() {
    const { product } = this.data
    return {
      title: product.title,
      path: `/pages/detail/detail?id=${product.id}`,
      imageUrl: product.images[0] || ''
    }
  },

  async loadProductDetail() {
    this.setData({ loading: true, error: null })

    try {
      const response = await get<any>('/api/product/detail', { id: this.data.productId })

      if (response.code === 0 && response.data) {
        const data = response.data
        const images = data.images ? data.images.split(',').filter((img: string) => img) : []

        const product: Product = {
          id: data.id,
          title: data.title,
          description: data.description || '',
          price: data.price,
          images,
          category: '二手好物',
          condition: '九成新',
          location: data.location || '',
          seller: {
            id: String(data.id),
            name: data.seller || '微信用户',
            avatar: data.avatar || '',
            rating: 4.8
          },
          createdAt: data.create_time ? data.create_time.substring(0, 10) : '',
          views: Math.floor(Math.random() * 500) + 50
        }

        this.setData({
          product,
          isFavorite: !!data.is_favorited,
          loading: false
        })
      } else {
        this.setData({ error: response.msg || '加载失败', loading: false })
      }
    } catch (err) {
      console.error('加载商品详情失败:', err)
      this.setData({ error: '网络错误，请重试', loading: false })
    }

    wx.stopPullDownRefresh()
  },

  async toggleFavorite() {
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }

    try {
      const res = await post<{ is_favorited: boolean }>('/api/favorite/toggle', {
        product_id: this.data.product.id
      })
      if (res.code === 0 && res.data) {
        const isFavorite = res.data.is_favorited
        this.setData({ isFavorite })
        wx.showToast({ title: isFavorite ? '已收藏' : '已移除收藏', icon: 'success' })
      }
    } catch (err) {
      console.error('收藏操作失败:', err)
    }
  },

  contactSeller() {
    wx.showToast({ title: '功能开发中', icon: 'none' })
  },

  reportProduct() {
    wx.showActionSheet({
      itemList: ['虚假信息', '骚扰信息', '违法违规'],
      success: () => {
        wx.showToast({ title: '举报成功', icon: 'success' })
      }
    })
  }
})
```

- [ ] **Step 2: Commit**

```bash
git add mall-mini/miniprogram/pages/detail/detail.ts
git commit -m "feat: wire up real favorite toggle in detail page"
```

---

## 验收标准

手动测试以下场景：

1. **未登录用户**进入详情页 → 心形图标为空心灰 → 点击心形 → 弹出"请先登录"提示
2. **已登录用户**进入详情页 → 心形图标反映真实收藏状态
3. 点击心形 → 变为红色实心 → 再点击 → 变回空心灰（切换流畅）
4. 进入「我的」→「我的收藏」→ 看到已收藏商品列表
5. 在收藏列表点击心形 → 该商品立即从列表消失
6. 在收藏列表点击商品行 → 跳转到详情页
7. 将一个已收藏商品下架/标记售出 → 重新进入「我的收藏」→ 该商品不再出现
8. 发布商品后返回首页 → 新商品出现在列表（前一个 bug 的回归验证）
