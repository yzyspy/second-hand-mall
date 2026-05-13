# 搜索结果页面 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为二手商城增加搜索结果页面，支持关键词搜索、按发布时间排序、按商品状态筛选。

**Architecture:** 后端新增 Product 实体、Repository 搜索函数和 Service 处理器，使用 GORM LIKE 实现模糊搜索，支持分页和排序。前端新增 search 页面，包含搜索栏、排序筛选栏和商品列表，复用首页卡片样式。

**Tech Stack:** Go / Gin / GORM / SQLite (后端)，TypeScript / WeChat Mini Program (前端)

---

## File Structure

### 后端新增
- `mall-server/internal/app/dao/product.entity.go` — Product 实体定义
- `mall-server/internal/app/dao/product.repo.go` — Product 仓库函数 (SearchProducts)
- `mall-server/internal/app/service/product.go` — 商品搜索处理器 (SearchProducts)

### 后端修改
- `mall-server/internal/app/models/init.go:29` — AutoMigrate 新增 Product
- `mall-server/internal/app/router/router.go:68` — 注册搜索路由

### 前端新增
- `miniprogram/pages/search/search.ts` — 页面逻辑
- `miniprogram/pages/search/search.wxml` — 页面结构
- `miniprogram/pages/search/search.wxss` — 页面样式
- `miniprogram/pages/search/search.json` — 页面配置

### 前端修改
- `miniprogram/app.json:4` — 注册 search 页面路由

---

### Task 1: 创建 Product 实体

**Files:**
- Create: `mall-server/internal/app/dao/product.entity.go`

- [ ] **Step 1: 编写 Product 实体**

```go
package dao

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Title       string  `gorm:"column:title;type:varchar(200);not null;default:''" json:"title" comment:"商品标题"`
	Description string  `gorm:"column:description;type:text;not null;default:''" json:"description" comment:"商品描述"`
	Price       float64 `gorm:"column:price;type:decimal(10,2);not null;default:0" json:"price" comment:"价格"`
	Images      string  `gorm:"column:images;type:varchar(1000);not null;default:''" json:"images" comment:"图片URL列表,逗号分隔"`
	Location    string  `gorm:"column:location;type:varchar(100);not null;default:''" json:"location" comment:"交易地点"`
	Status      int     `gorm:"column:status;type:int;not null;default:0" json:"status" comment:"状态:0在售,1已售出,2已下架"`
	UserId      uint    `gorm:"column:user_id;type:int;not null;default:0" json:"user_id" comment:"发布者ID"`
}

func (Product) TableName() string {
	return "product"
}
```

- [ ] **Step 2: 提交**

```bash
cd mall-server && git add internal/app/dao/product.entity.go && git commit -m "feat: add Product entity definition"
```

---

### Task 2: 创建 Product Repository 搜索函数

**Files:**
- Create: `mall-server/internal/app/dao/product.repo.go`

- [ ] **Step 1: 编写 SearchProducts 函数**

```go
package dao

import (
	"gorm.io/gorm"
)

type ProductSearchResult struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Images     string  `json:"images"`
	Location   string  `json:"location"`
	Status     int     `json:"status"`
	Seller     string  `json:"seller"`
	CreateTime string  `json:"create_time"`
}

func SearchProducts(db *gorm.DB, keyword string, sort string, status *int, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).Select("product.id, product.title, product.price, product.images, product.location, product.status, sys_user.nick_name as seller, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id")

	if keyword != "" {
		query = query.Where("product.title LIKE ?", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("product.status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch sort {
	case "time_asc":
		query = query.Order("product.created_at ASC")
	default:
		query = query.Order("product.created_at DESC")
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
```

- [ ] **Step 2: 提交**

```bash
cd mall-server && git add internal/app/dao/product.repo.go && git commit -m "feat: add SearchProducts repository function"
```

---

### Task 3: 创建 Product Service 处理器

**Files:**
- Create: `mall-server/internal/app/service/product.go`
- Modify: `mall-server/internal/app/service/types.go`

- [ ] **Step 1: 在 types.go 中添加搜索请求结构体**

在 `mall-server/internal/app/service/types.go` 末尾添加：

```go
// SearchProductRequest 商品搜索请求
type SearchProductRequest struct {
	Keyword  string `form:"keyword"`
	Sort     string `form:"sort"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
