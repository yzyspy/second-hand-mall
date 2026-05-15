# 重新编辑已发布商品 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在"我的"页面新增"我的发布"列表，用户可浏览自己的商品并执行编辑、标记已售、下架、删除操作。

**Architecture:** 后端新增 3 个受 JWT 保护的接口（mine / update / change-status），前端新增 2 个页面（myPublish 列表页 + productEdit 编辑页），编辑页独立于发布页，通过 `onLoad` 的 `options.id` 拉取原始数据回填。

**Tech Stack:** Go 1.21 + Gin + GORM + SQLite（后端）；微信小程序 TypeScript（前端）；Qiniu 图片存储。

---

## 文件结构

**新建：**
- `mall-server/internal/app/dao/product.repo.go` — 追加 3 个函数（GetMyProducts / UpdateProduct / ChangeProductStatus）
- `mall-mini/miniprogram/pages/myPublish/myPublish.{ts,wxml,wxss,json}`
- `mall-mini/miniprogram/pages/productEdit/productEdit.{ts,wxml,wxss,json}`

**修改：**
- `mall-server/internal/app/service/types.go` — 追加 3 个 Request struct + MyProductItem
- `mall-server/internal/app/service/product.go` — 追加 3 个 handler
- `mall-server/internal/app/router/router.go` — 注册 3 条路由
- `mall-mini/miniprogram/utils/request.ts` — 追加 `put` 工具函数
- `mall-mini/miniprogram/app.json` — 注册 2 个新页面路径

---

## Task 1: 后端 — 新增 Request/Response 类型

**Files:**
- Modify: `mall-server/internal/app/service/types.go`

- [ ] **Step 1: 在 types.go 末尾追加以下代码**

```go
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
	Status int  `json:"status"`
}
```

- [ ] **Step 2: 确认编译通过**

```bash
cd mall-server && go build ./...
```

期望：无错误输出。

- [ ] **Step 3: Commit**

```bash
git add mall-server/internal/app/service/types.go
git commit -m "feat: add request/response types for product management"
```

---

## Task 2: 后端 — dao 层新增 3 个函数

**Files:**
- Modify: `mall-server/internal/app/dao/product.repo.go`

- [ ] **Step 1: 在 product.repo.go 末尾追加以下代码**

```go
// GetMyProducts 获取指定用户发布的商品列表
func GetMyProducts(db *gorm.DB, userID uint, page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id").
		Where("product.user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []ProductSearchResult
	offset := (page - 1) * pageSize
	if err := query.Order("product.created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// UpdateProduct 更新商品内容，仅允许在售（status=0）且属于本人的商品
func UpdateProduct(db *gorm.DB, id uint, userID uint, updates map[string]interface{}) error {
	result := db.Model(&Product{}).
		Where("id = ? AND user_id = ? AND status = 0", id, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在或无权限编辑")
	}
	return nil
}

// ChangeProductStatus 变更商品状态，须本人校验
func ChangeProductStatus(db *gorm.DB, id uint, userID uint, status int) error {
	result := db.Model(&Product{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在或无权限操作")
	}
	return nil
}
```

- [ ] **Step 2: 确认编译通过**

```bash
cd mall-server && go build ./...
```

期望：无错误输出。

- [ ] **Step 3: Commit**

```bash
git add mall-server/internal/app/dao/product.repo.go
git commit -m "feat: add GetMyProducts, UpdateProduct, ChangeProductStatus to dao"
```

---

## Task 3: 后端 — service 层新增 3 个 handler

**Files:**
- Modify: `mall-server/internal/app/service/product.go`

- [ ] **Step 1: 在 product.go 顶部 import 块中确认已有 `"strings"` import（已存在），无需修改**

- [ ] **Step 2: 在 product.go 末尾追加以下 3 个 handler**

