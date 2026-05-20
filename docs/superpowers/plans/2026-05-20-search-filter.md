# 商品搜索分类与地区筛选 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 首页和搜索页支持按商品分类与发布地区（精确到省/市/县区）筛选商品，同时修复 category 字段从未写入数据库的问题。

**Architecture:** 后端 `product` 表新增 `category`/`province`/`city`/`district` 四列（GORM AutoMigrate 自动建），搜索 API 新增四个可选过滤参数，发布/编辑接口同步写入。前端首页和搜索页各增一行筛选芯片栏，点击分别弹出分类列表面板和三级地区面板。

**Tech Stack:** Go/GORM/Gin（后端），TypeScript/WeChat Mini Program（前端），`china-regions.ts` 地区数据。

---

### Task 1: 后端数据层 — entity & repo

**Files:**
- Modify: `mall-server/internal/app/dao/product.entity.go`
- Modify: `mall-server/internal/app/dao/product.repo.go`

- [ ] **Step 1: 在 Product 实体新增四字段**

打开 `mall-server/internal/app/dao/product.entity.go`，在 `ContactType` 字段之前插入四个新字段，使文件变为：

```go
package dao

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Title        string  `gorm:"column:title;type:varchar(200);not null;default:''" json:"title" comment:"商品标题"`
	Description  string  `gorm:"column:description;type:text;not null;default:''" json:"description" comment:"商品描述"`
	Price        float64 `gorm:"column:price;type:decimal(10,2);not null;default:0" json:"price" comment:"价格"`
	Images       string  `gorm:"column:images;type:varchar(1000);not null;default:''" json:"images" comment:"图片URL列表,逗号分隔"`
	Location     string  `gorm:"column:location;type:varchar(100);not null;default:''" json:"location" comment:"交易地点"`
	Status       int     `gorm:"column:status;type:int;not null;default:0" json:"status" comment:"状态:0在售,1已售出,2已下架"`
	UserId       uint    `gorm:"column:user_id;type:int;not null;default:0" json:"user_id" comment:"发布者ID"`
	BuyUid       uint    `gorm:"column:buy_uid;type:int;not null;default:0" json:"buy_uid" comment:"购买者ID,0表示未售出"`
	Category     string  `gorm:"column:category;type:varchar(50);not null;default:''" json:"category" comment:"商品分类"`
	Province     string  `gorm:"column:province;type:varchar(50);not null;default:''" json:"province" comment:"省"`
	City         string  `gorm:"column:city;type:varchar(50);not null;default:''" json:"city" comment:"市"`
	District     string  `gorm:"column:district;type:varchar(50);not null;default:''" json:"district" comment:"县区"`
	ContactType  string  `gorm:"column:contact_type;type:varchar(10);not null;default:''" json:"contact_type" comment:"联系方式类型:phone/wechat/qq"`
	ContactValue string  `gorm:"column:contact_value;type:varchar(100);not null;default:''" json:"contact_value" comment:"联系方式值"`
}

func (Product) TableName() string {
	return "product"
}
```

- [ ] **Step 2: 在 ProductSearchResult 新增 Category 字段**

打开 `mall-server/internal/app/dao/product.repo.go`，将 `type ProductSearchResult struct` 替换为：

```go
type ProductSearchResult struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Images     string  `json:"images"`
	Location   string  `json:"location"`
	Status     int     `json:"status"`
	Category   string  `json:"category"`
	Seller     string  `json:"seller"`
	Avatar     string  `json:"avatar"`
	BuyUid     uint    `json:"buy_uid"`
	CreateTime string  `json:"create_time"`
}
```

- [ ] **Step 3: 在 ProductDetail 新增四字段**

将 `type ProductDetail struct` 替换为：

```go
type ProductDetail struct {
	ID           uint    `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Images       string  `json:"images"`
	Location     string  `json:"location"`
	Status       int     `json:"status"`
	BuyUid       uint    `json:"buy_uid"`
	Category     string  `json:"category"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Seller       string  `json:"seller"`
	Avatar       string  `json:"avatar"`
	CreateTime   string  `json:"create_time"`
	ContactType  string  `json:"contact_type"`
	ContactValue string  `json:"contact_value"`
	IsFavorited  bool    `json:"is_favorited"`
}
```

- [ ] **Step 4: 更新 SearchProducts 函数签名与查询**

将 `func SearchProducts(...)` 完整替换为：