```

- [ ] **Step 2: 编写 SearchProducts 处理器**

创建 `mall-server/internal/app/service/product.go`：

```go
package service

import (
	"github.com/gin-gonic/gin"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"net/http"
)

func SearchProducts(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchProductRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "参数错误",
			})
			return
		}

		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 50 {
			req.PageSize = 10
		}

		results, total, err := dao.SearchProducts(svc.DB, req.Keyword, req.Sort, req.Status, req.Page, req.PageSize)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "搜索失败",
			})
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

- [ ] **Step 3: 提交**

```bash
cd mall-server && git add internal/app/service/product.go internal/app/service/types.go && git commit -m "feat: add SearchProducts service handler"
```

---

### Task 4: 注册路由和自动迁移

**Files:**
- Modify: `mall-server/internal/app/router/router.go:68`
- Modify: `mall-server/internal/app/models/init.go:29`

- [ ] **Step 1: 在 init.go 中添加 Product 自动迁移**

在 `mall-server/internal/app/models/init.go` 的 `NewDB()` 函数中，在已有 `AutoMigrate(&dao.SysUser{})` 后添加：

```go
if err := con.AutoMigrate(&dao.Product{}); err != nil {
    panic(fmt.Sprintf("db auto migrate error: %v", err))
}
```

- [ ] **Step 2: 在 router.go 中注册搜索路由**

在 `mall-server/internal/app/router/router.go` 的 `App()` 函数中，在 `cos-signature-v2` 路由之后添加：

```go
// 商品搜索接口
r.GET("/api/product/search", service.SearchProducts(svc))
```

- [ ] **Step 3: 验证编译通过**

Run: `cd mall-server && go build -o mall-server`
Expected: 编译成功，无错误

- [ ] **Step 4: 提交**

```bash
cd mall-server && git add internal/app/models/init.go internal/app/router/router.go && git commit -m "feat: register search route and Product auto-migration"
```

---

### Task 5: 创建搜索页面配置和逻辑

**Files:**
- Create: `miniprogram/pages/search/search.json`
- Create: `miniprogram/pages/search/search.ts`

- [ ] **Step 1: 创建 search.json**

```json
{
  "navigationBarTitleText": "搜索商品",
  "enablePullDownRefresh": false
}
```

- [ ] **Step 2: 创建 search.ts**