```go
// GetMyProducts 获取当前用户发布的商品列表
func GetMyProducts(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req GetMyProductsRequest
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

		results, total, err := dao.GetMyProducts(svc.DB, userID.(uint), req.Page, req.PageSize)
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

// UpdateProduct 更新商品内容
func UpdateProduct(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误: " + err.Error()})
			return
		}

		// title 取 description 前 50 字，与 publish 保持一致
		title := req.Description
		runes := []rune(title)
		if len(runes) > 50 {
			title = string(runes[:50])
		}

		updates := map[string]interface{}{
			"title":       title,
			"description": req.Description,
			"price":       req.Price,
			"location":    req.Location,
			"images":      strings.Join(req.Images, ","),
		}

		if err := dao.UpdateProduct(svc.DB, req.ID, userID.(uint), updates); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
	}
}

// ChangeProductStatus 变更商品状态
func ChangeProductStatus(svc *models.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未登录"})
			return
		}

		var req ChangeProductStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误: " + err.Error()})
			return
		}
		if req.Status != 1 && req.Status != 2 {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "状态值无效，只允许 1(已售) 或 2(下架)"})
			return
		}

		if err := dao.ChangeProductStatus(svc.DB, req.ID, userID.(uint), req.Status); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}
```

- [ ] **Step 3: 确认编译通过**

```bash
cd mall-server && go build ./...
```

期望：无错误输出。

- [ ] **Step 4: Commit**

```bash
git add mall-server/internal/app/service/product.go
git commit -m "feat: add GetMyProducts, UpdateProduct, ChangeProductStatus handlers"
```

---

## Task 4: 后端 — 注册路由并验证接口

**Files:**
- Modify: `mall-server/internal/app/router/router.go`

- [ ] **Step 1: 在 router.go 的认证路由组（`auth := r.Group("/")`）末尾追加 3 条路由**

在 `auth.POST("/api/product/publish", ...)` 之后添加：

```go
// 获取我发布的商品列表
auth.GET("/api/product/mine", service.GetMyProducts(svc))
// 更新商品内容
auth.PUT("/api/product/update", service.UpdateProduct(svc))
// 变更商品状态（下架/标记已售）
auth.POST("/api/product/change-status", service.ChangeProductStatus(svc))
```

- [ ] **Step 2: 编译并启动服务**

```bash
cd mall-server && go build -o mall-server && ./mall-server web -config configs/config.yaml
```

期望：服务正常启动，监听 8080 端口。

- [ ] **Step 3: 获取 token（替换 `<YOUR_TOKEN>` 为实际登录返回的 token）**

```bash
curl -s -X POST http://localhost:8080/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"kane","password":"111"}' | python3 -m json.tool
```

期望：返回 `{"code":0, "data":{"token":"eyJ..."}}`

- [ ] **Step 4: 验证 GET /api/product/mine**

```bash
curl -s "http://localhost:8080/api/product/mine?page=1&page_size=5" \
  -H "Authorization: Bearer <YOUR_TOKEN>" | python3 -m json.tool
```

期望：`{"code":0,"data":{"list":[...],"total":N,"page":1,"page_size":5}}`

- [ ] **Step 5: 验证 POST /api/product/change-status（下架一个自己的商品，替换 `<ID>`）**

```bash
curl -s -X POST http://localhost:8080/api/product/change-status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{"id":<ID>,"status":2}' | python3 -m json.tool
```

期望：`{"code":0,"msg":"操作成功"}`

- [ ] **Step 6: Commit**

```bash
git add mall-server/internal/app/router/router.go
git commit -m "feat: register mine/update/change-status routes"
```

---

## Task 5: 前端 — request.ts 追加 put 工具函数

**Files:**
- Modify: `mall-mini/miniprogram/utils/request.ts`

- [ ] **Step 1: 在 `post` 函数之后追加 `put` 函数**

```typescript
/**
 * PUT请求
 */
export function put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'PUT', data })
}
```

- [ ] **Step 2: Commit**

```bash
git add mall-mini/miniprogram/utils/request.ts
git commit -m "feat: add put helper to request util"
```

---

## Task 6: 前端 — app.json 注册新页面

**Files:**
- Modify: `mall-mini/miniprogram/app.json`

- [ ] **Step 1: 在 `pages` 数组末尾追加两个路径**

将 `app.json` 的 `pages` 数组改为：

```json
"pages": [
  "pages/home/home",
  "pages/detail/detail",
  "pages/search/search",
  "pages/publish/publish",
  "pages/my/my",
  "pages/myPublish/myPublish",
  "pages/productEdit/productEdit"
]
```

- [ ] **Step 2: Commit**

```bash
git add mall-mini/miniprogram/app.json
git commit -m "feat: register myPublish and productEdit pages"
```

---

## Task 7: 前端 — myPublish 列表页