```go
func SearchProducts(db *gorm.DB, keyword, sort string, status *int,
	category, province, city, district string,
	page, pageSize int) ([]ProductSearchResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	query := db.Model(&Product{}).
		Select("product.id, product.title, product.price, product.images, product.location, product.status, product.category, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id")

	if keyword != "" {
		query = query.Where("product.title LIKE ?", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("product.status = ?", *status)
	}
	if category != "" {
		query = query.Where("product.category = ?", category)
	}
	if province != "" {
		query = query.Where("product.province = ?", province)
	}
	if city != "" {
		query = query.Where("product.city = ?", city)
	}
	if district != "" {
		query = query.Where("product.district = ?", district)
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

- [ ] **Step 5: 更新 GetProductByID 的 SELECT**

将 `GetProductByID` 中的 `.Select(...)` 调用改为追加四列：

```go
func GetProductByID(db *gorm.DB, id uint) (*ProductDetail, error) {
	var detail ProductDetail
	err := db.Model(&Product{}).
		Select("product.id, product.title, product.description, product.price, product.images, product.location, product.status, product.buy_uid, product.category, product.province, product.city, product.district, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time, product.contact_type, product.contact_value").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id").
		Where("product.id = ?", id).
		First(&detail).Error

	if err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	return &detail, nil
}
```

- [ ] **Step 6: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误输出。

- [ ] **Step 7: Commit**

```bash
git add mall-server/internal/app/dao/product.entity.go \
        mall-server/internal/app/dao/product.repo.go
git commit -m "feat: add category/province/city/district to product entity and search query"
```

---

### Task 2: 后端服务层 — DTO & Handler

**Files:**
- Modify: `mall-server/internal/app/service/types.go`
- Modify: `mall-server/internal/app/service/product.go`

- [ ] **Step 1: 更新 SearchProductRequest**

打开 `mall-server/internal/app/service/types.go`，将 `type SearchProductRequest struct` 替换为：

```go
type SearchProductRequest struct {
	Keyword  string `form:"keyword"`
	Sort     string `form:"sort"`
	Status   *int   `form:"status"`
	Category string `form:"category"`
	Province string `form:"province"`
	City     string `form:"city"`
	District string `form:"district"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
```

- [ ] **Step 2: 更新 PublishProductRequest**

将 `type PublishProductRequest struct` 替换为：

```go
type PublishProductRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Price        float64  `json:"price" binding:"required"`
	Location     string   `json:"location" binding:"required"`
	Category     string   `json:"category"`
	Province     string   `json:"province"`
	City         string   `json:"city"`
	District     string   `json:"district"`
	Images       []string `json:"images" binding:"required"`
	ContactType  string   `json:"contact_type" binding:"required,oneof=phone wechat qq"`
	ContactValue string   `json:"contact_value" binding:"required,max=100"`
}
```

- [ ] **Step 3: 更新 UpdateProductRequest**

将 `type UpdateProductRequest struct` 替换为：

```go
type UpdateProductRequest struct {
	ID           uint     `json:"id" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Price        float64  `json:"price" binding:"required"`
	Location     string   `json:"location" binding:"required"`
	Category     string   `json:"category"`
	Province     string   `json:"province"`
	City         string   `json:"city"`
	District     string   `json:"district"`
	Images       []string `json:"images" binding:"required"`
	ContactType  string   `json:"contact_type" binding:"required,oneof=phone wechat qq"`
	ContactValue string   `json:"contact_value" binding:"required,max=100"`
}
```

- [ ] **Step 4: 更新 SearchProducts handler**

打开 `mall-server/internal/app/service/product.go`，将 `SearchProducts` handler 中 `dao.SearchProducts(...)` 的调用行替换为：

```go
results, total, err := dao.SearchProducts(svc.DB,
    req.Keyword, req.Sort, req.Status,
    req.Category, req.Province, req.City, req.District,
    req.Page, req.PageSize)