```typescript
import { get } from '../../utils/request'

interface ProductItem {
  id: number
  title: string
  price: number
  images: string
  location: string
  status: number
  seller: string
  create_time: string
}

interface SearchData {
  keyword: string
  sort: string
  status: number | null
  products: ProductItem[]
  loading: boolean
  page: number
  hasMore: boolean
  sortOptions: string[]
  statusOptions: string[]
  sortIndex: number
  statusIndex: number
}

Page<SearchData, WechatMiniprogram.IAnyObject>({
  data: {
    keyword: '',
    sort: 'time_desc',
    status: null,
    products: [],
    loading: false,
    page: 1,
    hasMore: true,
    sortOptions: ['最新发布', '最早发布'],
    statusOptions: ['全部', '在售', '已售出'],
    sortIndex: 0,
    statusIndex: 0
  },

  onLoad(options: Record<string, string>) {
    if (options.keyword) {
      this.setData({ keyword: options.keyword })
      this.search()
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.search()
    }
  },

  onKeywordInput(e: WechatMiniprogram.Input) {
    this.setData({ keyword: e.detail.value })
  },

  onSearch() {
    this.setData({ page: 1, hasMore: true, products: [] })
    this.search()
  },

  onSortChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const sortMap = ['time_desc', 'time_asc']
    this.setData({ sortIndex: index, sort: sortMap[index], page: 1, hasMore: true, products: [] })
    this.search()
  },

  onStatusChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const statusMap: (number | null)[] = [null, 0, 1]
    this.setData({ statusIndex: index, status: statusMap[index], page: 1, hasMore: true, products: [] })
    this.search()
  },

  async search() {
    if (this.data.loading) return
    this.setData({ loading: true })

    try {
      const res = await get<{
        list: ProductItem[]
        total: number
        page: number
        page_size: number
      }>('/api/product/search', {
        keyword: this.data.keyword,
        sort: this.data.sort,
        status: this.data.status,
        page: this.data.page,
        page_size: 10
      })

      if (res.code === 0 && res.data) {
        const newList = res.data.list || []
        this.setData({
          products: this.data.page === 1 ? newList : [...this.data.products, ...newList],
          hasMore: this.data.products.length + newList.length < res.data.total,
          page: this.data.page + 1
        })
      }
    } catch {
      // 请求失败，request.ts 已处理 toast 提示
    } finally {
      this.setData({ loading: false })
    }
  },

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/detail/detail?id=${id}` })
  }
})
```

- [ ] **Step 3: 提交**

```bash
git add miniprogram/pages/search/search.json miniprogram/pages/search/search.ts && git commit -m "feat: add search page logic"
```

---

### Task 6: 创建搜索页面模板和样式

**Files:**
- Create: `miniprogram/pages/search/search.wxml`
- Create: `miniprogram/pages/search/search.wxss`
- Modify: `miniprogram/app.json:4`

- [ ] **Step 1: 注册 search 页面路由**

在 `miniprogram/app.json` 的 `pages` 数组中，在 `pages/detail/detail` 后添加：

```json
"pages/search/search"
```

- [ ] **Step 2: 创建 search.wxml**

```xml
<view class="search-container">
  <!-- 搜索栏 -->
  <view class="search-bar">
    <input
      class="search-input"
      placeholder="搜索商品"
      value="{{keyword}}"
      bindinput="onKeywordInput"
      bindconfirm="onSearch"
      confirm-type="search"
    />
    <view class="search-btn" bindtap="onSearch">搜索</view>
  </view>

  <!-- 排序筛选栏 -->
  <view class="filter-bar">
    <view class="filter-item">
      <text class="filter-label">排序：</text>
      <picker mode="selector" range="{{sortOptions}}" value="{{sortIndex}}" bindchange="onSortChange">
        <view class="filter-picker">
          {{sortOptions[sortIndex]}}
          <text class="picker-arrow">▼</text>
        </view>
      </picker>
    </view>
    <view class="filter-item">
      <text class="filter-label">状态：</text>
      <picker mode="selector" range="{{statusOptions}}" value="{{statusIndex}}" bindchange="onStatusChange">
        <view class="filter-picker">
          {{statusOptions[statusIndex]}}
          <text class="picker-arrow">▼</text>
        </view>
      </picker>
    </view>
  </view>

  <!-- 商品列表 -->
  <view class="product-list">
    <block wx:for="{{products}}" wx:key="id">
      <view class="product-card" bindtap="goToDetail" data-id="{{item.id}}">
        <view class="img-wrap">
          <image class="product-image" src="{{item.images.split(',')[0]}}" mode="aspectFill" />
          <view class="price-badge">¥{{item.price}}</view>
        </view>
        <view class="product-info">
          <view class="product-title">{{item.title}}</view>
          <view class="product-meta">
            <text class="product-location">📍 {{item.location}}</text>
            <text class="product-seller">卖家：{{item.seller}}</text>
          </view>
          <view class="product-status">
            <text class="status-tag status-{{item.status}}">{{item.status === 0 ? '在售' : item.status === 1 ? '已售出' : '已下架'}}</text>
          </view>
        </view>
      </view>
    </block>
  </view>

  <!-- 加载状态 -->
  <view class="loading-container" wx:if="{{loading}}">
    <text>加载中...</text>
  </view>

  <!-- 没有更多 -->
  <view class="no-more" wx:if="{{!hasMore && products.length > 0}}">
    <text>已经到底了~</text>
  </view>

  <!-- 空状态 -->
  <view class="empty-state" wx:if="{{!loading && products.length === 0 && keyword}}">
    <text class="empty-text">没有找到相关商品</text>
  </view>

  <!-- 未搜索状态 -->
  <view class="empty-state" wx:if="{{!loading && products.length === 0 && !keyword}}">
    <text class="empty-text">输入关键词搜索商品</text>
  </view>
</view>
```

- [ ] **Step 3: 创建 search.wxss**

```css
/* pages/search/search.wxss */

.search-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 20rpx;
}