**Files:**
- Create: `mall-mini/miniprogram/pages/myPublish/myPublish.json`
- Create: `mall-mini/miniprogram/pages/myPublish/myPublish.wxml`
- Create: `mall-mini/miniprogram/pages/myPublish/myPublish.wxss`
- Create: `mall-mini/miniprogram/pages/myPublish/myPublish.ts`

- [ ] **Step 1: 创建 myPublish.json**

```json
{
  "navigationBarTitleText": "我发布的",
  "enablePullDownRefresh": true,
  "backgroundColor": "#f7f8fa"
}
```

- [ ] **Step 2: 创建 myPublish.wxml**

```xml
<!--pages/myPublish/myPublish.wxml-->
<view class="container">
  <view wx:if="{{products.length === 0 && !loading}}" class="empty">
    <text class="empty-text">还没有发布的商品</text>
  </view>

  <view wx:for="{{products}}" wx:key="id" class="product-card">
    <view class="product-info">
      <image
        wx:if="{{item.firstImage}}"
        class="product-thumb"
        src="{{item.firstImage}}"
        mode="aspectFill"
      />
      <view wx:else class="product-thumb placeholder" />
      <view class="product-meta">
        <text class="product-title">{{item.title}}</text>
        <view class="product-bottom">
          <text class="product-price">¥{{item.price}}</text>
          <text class="status-tag {{item.statusClass}}">{{item.statusText}}</text>
        </view>
      </view>
    </view>
    <view class="product-actions">
      <block wx:if="{{item.status === 0}}">
        <view class="action-btn btn-edit" bindtap="onEditTap" data-id="{{item.id}}">编辑</view>
        <view class="action-btn btn-sold" bindtap="onMarkSold" data-id="{{item.id}}">标记已售</view>
        <view class="action-btn btn-delist" bindtap="onDelist" data-id="{{item.id}}">下架</view>
      </block>
      <block wx:else>
        <view class="action-btn btn-delete" bindtap="onDelete" data-id="{{item.id}}">删除</view>
      </block>
    </view>
  </view>

  <view wx:if="{{loading}}" class="footer-tip">加载中...</view>
  <view wx:elif="{{!hasMore && products.length > 0}}" class="footer-tip">没有更多了</view>
</view>
```

- [ ] **Step 3: 创建 myPublish.wxss**

```css
/* pages/myPublish/myPublish.wxss */
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
}

.product-info {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
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
}

.product-title {
  font-size: 28rpx;
  color: #111;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
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

.status-tag {
  font-size: 22rpx;
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
}

.status-on-sale {
  background: #e6f7ee;
  color: #07C160;
}

.status-sold {
  background: #fff0f0;
  color: #ff4d4f;
}

.status-off {
  background: #f0f0f0;
  color: #999;
}

.product-actions {
  display: flex;
  gap: 16rpx;
  border-top: 1rpx solid #f0f0f0;
  padding-top: 16rpx;
}

.action-btn {
  flex: 1;
  height: 60rpx;
  line-height: 60rpx;
  text-align: center;
  border-radius: 30rpx;
  font-size: 24rpx;
}

.btn-edit {
  background: #07C160;
  color: #fff;
}

.btn-sold {
  background: #fff0e6;
  color: #fa8c16;
  border: 1rpx solid #ffd591;
}

.btn-delist {
  background: #f0f0f0;
  color: #666;
}

.btn-delete {
  background: #fff0f0;
  color: #ff4d4f;
  border: 1rpx solid #ffccc7;
}

.footer-tip {
  text-align: center;
  font-size: 24rpx;
  color: #9aa6b2;
  padding: 24rpx 0;
}
```

- [ ] **Step 4: 创建 myPublish.ts**