```

- [ ] **Step 5: 更新 PublishProduct handler 写入新字段**

在 `PublishProduct` handler 中，将 `product := dao.Product{...}` 替换为：

```go
product := dao.Product{
    Title:        req.Title,
    Description:  req.Description,
    Price:        req.Price,
    Images:       imagesStr,
    Location:     req.Location,
    Status:       0,
    UserId:       userID.(uint),
    BuyUid:       0,
    Category:     req.Category,
    Province:     req.Province,
    City:         req.City,
    District:     req.District,
    ContactType:  req.ContactType,
    ContactValue: req.ContactValue,
}
```

- [ ] **Step 6: 更新 UpdateProduct handler 写入新字段**

在 `UpdateProduct` handler 中，将 `updates := map[string]interface{}{...}` 替换为：

```go
updates := map[string]interface{}{
    "title":         title,
    "description":   req.Description,
    "price":         req.Price,
    "location":      req.Location,
    "images":        strings.Join(req.Images, ","),
    "category":      req.Category,
    "province":      req.Province,
    "city":          req.City,
    "district":      req.District,
    "contact_type":  req.ContactType,
    "contact_value": req.ContactValue,
}
```

- [ ] **Step 7: 编译 & 测试**

```bash
cd mall-server && go build ./... && go test ./...
```

Expected: 编译无错误，`ok  mall-server/...`。

- [ ] **Step 8: Commit**

```bash
git add mall-server/internal/app/service/types.go \
        mall-server/internal/app/service/product.go
git commit -m "feat: add category/province/city/district to search, publish, and update APIs"
```

---

### Task 3: 前端发布页 & 编辑页 — 提交时携带省市区与分类

**Files:**
- Modify: `mall-mini/miniprogram/pages/publish/publish.ts`
- Modify: `mall-mini/miniprogram/pages/productEdit/productEdit.ts`

- [ ] **Step 1: 更新 publish.ts 的 submitForm()**

打开 `mall-mini/miniprogram/pages/publish/publish.ts`，找到 `submitForm()` 中的 `const productData = { ... }` 部分，替换为：

```typescript
const [pi, ci, di] = this.data.regionIndexes
const province = regionsData[pi].name
const city = regionsData[pi].children[ci].name
const district = (regionsData[pi].children[ci].children as string[])[di]

const productData = {
  title: this.data.description.substring(0, 50),
  description: this.data.description,
  price: parseFloat(this.data.price),
  location: this.data.location,
  category: this.data.categories[this.data.categoryIndex],
  province,
  city,
  district,
  images: imageUrls,
  contact_type: this.data.contactType,
  contact_value: this.data.contactValue,
}
```

- [ ] **Step 2: 更新 productEdit.ts 的 submitForm()**

打开 `mall-mini/miniprogram/pages/productEdit/productEdit.ts`，找到 `submitForm()` 中的 `await put('/api/product/update', { ... })` 调用，替换为：

```typescript
const [pi, ci, di] = this.data.regionIndexes
const province = regionsData[pi].name
const city = regionsData[pi].children[ci].name
const district = (regionsData[pi].children[ci].children as string[])[di]

await put('/api/product/update', {
  id: this.data.productId,
  description: this.data.description,
  price: parseFloat(this.data.price),
  location: this.data.location,
  category: CATEGORIES[this.data.categoryIndex],
  province,
  city,
  district,
  images: imageUrls,
  contact_type: this.data.contactType,
  contact_value: this.data.contactValue,
})
```

- [ ] **Step 3: 编译验证**

在 WeChat Developer Tools 打开项目，查看编译输出面板。`publish.ts` 和 `productEdit.ts` 无红色 TypeScript 错误。

- [ ] **Step 4: Commit**

```bash
git add mall-mini/miniprogram/pages/publish/publish.ts \
        mall-mini/miniprogram/pages/productEdit/productEdit.ts