/* 搜索栏 */
.search-bar {
  background: #fff;
  padding: 20rpx 30rpx;
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.search-input {
  flex: 1;
  background: #f5f5f5;
  border-radius: 36rpx;
  padding: 16rpx 28rpx;
  font-size: 28rpx;
}

.search-btn {
  background: #07C160;
  color: #fff;
  border-radius: 36rpx;
  padding: 16rpx 32rpx;
  font-size: 28rpx;
  font-weight: 600;
  white-space: nowrap;
}

.search-btn:active {
  background: #06ad56;
}

/* 排序筛选栏 */
.filter-bar {
  background: #fff;
  padding: 16rpx 30rpx;
  display: flex;
  align-items: center;
  gap: 30rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-item {
  display: flex;
  align-items: center;
}

.filter-label {
  font-size: 26rpx;
  color: #666;
}

.filter-picker {
  display: flex;
  align-items: center;
  font-size: 26rpx;
  color: #333;
  background: #f5f5f5;
  padding: 8rpx 16rpx;
  border-radius: 20rpx;
}

.picker-arrow {
  font-size: 20rpx;
  color: #999;
  margin-left: 8rpx;
}

/* 商品列表 */
.product-list {
  padding: 0 20rpx;
  margin-top: 20rpx;
}

.product-card {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
  display: flex;
  padding: 16rpx;
  box-shadow: 0 6rpx 20rpx rgba(12, 18, 29, 0.06);
  align-items: flex-start;
}

.img-wrap {
  position: relative;
  width: 220rpx;
  height: 220rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  overflow: hidden;
}

.product-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  background: linear-gradient(180deg, #fafafa, #f5f5f5);
}

.price-badge {
  position: absolute;
  left: 12rpx;
  bottom: 12rpx;
  background: rgba(7, 193, 96, 0.95);
  color: #fff;
  padding: 8rpx 14rpx;
  border-radius: 20rpx;
  font-size: 26rpx;
  font-weight: 600;
}

.product-info {
  flex: 1;
  margin-left: 18rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.product-title {
  font-size: 30rpx;
  color: #111;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  line-height: 1.3;
  margin-bottom: 10rpx;
}

.product-meta {
  display: flex;
  justify-content: flex-start;
  gap: 20rpx;
  font-size: 24rpx;
  color: #888;
  align-items: center;
  margin-bottom: 8rpx;
}

.product-location,
.product-seller {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 320rpx;
}

/* 状态标签 */
.product-status {
  margin-top: 4rpx;
}

.status-tag {
  font-size: 22rpx;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.status-0 {
  background: #e6f9ee;
  color: #07C160;
}

.status-1 {
  background: #fff3e0;
  color: #ff9800;
}

.status-2 {
  background: #f5f5f5;
  color: #999;
}

/* 加载状态 */
.loading-container {
  text-align: center;
  padding: 40rpx 0;
  color: #999;
}

.no-more {
  text-align: center;
  padding: 40rpx 0;
  color: #999;
  font-size: 24rpx;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 200rpx 0;
}

.empty-text {
  color: #999;
  font-size: 28rpx;
}
```

- [ ] **Step 4: 提交**

```bash
git add miniprogram/pages/search/search.wxml miniprogram/pages/search/search.wxss miniprogram/app.json && git commit -m "feat: add search page template, styles, and register route"
```

---

### Task 7: 验证整体功能

- [ ] **Step 1: 验证后端编译**

Run: `cd mall-server && go build -o mall-server`
Expected: 编译成功

- [ ] **Step 2: 启动后端服务并测试搜索 API**

Run: `cd mall-server && ./mall-server web -config configs/config.yaml`

在另一个终端测试：
```bash
# 测试无参数搜索（返回全部）
curl "http://localhost:8080/api/product/search"

# 测试关键词搜索
curl "http://localhost:8080/api/product/search?keyword=MacBook"

# 测试排序
curl "http://localhost:8080/api/product/search?sort=time_asc"

# 测试状态筛选
curl "http://localhost:8080/api/product/search?status=0"

# 测试组合查询
curl "http://localhost:8080/api/product/search?keyword=MacBook&sort=time_desc&status=0&page=1&page_size=10"
```

Expected: 返回 `{ "code": 0, "msg": "success", "data": { "list": [], "total": 0, "page": 1, "page_size": 10 } }`

- [ ] **Step 3: 在微信开发者工具中验证前端**

1. 打开 mall-mini 项目
2. 确认 app.json 中 search 页面已注册，编译无报错
3. 在首页点击搜索栏，应跳转到搜索页面
4. 输入关键词点击搜索，应显示加载状态
5. 切换排序和状态筛选，应重新搜索
6. 滚动到底部，应加载更多（如有多页数据）

- [ ] **Step 4: 最终提交**

如有任何修复，提交所有更改：

```bash
git add -A && git commit -m "feat: search results page with keyword, time sort, and status filter"
```