```typescript
// pages/myPublish/myPublish.ts
import { get, post } from '../../utils/request'

interface ProductItem {
  id: number
  title: string
  price: number
  images: string
  status: number
  create_time: string
  firstImage: string
  statusText: string
  statusClass: string
}

interface MyPublishData {
  products: ProductItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  hasMore: boolean
}

const STATUS_MAP: Record<number, { text: string; cls: string }> = {
  0: { text: '在售', cls: 'status-on-sale' },
  1: { text: '已售出', cls: 'status-sold' },
  2: { text: '已下架', cls: 'status-off' }
}

function processProducts(list: any[]): ProductItem[] {
  return list.map(item => ({
    ...item,
    firstImage: item.images ? item.images.split(',')[0] : '',
    statusText: STATUS_MAP[item.status]?.text ?? '未知',
    statusClass: STATUS_MAP[item.status]?.cls ?? ''
  }))
}

Page<MyPublishData, WechatMiniprogram.IAnyObject>({
  data: {
    products: [],
    total: 0,
    page: 1,
    pageSize: 10,
    loading: false,
    hasMore: false
  },

  onLoad() {
    this.loadProducts()
  },

  onShow() {
    this.setData({ products: [], page: 1, hasMore: false })
    this.loadProducts()
  },

  onPullDownRefresh() {
    this.setData({ products: [], page: 1, hasMore: false })
    this.loadProducts().finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadProducts() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      const res = await get('/api/product/mine', { page: 1, page_size: this.data.pageSize })
      const { list, total } = res.data
      const products = processProducts(list || [])
      this.setData({
        products,
        total,
        page: 1,
        hasMore: products.length < total
      })
    } catch (err) {
      console.error('加载失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadMore() {
    if (this.data.loading) return
    const nextPage = this.data.page + 1
    this.setData({ loading: true })
    try {
      const res = await get('/api/product/mine', { page: nextPage, page_size: this.data.pageSize })
      const { list } = res.data
      const more = processProducts(list || [])
      const all = [...this.data.products, ...more]
      this.setData({
        products: all,
        page: nextPage,
        hasMore: all.length < this.data.total
      })
    } catch (err) {
      console.error('加载更多失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  onEditTap(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/productEdit/productEdit?id=${id}` })
  },

  onMarkSold(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认',
      content: '确认将此商品标记为已售出？',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 1 })
          .then(() => {
            wx.showToast({ title: '已标记为售出', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  },

  onDelist(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认下架',
      content: '下架后商品将不再展示，确认下架？',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 2 })
          .then(() => {
            wx.showToast({ title: '已下架', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  },

  onDelete(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认删除',
      content: '删除后不可恢复，确认删除此商品？',
      confirmColor: '#ff4d4f',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 2 })
          .then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  }
})
```

- [ ] **Step 5: Commit**

```bash
git add mall-mini/miniprogram/pages/myPublish/
git commit -m "feat: add myPublish list page"
```

---

## Task 8: 前端 — productEdit 编辑页

**Files:**
- Create: `mall-mini/miniprogram/pages/productEdit/productEdit.json`
- Create: `mall-mini/miniprogram/pages/productEdit/productEdit.wxml`
- Create: `mall-mini/miniprogram/pages/productEdit/productEdit.wxss`
- Create: `mall-mini/miniprogram/pages/productEdit/productEdit.ts`

- [ ] **Step 1: 创建 productEdit.json**

```json
{
  "navigationBarTitleText": "编辑商品"
}
```

- [ ] **Step 2: 创建 productEdit.wxml**

```xml
<!--pages/productEdit/productEdit.wxml-->
<view class="publish-container">
  <!-- 图片上传区域 -->
  <view class="upload-section card">
    <view class="section-title">商品图片</view>
    <view class="image-grid">
      <block wx:for="{{images}}" wx:key="index">
        <view class="image-item">
          <image
            class="preview-image"
            src="{{item.localPath}}"
            mode="aspectFill"
            bindtap="previewImage"
            data-index="{{index}}"
          />
          <view class="upload-status" wx:if="{{item.uploading}}">
            <text>上传中</text>
          </view>
          <view class="delete-btn" catchtap="deleteImage" data-index="{{index}}">
            <text>×</text>
          </view>
        </view>
      </block>
      <view class="add-image-btn" wx:if="{{images.length < maxImages}}" bindtap="chooseImage">
        <text class="add-icon">+</text>
        <text class="add-text">添加图片</text>
      </view>
    </view>
    <view class="image-tip">最多上传{{maxImages}}张，支持jpg、png格式</view>
  </view>

  <!-- 商品描述 -->
  <view class="form-section card">
    <view class="section-title">商品描述</view>
    <textarea
      class="description-input"
      placeholder="描述一下你要出售的物品（品牌、型号、新旧程度等）"
      value="{{description}}"
      bindinput="onDescriptionInput"
      maxlength="500"
      auto-height
    />
    <view class="char-count">{{description.length}}/500</view>
  </view>

  <!-- 价格 -->
  <view class="form-section card">
    <view class="section-title">价格</view>
    <view class="price-input-wrapper">
      <text class="price-symbol">¥</text>
      <input
        class="price-input"
        type="digit"
        placeholder="0.00"
        value="{{price}}"
        bindinput="onPriceInput"
      />
    </view>
  </view>

  <!-- 分类 -->
  <view class="form-section card">
    <view class="section-title">商品分类</view>
    <picker
      mode="selector"
      range="{{categories}}"
      value="{{categoryIndex}}"
      bindchange="onCategoryChange"
    >
      <view class="picker-display">
        <text>{{categories[categoryIndex]}}</text>
        <text class="picker-arrow">›</text>
      </view>
    </picker>
  </view>

  <!-- 地点 -->
  <view class="form-section card">
    <view class="section-title">交易地点</view>
    <picker
      mode="multiSelector"
      range="{{regionNames}}"
      value="{{regionIndexes}}"
      bindcolumnchange="onRegionColumnChange"
      bindchange="onRegionChange"
    >
      <view class="picker-display">
        <text class="{{location ? '' : 'picker-placeholder'}}">{{location || '请选择省/市/区县'}}</text>
        <text class="picker-arrow">›</text>
      </view>
    </picker>
  </view>

  <!-- 提交按钮 -->
  <view class="submit-section">
    <button
      class="submit-btn {{submitting ? 'disabled' : ''}}"
      bindtap="submitForm"
      disabled="{{submitting}}"
    >
      {{submitting ? '保存中...' : '保存修改'}}
    </button>
  </view>
</view>
```

- [ ] **Step 3: 创建 productEdit.wxss（与 publish.wxss 完全相同，复制过来）**

```css
/* pages/productEdit/productEdit.wxss */
.publish-container {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 180rpx;
}

.card {
  background: #fff;
  margin: 18rpx 16rpx;
  padding: 20rpx;
  border-radius: 16rpx;
  box-shadow: 0 8rpx 24rpx rgba(12, 18, 29, 0.06);
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #111;
  margin-bottom: 16rpx;
}

.image-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  overflow: hidden;
}