git commit -m "feat: send province/city/district and fix category in publish/edit payloads"
```

---

### Task 4: 前端首页 — 筛选芯片栏

**Files:**
- Modify: `mall-mini/miniprogram/pages/home/home.ts`
- Modify: `mall-mini/miniprogram/pages/home/home.wxml`
- Modify: `mall-mini/miniprogram/pages/home/home.wxss`

- [ ] **Step 1: 更新 home.ts — 新增 import 和接口字段**

打开 `mall-mini/miniprogram/pages/home/home.ts`，在文件顶部 `import { get } from '../../utils/request'` 下方新增一行：

```typescript
import regionsData from '../../data/china-regions'
```

将 `interface HomeData` 替换为：

```typescript
interface HomeData {
  products: ProductItem[]
  loading: boolean
  page: number
  hasMore: boolean
  selectedCategory: string
  selectedProvince: string
  selectedCity: string
  selectedDistrict: string
  showCategoryPanel: boolean
  showRegionPanel: boolean
  regionStep: number
  regionProvinceIndex: number
  regionCityIndex: number
  regionCities: string[]
  regionDistricts: string[]
  provinces: string[]
  categories: string[]
}
```

- [ ] **Step 2: 更新 home.ts — data 初始值**

将 `data: { ... }` 替换为：

```typescript
data: {
  products: [],
  loading: false,
  page: 1,
  hasMore: true,
  selectedCategory: '',
  selectedProvince: '',
  selectedCity: '',
  selectedDistrict: '',
  showCategoryPanel: false,
  showRegionPanel: false,
  regionStep: 0,
  regionProvinceIndex: 0,
  regionCityIndex: 0,
  regionCities: [],
  regionDistricts: [],
  provinces: regionsData.map(p => p.name),
  categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
},
```

- [ ] **Step 3: 更新 home.ts — loadProducts() 携带筛选参数**

将 `loadProducts()` 中的 `const response = await get<...>(...)` 调用替换为：

```typescript
const params: Record<string, string | number> = {
  page: this.data.page,
  page_size: 10,
  status: 0,
}
if (this.data.selectedCategory) params.category = this.data.selectedCategory
if (this.data.selectedProvince) params.province = this.data.selectedProvince
if (this.data.selectedCity)     params.city     = this.data.selectedCity
if (this.data.selectedDistrict) params.district = this.data.selectedDistrict

const response = await get<{ list: any[], total: number, page: number, page_size: number }>(
  '/api/product/search',
  params
)
```

- [ ] **Step 4: 更新 home.ts — 新增筛选事件处理器**

在 `onSearch()` 方法之后添加以下所有方法（保留 `onSearch`，在其后追加）：

```typescript
onOpenCategoryPanel() {
  this.setData({ showCategoryPanel: true })
},

onCloseCategoryPanel() {
  this.setData({ showCategoryPanel: false })
},

onSelectCategory(e: WechatMiniprogram.TouchEvent) {
  const value: string = e.currentTarget.dataset.value
  this.setData({ selectedCategory: value, showCategoryPanel: false, page: 1, hasMore: true, products: [] })
  this.loadProducts()
},

onClearCategory(_e: WechatMiniprogram.TouchEvent) {
  this.setData({ selectedCategory: '', page: 1, hasMore: true, products: [] })
  this.loadProducts()
},

onOpenRegionPanel() {
  this.setData({ showRegionPanel: true, regionStep: 0 })
},

onCloseRegionPanel() {
  this.setData({ showRegionPanel: false })
},

onSelectProvince(e: WechatMiniprogram.TouchEvent) {
  const pi: number = e.currentTarget.dataset.index
  const province = regionsData[pi]
  this.setData({
    selectedProvince: province.name,
    selectedCity: '',
    selectedDistrict: '',
    regionProvinceIndex: pi,
    regionCities: province.children.map((c: any) => c.name),
    regionStep: 1,
  })
},

onSelectCity(e: WechatMiniprogram.TouchEvent) {
  const ci: number = e.currentTarget.dataset.index
  const pi = this.data.regionProvinceIndex
  const city = regionsData[pi].children[ci]
  this.setData({
    selectedCity: city.name,
    selectedDistrict: '',
    regionCityIndex: ci,
    regionDistricts: city.children as string[],
    regionStep: 2,
  })
},

onSelectDistrict(e: WechatMiniprogram.TouchEvent) {
  const district: string = e.currentTarget.dataset.district
  this.setData({ selectedDistrict: district, showRegionPanel: false, page: 1, hasMore: true, products: [] })
  this.loadProducts()
},

onConfirmRegion() {
  this.setData({ showRegionPanel: false, page: 1, hasMore: true, products: [] })
  this.loadProducts()
},