.preview-image {
  width: 100%;
  height: 100%;
  border-radius: 12rpx;
  background: #f5f5f5;
}

.upload-status {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: #fff;
  font-size: 24rpx;
}

.delete-btn {
  position: absolute;
  top: -12rpx;
  right: -12rpx;
  width: 44rpx;
  height: 44rpx;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.delete-btn text {
  color: #fff;
  font-size: 28rpx;
  line-height: 1;
}

.add-image-btn {
  width: 200rpx;
  height: 200rpx;
  border: 2rpx dashed #e6e9ee;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #fff;
}

.add-icon { font-size: 48rpx; color: #bfc6cc; }
.add-text { font-size: 24rpx; color: #95a0aa; margin-top: 8rpx; }
.image-tip { font-size: 24rpx; color: #9aa6b2; margin-top: 12rpx; }

.description-input {
  width: 100%;
  min-height: 200rpx;
  font-size: 28rpx;
  line-height: 1.6;
}

.char-count {
  text-align: right;
  font-size: 24rpx;
  color: #9aa6b2;
  margin-top: 12rpx;
}

.price-input-wrapper { display: flex; align-items: center; }
.price-symbol { font-size: 36rpx; color: #ff4d4f; margin-right: 12rpx; }
.price-input { flex: 1; font-size: 40rpx; font-weight: 700; color: #111; }

.picker-display {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18rpx 0;
}

.picker-arrow { font-size: 32rpx; color: #c6ced6; }
.picker-placeholder { color: #c0c4cc; }

.submit-section {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  padding: 20rpx 18rpx;
  background: linear-gradient(180deg, rgba(255,255,255,0.9), #fff);
  box-shadow: 0 -8rpx 24rpx rgba(12,18,29,0.06);
}

.submit-btn {
  width: 100%;
  height: 88rpx;
  background: #07C160;
  color: #fff;
  font-size: 34rpx;
  font-weight: 700;
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.submit-btn:active { background: #06ad56; }
.submit-btn.disabled { background: #cdd7df; }
```

- [ ] **Step 4: 创建 productEdit.ts**

```typescript
// pages/productEdit/productEdit.ts
import { get, put } from '../../utils/request'
import { uploadToQiniu } from '../../utils/qiniu-upload'
import regionsData from '../../data/china-regions'

interface UploadedImage {
  localPath: string
  remoteUrl?: string
  uploading?: boolean
}

interface EditData {
  productId: number
  images: UploadedImage[]
  maxImages: number
  description: string
  price: string
  location: string
  regionNames: string[][]
  regionIndexes: number[]
  categoryIndex: number
  categories: string[]
  submitting: boolean
}

const provinceNames = regionsData.map(p => p.name)

function getCityNames(pi: number): string[] {
  return regionsData[pi].children.map(c => c.name)
}

function getDistrictNames(pi: number, ci: number): string[] {
  return regionsData[pi].children[ci].children as string[]
}

function buildInitialRegionNames(): string[][] {
  return [provinceNames, getCityNames(0), getDistrictNames(0, 0)]
}

// 将 location 字符串解析回省/市/区三级索引
function parseLocationToIndexes(locationStr: string): number[] {
  for (let pi = 0; pi < regionsData.length; pi++) {
    const province = regionsData[pi]
    if (!locationStr.startsWith(province.name)) continue
    const afterProvince = locationStr.slice(province.name.length)
    const cities = province.children
    for (let ci = 0; ci < cities.length; ci++) {
      const city = cities[ci]
      // 直辖市：province.name === city.name，location = province + district
      if (province.name === city.name) {
        const districts = city.children as string[]
        for (let di = 0; di < districts.length; di++) {
          if (afterProvince === districts[di]) return [pi, ci, di]
        }
        return [pi, ci, 0]
      }
      if (!afterProvince.startsWith(city.name)) continue
      const afterCity = afterProvince.slice(city.name.length)
      const districts = city.children as string[]
      for (let di = 0; di < districts.length; di++) {
        if (afterCity === districts[di]) return [pi, ci, di]
      }
      return [pi, ci, 0]
    }
  }
  return [0, 0, 0]
}

const CATEGORIES = ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他']

Page<EditData, WechatMiniprogram.IAnyObject>({
  data: {
    productId: 0,
    images: [],
    maxImages: 9,
    description: '',
    price: '',
    location: '',
    regionNames: buildInitialRegionNames(),
    regionIndexes: [0, 0, 0],
    categoryIndex: 0,
    categories: CATEGORIES,
    submitting: false
  },

  async onLoad(options: Record<string, string>) {
    const id = Number(options.id)
    if (!id) {
      wx.showToast({ title: '商品不存在', icon: 'none' })
      return
    }
    this.setData({ productId: id })
    await this.loadProduct(id)
  },

  async loadProduct(id: number) {
    wx.showLoading({ title: '加载中...', mask: true })
    try {
      const res = await get(`/api/product/detail?id=${id}`)
      const p = res.data
      // 图片回填：已上传的 remote URL 用 localPath 展示，remoteUrl 标记已上传
      const images: UploadedImage[] = p.images
        ? p.images.split(',').filter(Boolean).map((url: string) => ({
            localPath: url,
            remoteUrl: url
          }))
        : []

      // location 解析回索引
      const [pi, ci, di] = parseLocationToIndexes(p.location || '')
      const regionNames = [provinceNames, getCityNames(pi), getDistrictNames(pi, ci)]

      // 分类匹配
      const categoryIndex = Math.max(0, CATEGORIES.indexOf(p.category || ''))

      this.setData({
        images,
        description: p.description || '',
        price: p.price ? String(p.price) : '',
        location: p.location || '',
        regionNames,
        regionIndexes: [pi, ci, di],
        categoryIndex
      })
    } catch (err) {
      wx.showToast({ title: '加载商品失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

  async chooseImage() {
    const { images, maxImages } = this.data
    const remaining = maxImages - images.length
    if (remaining <= 0) {
      wx.showToast({ title: `最多上传${maxImages}张图片`, icon: 'none' })
      return
    }
    try {
      const res = await wx.chooseMedia({
        count: remaining,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed']
      })
      const newImages: UploadedImage[] = res.tempFiles.map(f => ({
        localPath: f.tempFilePath,
        uploading: false
      }))
      this.setData({ images: [...images, ...newImages] })
    } catch (_) {}
  },

  deleteImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const images = [...this.data.images]
    images.splice(index, 1)
    this.setData({ images })
  },

  previewImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const urls = this.data.images.map(img => img.localPath)
    wx.previewImage({ current: urls[index], urls })
  },

  onDescriptionInput(e: WechatMiniprogram.InputEvent) {
    this.setData({ description: e.detail.value })
  },

  onPriceInput(e: WechatMiniprogram.InputEvent) {
    const formatted = e.detail.value
      .replace(/[^\d.]/g, '')
      .replace(/\.{2,}/g, '.')
      .replace(/^(\d+\.\d{2}).*$/, '$1')
    this.setData({ price: formatted })
  },

  onCategoryChange(e: WechatMiniprogram.PickerChange) {
    this.setData({ categoryIndex: Number(e.detail.value) })
  },

  onRegionColumnChange(e: WechatMiniprogram.PickerColumnChange) {
    const { column, value } = e.detail
    const indexes = [...this.data.regionIndexes]
    indexes[column] = value
    if (column === 0) {
      indexes[1] = 0
      indexes[2] = 0
      this.setData({
        regionIndexes: indexes,
        'regionNames[1]': getCityNames(value),
        'regionNames[2]': getDistrictNames(value, 0)
      })
    } else if (column === 1) {
      indexes[2] = 0
      this.setData({
        regionIndexes: indexes,
        'regionNames[2]': getDistrictNames(indexes[0], value)
      })
    } else {
      this.setData({ regionIndexes: indexes })
    }
  },

  onRegionChange(e: WechatMiniprogram.PickerChange) {
    const [pi, ci, di] = e.detail.value as number[]
    const province = regionsData[pi].name
    const city = regionsData[pi].children[ci].name
    const district = (regionsData[pi].children[ci].children as string[])[di]
    const location = province === city ? `${province}${district}` : `${province}${city}${district}`
    this.setData({ regionIndexes: [pi, ci, di], location })
  },

  validateForm(): boolean {
    const { images, description, price, location } = this.data
    if (images.length === 0) {
      wx.showToast({ title: '请至少上传一张图片', icon: 'none' })
      return false
    }
    if (!description.trim()) {
      wx.showToast({ title: '请填写商品描述', icon: 'none' })
      return false
    }
    if (!price) {
      wx.showToast({ title: '请填写价格', icon: 'none' })
      return false
    }
    if (!location.trim()) {
      wx.showToast({ title: '请选择交易地点', icon: 'none' })
      return false
    }
    return true
  },

  async uploadImages(): Promise<string[]> {
    const { images } = this.data
    const urls: string[] = []
    for (let i = 0; i < images.length; i++) {
      const img = images[i]
      if (img.remoteUrl) {
        urls.push(img.remoteUrl)
        continue
      }
      this.setData({ [`images[${i}].uploading`]: true })
      try {
        const result = await uploadToQiniu(img.localPath)
        urls.push(result.url)
        this.setData({
          [`images[${i}].remoteUrl`]: result.url,
          [`images[${i}].uploading`]: false
        })
      } catch (err) {
        this.setData({ [`images[${i}].uploading`]: false })
        throw err
      }
    }
    return urls
  },

  async submitForm() {
    if (!this.validateForm()) return
    if (this.data.submitting) return
    this.setData({ submitting: true })
    wx.showLoading({ title: '保存中...', mask: true })
    try {
      const imageUrls = await this.uploadImages()
      await put('/api/product/update', {
        id: this.data.productId,
        description: this.data.description,
        price: parseFloat(this.data.price),
        location: this.data.location,
        images: imageUrls
      })
      wx.hideLoading()
      wx.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => wx.navigateBack(), 1500)
    } catch (err) {
      wx.hideLoading()
      wx.showToast({ title: '保存失败，请重试', icon: 'none' })
      console.error('保存失败:', err)
    } finally {
      this.setData({ submitting: false })
    }
  }
})
```

- [ ] **Step 5: Commit**

```bash
git add mall-mini/miniprogram/pages/productEdit/
git commit -m "feat: add productEdit page with data prefill and image handling"
```

---

## 完成验收

- [ ] 微信开发者工具中打开"我的" → 点击"我发布的" → 显示商品列表（含状态标签）
- [ ] 点击在售商品的"编辑" → 进入编辑页，所有字段已回填（图片、描述、价格、地点）
- [ ] 修改任意字段后点"保存修改" → 返回列表，列表数据已更新
- [ ] 点击"标记已售" → 商品状态变为已售出（红色标签）
- [ ] 点击"下架" → 商品状态变为已下架（灰色标签），操作栏变为"删除"
- [ ] 点击"删除" → 确认后商品从列表消失