onClearRegion(_e: WechatMiniprogram.TouchEvent) {
  this.setData({ selectedProvince: '', selectedCity: '', selectedDistrict: '', page: 1, hasMore: true, products: [] })
  this.loadProducts()
},
```

- [ ] **Step 5: 更新 home.wxml — 插入筛选栏和面板**

打开 `mall-mini/miniprogram/pages/home/home.wxml`，将整个文件替换为：

```xml
<!--pages/home/home.wxml-->
<view class="home-container">
  <!-- 搜索栏 -->
  <view class="search-bar" bindtap="onSearch">
    <view class="search-icon">
      <text class="iconfont">🔍</text>
    </view>
    <text class="search-placeholder">搜索商品</text>
  </view>

  <!-- 筛选芯片栏 -->
  <view class="chip-filter-bar">
    <view
      class="filter-chip {{selectedCategory ? 'active' : ''}}"
      bindtap="onOpenCategoryPanel"
    >
      <text>{{selectedCategory || '全部分类'}}</text>
      <text wx:if="{{selectedCategory}}" catchtap="onClearCategory"> ✕</text>
      <text wx:else> ▾</text>
    </view>
    <view
      class="filter-chip {{selectedProvince ? 'active' : ''}}"
      bindtap="onOpenRegionPanel"
    >
      <text>{{selectedDistrict || selectedCity || selectedProvince || '全部地区'}}</text>
      <text wx:if="{{selectedProvince}}" catchtap="onClearRegion"> ✕</text>
      <text wx:else> ▾</text>
    </view>
  </view>

  <!-- 商品列表 -->
  <view class="product-list">
    <block wx:for="{{products}}" wx:key="id">
      <view class="product-card" bindtap="goToDetail" data-id="{{item.id}}">
        <view class="img-wrap">
          <image class="product-image" src="{{item.images[0]}}" mode="aspectFill" />
          <view class="price-badge">¥{{item.price}}</view>
        </view>
        <view class="product-info">
          <view class="product-title">{{item.title}}</view>
          <view class="product-meta">
            <text class="product-location">📍 {{item.location}}</text>
            <text class="product-seller">卖家：{{item.seller}}</text>
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
  <view class="empty-state" wx:if="{{!loading && products.length === 0}}">
    <text class="empty-text">暂无商品</text>
  </view>

  <!-- 分类面板遮罩 -->
  <view wx:if="{{showCategoryPanel}}" class="panel-mask" bindtap="onCloseCategoryPanel" />
  <!-- 分类面板 -->
  <view wx:if="{{showCategoryPanel}}" class="category-panel">
    <view
      class="category-item {{!selectedCategory ? 'selected' : ''}}"
      bindtap="onSelectCategory"
      data-value=""
    >全部</view>
    <view
      wx:for="{{categories}}"
      wx:key="index"
      class="category-item {{item === selectedCategory ? 'selected' : ''}}"
      bindtap="onSelectCategory"
      data-value="{{item}}"
    >{{item}}</view>
  </view>

  <!-- 地区面板遮罩 -->
  <view wx:if="{{showRegionPanel}}" class="panel-mask" bindtap="onCloseRegionPanel" />
  <!-- 地区面板 -->
  <view wx:if="{{showRegionPanel}}" class="region-panel">
    <view class="region-header">
      <text class="region-breadcrumb">
        {{selectedProvince || '请选择省份'}}{{selectedCity ? ' / ' + selectedCity : ''}}
      </text>
      <view wx:if="{{regionStep > 0}}" class="region-confirm-btn" bindtap="onConfirmRegion">确认</view>
    </view>
    <!-- 省列表 -->
    <scroll-view wx:if="{{regionStep === 0}}" scroll-y class="region-list">
      <view
        wx:for="{{provinces}}"
        wx:key="index"
        class="region-item {{item === selectedProvince ? 'selected' : ''}}"
        bindtap="onSelectProvince"
        data-index="{{index}}"
      >{{item}}</view>
    </scroll-view>
    <!-- 市列表 -->
    <scroll-view wx:if="{{regionStep === 1}}" scroll-y class="region-list">
      <view
        wx:for="{{regionCities}}"
        wx:key="index"
        class="region-item {{item === selectedCity ? 'selected' : ''}}"
        bindtap="onSelectCity"
        data-index="{{index}}"
      >{{item}}</view>
    </scroll-view>
    <!-- 县区列表 -->
    <scroll-view wx:if="{{regionStep === 2}}" scroll-y class="region-list">
      <view
        wx:for="{{regionDistricts}}"
        wx:key="index"
        class="region-item {{item === selectedDistrict ? 'selected' : ''}}"
        bindtap="onSelectDistrict"
        data-district="{{item}}"
      >{{item}}</view>
    </scroll-view>
  </view>
</view>
```

- [ ] **Step 6: 更新 home.wxss — 新增筛选栏样式**

打开 `mall-mini/miniprogram/pages/home/home.wxss`，在文件末尾追加：

```css
/* 筛选芯片栏 */
.chip-filter-bar {
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-chip {
  display: flex;
  align-items: center;
  padding: 10rpx 20rpx;
  border: 2rpx solid #e0e0e0;
  border-radius: 32rpx;
  font-size: 26rpx;
  color: #666;
  background: #fafafa;
}

.filter-chip.active {
  border-color: #fa5151;
  color: #fa5151;
  background: #fff3f0;
}

/* 面板遮罩 */
.panel-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 99;
}

/* 分类面板 */
.category-panel {
  position: fixed;
  top: 200rpx;
  left: 0;
  right: 0;
  background: #fff;
  z-index: 100;
  padding: 8rpx 0;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.12);
}

.category-item {
  padding: 24rpx 32rpx;
  font-size: 28rpx;
  color: #333;
}

.category-item.selected {
  color: #fa5151;
  font-weight: 500;
}

/* 地区面板 */
.region-panel {
  position: fixed;
  top: 200rpx;
  left: 0;
  right: 0;
  height: 600rpx;
  background: #fff;
  z-index: 100;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.12);
  display: flex;
  flex-direction: column;
}

.region-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
  flex-shrink: 0;
}

.region-breadcrumb {
  font-size: 26rpx;
  color: #666;
}

.region-confirm-btn {
  font-size: 28rpx;
  color: #fa5151;
  padding: 8rpx 16rpx;
}

.region-list {
  flex: 1;
}

.region-item {
  padding: 24rpx 32rpx;
  font-size: 28rpx;
  color: #333;
}

.region-item.selected {
  color: #fa5151;
  font-weight: 500;
}
```

- [ ] **Step 7: 手动验证**

在 WeChat DevTools 打开首页：
- 点击"全部分类"芯片 → 弹出分类列表，选"电子产品" → 芯片高亮，列表刷新
- 点击"全部地区"芯片 → 弹出省级列表，选一个省 → 自动进入市级列表，选一个市 → 进入县区列表，选一个区 → 面板关闭，芯片显示县区名，列表刷新
- 点击地区芯片 → 选到省级点"确认" → 面板关闭，芯片显示省名，列表刷新
- 点击分类芯片 ✕ → 清除分类筛选，列表刷新
- 点击遮罩 → 面板关闭，不改变已选值

- [ ] **Step 8: Commit**

```bash
git add mall-mini/miniprogram/pages/home/home.ts \
        mall-mini/miniprogram/pages/home/home.wxml \
        mall-mini/miniprogram/pages/home/home.wxss
git commit -m "feat: add category and region filter chips to home page"
```

---

### Task 5: 前端搜索页 — 筛选芯片栏

**Files:**
- Modify: `mall-mini/miniprogram/pages/search/search.ts`
- Modify: `mall-mini/miniprogram/pages/search/search.wxml`
- Modify: `mall-mini/miniprogram/pages/search/search.wxss`

- [ ] **Step 1: 更新 search.ts — 新增 import 和接口字段**

打开 `mall-mini/miniprogram/pages/search/search.ts`，在 `import { get } from '../../utils/request'` 下方新增：

```typescript
import regionsData from '../../data/china-regions'
```

将 `interface SearchData` 替换为：

```typescript
interface SearchData {
  keyword: string
  sort: string
  status: number | null
  selectedCategory: string
  selectedProvince: string
  selectedCity: string
  selectedDistrict: string
  showCategoryPanel: boolean
  showRegionPanel: boolean
  regionStep: number
  regionProvinceIndex: number
  regionCityIndex: number
  regionCities: string[]
  regionDistricts: string[]
  provinces: string[]
  categories: string[]
  products: ProductItem[]
  loading: boolean
  page: number
  hasMore: boolean
  sortOptions: string[]
  statusOptions: string[]
  sortIndex: number
  statusIndex: number
}
```

- [ ] **Step 2: 更新 search.ts — data 初始值**

在 `data: { ... }` 中，在 `keyword: '',` 之后追加：

```typescript
selectedCategory: '',
selectedProvince: '',
selectedCity: '',
selectedDistrict: '',
showCategoryPanel: false,
showRegionPanel: false,
regionStep: 0,
regionProvinceIndex: 0,
regionCityIndex: 0,
regionCities: [],
regionDistricts: [],
provinces: regionsData.map(p => p.name),
categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
```

- [ ] **Step 3: 更新 search.ts — search() 携带筛选参数**

将 `search()` 中的 `const params: Record<string, string | number> = { ... }` 块替换为：

```typescript
const params: Record<string, string | number> = {
  keyword: this.data.keyword,
  sort: this.data.sort,
  page: this.data.page,
  page_size: 10,
}
if (this.data.status !== null) {
  params.status = this.data.status
}
if (this.data.selectedCategory) params.category = this.data.selectedCategory
if (this.data.selectedProvince) params.province = this.data.selectedProvince
if (this.data.selectedCity)     params.city     = this.data.selectedCity
if (this.data.selectedDistrict) params.district = this.data.selectedDistrict
```

- [ ] **Step 4: 更新 search.ts — 新增筛选事件处理器**

在 `goToDetail` 方法之前追加以下所有方法：

```typescript
onOpenCategoryPanel() {
  this.setData({ showCategoryPanel: true })
},

onCloseCategoryPanel() {
  this.setData({ showCategoryPanel: false })
},

onSelectCategory(e: WechatMiniprogram.TouchEvent) {
  const value: string = e.currentTarget.dataset.value
  this.setData({ selectedCategory: value, showCategoryPanel: false, page: 1, hasMore: true, products: [] })
  this.search()
},

onClearCategory(_e: WechatMiniprogram.TouchEvent) {
  this.setData({ selectedCategory: '', page: 1, hasMore: true, products: [] })
  this.search()
},

onOpenRegionPanel() {
  this.setData({ showRegionPanel: true, regionStep: 0 })
},

onCloseRegionPanel() {
  this.setData({ showRegionPanel: false })
},

onSelectProvince(e: WechatMiniprogram.TouchEvent) {
  const pi: number = e.currentTarget.dataset.index
  const province = regionsData[pi]
  this.setData({
    selectedProvince: province.name,
    selectedCity: '',
    selectedDistrict: '',
    regionProvinceIndex: pi,
    regionCities: province.children.map((c: any) => c.name),
    regionStep: 1,
  })
},

onSelectCity(e: WechatMiniprogram.TouchEvent) {
  const ci: number = e.currentTarget.dataset.index
  const pi = this.data.regionProvinceIndex
  const city = regionsData[pi].children[ci]
  this.setData({
    selectedCity: city.name,
    selectedDistrict: '',
    regionCityIndex: ci,
    regionDistricts: city.children as string[],
    regionStep: 2,
  })
},

onSelectDistrict(e: WechatMiniprogram.TouchEvent) {
  const district: string = e.currentTarget.dataset.district
  this.setData({ selectedDistrict: district, showRegionPanel: false, page: 1, hasMore: true, products: [] })
  this.search()
},

onConfirmRegion() {
  this.setData({ showRegionPanel: false, page: 1, hasMore: true, products: [] })
  this.search()
},

onClearRegion(_e: WechatMiniprogram.TouchEvent) {
  this.setData({ selectedProvince: '', selectedCity: '', selectedDistrict: '', page: 1, hasMore: true, products: [] })
  this.search()
},
```

- [ ] **Step 5: 更新 search.wxml — 插入筛选芯片栏和面板**

打开 `mall-mini/miniprogram/pages/search/search.wxml`，在 `<!-- 排序筛选栏 -->` 的 `<view class="filter-bar">` 之前插入：

```xml
<!-- 分类/地区筛选芯片栏 -->
<view class="chip-filter-bar">
  <view
    class="filter-chip {{selectedCategory ? 'active' : ''}}"
    bindtap="onOpenCategoryPanel"
  >
    <text>{{selectedCategory || '全部分类'}}</text>
    <text wx:if="{{selectedCategory}}" catchtap="onClearCategory"> ✕</text>
    <text wx:else> ▾</text>
  </view>
  <view
    class="filter-chip {{selectedProvince ? 'active' : ''}}"
    bindtap="onOpenRegionPanel"
  >
    <text>{{selectedDistrict || selectedCity || selectedProvince || '全部地区'}}</text>
    <text wx:if="{{selectedProvince}}" catchtap="onClearRegion"> ✕</text>
    <text wx:else> ▾</text>
  </view>
</view>
```

在文件末尾 `</view>` 之前（`search-container` 结束标签之前）插入面板：

```xml
  <!-- 分类面板遮罩 -->
  <view wx:if="{{showCategoryPanel}}" class="panel-mask" bindtap="onCloseCategoryPanel" />
  <!-- 分类面板 -->
  <view wx:if="{{showCategoryPanel}}" class="category-panel">
    <view
      class="category-item {{!selectedCategory ? 'selected' : ''}}"
      bindtap="onSelectCategory"
      data-value=""
    >全部</view>
    <view
      wx:for="{{categories}}"
      wx:key="index"
      class="category-item {{item === selectedCategory ? 'selected' : ''}}"
      bindtap="onSelectCategory"
      data-value="{{item}}"
    >{{item}}</view>
  </view>

  <!-- 地区面板遮罩 -->
  <view wx:if="{{showRegionPanel}}" class="panel-mask" bindtap="onCloseRegionPanel" />
  <!-- 地区面板 -->
  <view wx:if="{{showRegionPanel}}" class="region-panel">
    <view class="region-header">
      <text class="region-breadcrumb">
        {{selectedProvince || '请选择省份'}}{{selectedCity ? ' / ' + selectedCity : ''}}
      </text>
      <view wx:if="{{regionStep > 0}}" class="region-confirm-btn" bindtap="onConfirmRegion">确认</view>
    </view>
    <scroll-view wx:if="{{regionStep === 0}}" scroll-y class="region-list">
      <view
        wx:for="{{provinces}}"
        wx:key="index"
        class="region-item {{item === selectedProvince ? 'selected' : ''}}"
        bindtap="onSelectProvince"
        data-index="{{index}}"
      >{{item}}</view>
    </scroll-view>
    <scroll-view wx:if="{{regionStep === 1}}" scroll-y class="region-list">
      <view
        wx:for="{{regionCities}}"
        wx:key="index"
        class="region-item {{item === selectedCity ? 'selected' : ''}}"
        bindtap="onSelectCity"
        data-index="{{index}}"
      >{{item}}</view>
    </scroll-view>
    <scroll-view wx:if="{{regionStep === 2}}" scroll-y class="region-list">
      <view
        wx:for="{{regionDistricts}}"
        wx:key="index"
        class="region-item {{item === selectedDistrict ? 'selected' : ''}}"
        bindtap="onSelectDistrict"
        data-district="{{item}}"
      >{{item}}</view>
    </scroll-view>
  </view>
```

- [ ] **Step 6: 更新 search.wxss — 新增筛选栏样式**

打开 `mall-mini/miniprogram/pages/search/search.wxss`，在文件末尾追加（与 home.wxss 相同的样式块）：

```css
/* 筛选芯片栏 */
.chip-filter-bar {
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-chip {
  display: flex;
  align-items: center;
  padding: 10rpx 20rpx;
  border: 2rpx solid #e0e0e0;
  border-radius: 32rpx;
  font-size: 26rpx;
  color: #666;
  background: #fafafa;
}

.filter-chip.active {
  border-color: #fa5151;
  color: #fa5151;
  background: #fff3f0;
}

/* 面板遮罩 */
.panel-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 99;
}

/* 分类面板 */
.category-panel {
  position: fixed;
  top: 280rpx;
  left: 0;
  right: 0;
  background: #fff;
  z-index: 100;
  padding: 8rpx 0;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.12);
}

.category-item {
  padding: 24rpx 32rpx;
  font-size: 28rpx;
  color: #333;
}

.category-item.selected {
  color: #fa5151;
  font-weight: 500;
}

/* 地区面板 */
.region-panel {
  position: fixed;
  top: 280rpx;
  left: 0;
  right: 0;
  height: 600rpx;
  background: #fff;
  z-index: 100;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.12);
  display: flex;
  flex-direction: column;
}

.region-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
  flex-shrink: 0;
}

.region-breadcrumb {
  font-size: 26rpx;
  color: #666;
}

.region-confirm-btn {
  font-size: 28rpx;
  color: #fa5151;
  padding: 8rpx 16rpx;
}

.region-list {
  flex: 1;
}

.region-item {
  padding: 24rpx 32rpx;
  font-size: 28rpx;
  color: #333;
}

.region-item.selected {
  color: #fa5151;
  font-weight: 500;
}
```

- [ ] **Step 7: 手动验证**

在 WeChat DevTools 打开搜索页：
- 输入关键词 + 选分类 → 结果同时满足两个条件
- 地区三级选择流程正常（省 → 市 → 县区）
- 在市级点"确认" → 仅按省+市筛选
- ✕ 清除分类/地区筛选后列表还原
- 遮罩点击关闭面板，不改变已选值

- [ ] **Step 8: Commit**

```bash
git add mall-mini/miniprogram/pages/search/search.ts \
        mall-mini/miniprogram/pages/search/search.wxml \
        mall-mini/miniprogram/pages/search/search.wxss
git commit -m "feat: add category and region filter chips to search page"
```
