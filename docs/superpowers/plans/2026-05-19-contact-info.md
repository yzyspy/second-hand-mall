# 商品联系方式字段 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 卖家发布/编辑商品时必填联系方式（手机号/微信/QQ），买家在详情页以纯文字查看。

**Architecture:** 后端 `product` 表新增 `contact_type` + `contact_value` 两列，GORM AutoMigrate 自动建列。前端发布页/编辑页增加单选按钮+输入框，详情页新增联系方式展示区块，移除原"联系卖家"功能占位按钮。

**Tech Stack:** Go/GORM/Gin（后端），TypeScript/WeChat Mini Program（前端）。

---

### Task 1: 后端数据层 — 实体 & 查询

**Files:**
- Modify: `mall-server/internal/app/dao/product.entity.go`
- Modify: `mall-server/internal/app/dao/product.repo.go`

- [ ] **Step 1: 在 Product 实体中新增两个字段**

打开 `mall-server/internal/app/dao/product.entity.go`，在 `BuyUid` 字段之后插入：

```go
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
	ContactType  string  `gorm:"column:contact_type;type:varchar(10);not null;default:''" json:"contact_type" comment:"联系方式类型:phone/wechat/qq"`
	ContactValue string  `gorm:"column:contact_value;type:varchar(100);not null;default:''" json:"contact_value" comment:"联系方式值"`
}
```

- [ ] **Step 2: 在 ProductDetail 结构体中新增两个字段**

打开 `mall-server/internal/app/dao/product.repo.go`，找到 `type ProductDetail struct`，在 `IsFavorited` 字段之前插入：

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
	Seller       string  `json:"seller"`
	Avatar       string  `json:"avatar"`
	CreateTime   string  `json:"create_time"`
	ContactType  string  `json:"contact_type"`
	ContactValue string  `json:"contact_value"`
	IsFavorited  bool    `json:"is_favorited"`
}
```

- [ ] **Step 3: 更新 GetProductByID 的 SELECT 语句**

在同一文件中，找到 `GetProductByID` 函数，将 `Select(...)` 调用更新为（追加 `product.contact_type, product.contact_value`）：

```go
func GetProductByID(db *gorm.DB, id uint) (*ProductDetail, error) {
	var detail ProductDetail
	err := db.Model(&Product{}).
		Select("product.id, product.title, product.description, product.price, product.images, product.location, product.status, product.buy_uid, sys_user.nick_name as seller, sys_user.avatar, product.created_at as create_time, product.contact_type, product.contact_value").
		Joins("LEFT JOIN sys_user ON product.user_id = sys_user.id").
		Where("product.id = ?", id).
		First(&detail).Error

	if err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	return &detail, nil
}
```

- [ ] **Step 4: 编译验证**

```bash
cd mall-server && go build ./...
```

Expected: 无错误输出。

- [ ] **Step 5: Commit**

```bash
git add mall-server/internal/app/dao/product.entity.go mall-server/internal/app/dao/product.repo.go
git commit -m "feat: add contact_type and contact_value to product entity and detail query"
```

---

### Task 2: 后端服务层 — DTO & Handler

**Files:**
- Modify: `mall-server/internal/app/service/types.go`
- Modify: `mall-server/internal/app/service/product.go`

- [ ] **Step 1: 更新 PublishProductRequest**

打开 `mall-server/internal/app/service/types.go`，将 `PublishProductRequest` 替换为：

```go
type PublishProductRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Price        float64  `json:"price" binding:"required"`
	Location     string   `json:"location" binding:"required"`
	Category     string   `json:"category"`
	Images       []string `json:"images" binding:"required"`
	ContactType  string   `json:"contact_type" binding:"required,oneof=phone wechat qq"`
	ContactValue string   `json:"contact_value" binding:"required"`
}
```

- [ ] **Step 2: 更新 UpdateProductRequest**

在同一文件中，将 `UpdateProductRequest` 替换为：

```go
type UpdateProductRequest struct {
	ID           uint     `json:"id" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Price        float64  `json:"price" binding:"required"`
	Location     string   `json:"location" binding:"required"`
	Images       []string `json:"images" binding:"required"`
	ContactType  string   `json:"contact_type" binding:"required,oneof=phone wechat qq"`
	ContactValue string   `json:"contact_value" binding:"required"`
}
```

- [ ] **Step 3: 更新 PublishProduct handler 写入新字段**

打开 `mall-server/internal/app/service/product.go`，找到 `PublishProduct` 函数中构建 `product := dao.Product{...}` 的部分，替换为：

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
    ContactType:  req.ContactType,
    ContactValue: req.ContactValue,
}
```

- [ ] **Step 4: 更新 UpdateProduct handler 写入新字段**

在同一文件中，找到 `UpdateProduct` 函数的 `updates := map[string]interface{}{...}`，替换为：

```go
updates := map[string]interface{}{
    "title":         title,
    "description":   req.Description,
    "price":         req.Price,
    "location":      req.Location,
    "images":        strings.Join(req.Images, ","),
    "contact_type":  req.ContactType,
    "contact_value": req.ContactValue,
}
```

- [ ] **Step 5: 编译 & 测试**

```bash
cd mall-server && go build ./... && go test ./...
```

Expected: 编译无错误，`ok mall-server/pkg/utils`。

- [ ] **Step 6: Commit**

```bash
git add mall-server/internal/app/service/types.go mall-server/internal/app/service/product.go
git commit -m "feat: add contact_type/contact_value to publish and update DTOs and handlers"
```

---

### Task 3: 前端发布页 — publish.ts & publish.wxml & publish.wxss

**Files:**
- Modify: `mall-mini/miniprogram/pages/publish/publish.ts`
- Modify: `mall-mini/miniprogram/pages/publish/publish.wxml`
- Modify: `mall-mini/miniprogram/pages/publish/publish.wxss`

- [ ] **Step 1: 更新 PublishData interface，新增 contact 相关字段**

打开 `mall-mini/miniprogram/pages/publish/publish.ts`，将 `interface PublishData` 替换为：

```typescript
interface PublishData {
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
  contactType: 'phone' | 'wechat' | 'qq'
  contactValue: string
  contactTypes: string[]
}
```

- [ ] **Step 2: 更新 data 初始值**

在 `Page<PublishData, ...>({ data: { ... } })` 中，追加三行初始值（在 `submitting: false` 之后）：

```typescript
data: {
  images: [],
  maxImages: 9,
  description: '',
  price: '',
  location: '',
  regionNames: buildInitialRegionNames(),
  regionIndexes: [0, 0, 0],
  categoryIndex: 0,
  categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
  submitting: false,
  contactType: 'phone',
  contactValue: '',
  contactTypes: ['手机号', '微信', 'QQ'],
},
```

- [ ] **Step 3: 新增两个事件处理方法**

在 `onCategoryChange` 方法之后添加：

```typescript
onContactTypeSelect(e: WechatMiniprogram.TouchEvent) {
  const types: Array<'phone' | 'wechat' | 'qq'> = ['phone', 'wechat', 'qq']
  this.setData({ contactType: types[e.currentTarget.dataset.index] })
},

onContactValueInput(e: WechatMiniprogram.Input) {
  this.setData({ contactValue: e.detail.value })
},
```

- [ ] **Step 4: 在 validateForm() 末尾新增 contactValue 校验**

找到 `validateForm()` 函数，在 `return true` 之前插入：

```typescript
if (!this.data.contactValue.trim()) {
  wx.showToast({ title: '请填写联系方式', icon: 'none' })
  return false
}
```

- [ ] **Step 5: 在 submitForm() 的 productData 中加入联系方式字段**

找到 `const productData = { ... }` 部分，替换为：

```typescript
const productData = {
  title: this.data.description.substring(0, 50),
  description: this.data.description,
  price: parseFloat(this.data.price),
  location: this.data.location,
  category: this.data.categories[this.data.categoryIndex],
  images: imageUrls,
  contact_type: this.data.contactType,
  contact_value: this.data.contactValue,
}
```

- [ ] **Step 6: 在 publish.wxml 中插入联系方式区块**

打开 `mall-mini/miniprogram/pages/publish/publish.wxml`，在价格区块（`<!-- 价格 -->`）和分类区块（`<!-- 分类 -->`）之间插入：

```xml
<!-- 联系方式 -->
<view class="form-section card">
  <view class="section-title">联系方式</view>
  <view class="contact-type-row">
    <view
      class="contact-type-btn {{contactType === 'phone' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="0"
    >手机号</view>
    <view
      class="contact-type-btn {{contactType === 'wechat' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="1"
    >微信</view>
    <view
      class="contact-type-btn {{contactType === 'qq' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="2"
    >QQ</view>
  </view>
  <input
    class="contact-input"
    placeholder="{{contactType === 'phone' ? '请输入手机号' : contactType === 'wechat' ? '请输入微信号' : '请输入QQ号'}}"
    value="{{contactValue}}"
    bindinput="onContactValueInput"
  />
</view>
```

- [ ] **Step 7: 在 publish.wxss 末尾追加样式**

打开 `mall-mini/miniprogram/pages/publish/publish.wxss`，在文件末尾追加：

```css
/* 联系方式 */
.contact-type-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.contact-type-btn {
  flex: 1;
  text-align: center;
  padding: 14rpx 0;
  border: 2rpx solid #e0e0e0;
  border-radius: 8rpx;
  font-size: 28rpx;
  color: #666;
}

.contact-type-btn.active {
  border-color: #fa5151;
  color: #fa5151;
}

.contact-input {
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
}
```

- [ ] **Step 8: 手动验证**

在 WeChat DevTools 中打开发布页：
- 不填联系方式直接提交 → Toast "请填写联系方式"
- 点击"微信"按钮 → 边框变红，placeholder 变为"请输入微信号"
- 填写所有字段后提交 → 发布成功

- [ ] **Step 9: Commit**

```bash
git add mall-mini/miniprogram/pages/publish/publish.ts \
        mall-mini/miniprogram/pages/publish/publish.wxml \
        mall-mini/miniprogram/pages/publish/publish.wxss
git commit -m "feat: add contact type selector and input to publish page"
```

---

### Task 4: 前端编辑页 — productEdit.ts & productEdit.wxml & productEdit.wxss

**Files:**
- Modify: `mall-mini/miniprogram/pages/productEdit/productEdit.ts`
- Modify: `mall-mini/miniprogram/pages/productEdit/productEdit.wxml`
- Modify: `mall-mini/miniprogram/pages/productEdit/productEdit.wxss`

- [ ] **Step 1: 更新 EditData interface，新增 contact 相关字段**

打开 `mall-mini/miniprogram/pages/productEdit/productEdit.ts`，将 `interface EditData` 替换为：

```typescript
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
  contactType: 'phone' | 'wechat' | 'qq'
  contactValue: string
  contactTypes: string[]
}
```

- [ ] **Step 2: 更新 data 初始值**

在 `Page<EditData, ...>({ data: { ... } })` 中，在 `submitting: false` 之后追加：

```typescript
contactType: 'phone',
contactValue: '',
contactTypes: ['手机号', '微信', 'QQ'],
```

- [ ] **Step 3: 在 loadProduct() 的 setData 中回填联系方式**

找到 `loadProduct` 函数中的 `this.setData({ ... })` 调用，在已有字段之后追加：

```typescript
this.setData({
  images,
  description: p.description || '',
  price: p.price ? String(p.price) : '',
  location: p.location || '',
  regionNames,
  regionIndexes: [pi, ci, di],
  categoryIndex,
  contactType: (p.contact_type as 'phone' | 'wechat' | 'qq') || 'phone',
  contactValue: p.contact_value || '',
})
```

- [ ] **Step 4: 新增两个事件处理方法**

在 `onCategoryChange` 方法之后添加：

```typescript
onContactTypeSelect(e: WechatMiniprogram.TouchEvent) {
  const types: Array<'phone' | 'wechat' | 'qq'> = ['phone', 'wechat', 'qq']
  this.setData({ contactType: types[e.currentTarget.dataset.index] })
},

onContactValueInput(e: WechatMiniprogram.Input) {
  this.setData({ contactValue: e.detail.value })
},
```

- [ ] **Step 5: 在 validateForm() 末尾新增 contactValue 校验**

在 `return true` 之前插入：

```typescript
if (!this.data.contactValue.trim()) {
  wx.showToast({ title: '请填写联系方式', icon: 'none' })
  return false
}
```

- [ ] **Step 6: 在 submitForm() 的 put 调用中加入联系方式字段**

找到 `await put('/api/product/update', { ... })` 调用，替换为：

```typescript
await put('/api/product/update', {
  id: this.data.productId,
  description: this.data.description,
  price: parseFloat(this.data.price),
  location: this.data.location,
  images: imageUrls,
  contact_type: this.data.contactType,
  contact_value: this.data.contactValue,
})
```

- [ ] **Step 7: 在 productEdit.wxml 中插入联系方式区块**

打开 `mall-mini/miniprogram/pages/productEdit/productEdit.wxml`，在价格区块和分类区块之间插入（与 publish.wxml 完全相同的区块）：

```xml
<!-- 联系方式 -->
<view class="form-section card">
  <view class="section-title">联系方式</view>
  <view class="contact-type-row">
    <view
      class="contact-type-btn {{contactType === 'phone' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="0"
    >手机号</view>
    <view
      class="contact-type-btn {{contactType === 'wechat' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="1"
    >微信</view>
    <view
      class="contact-type-btn {{contactType === 'qq' ? 'active' : ''}}"
      bindtap="onContactTypeSelect"
      data-index="2"
    >QQ</view>
  </view>
  <input
    class="contact-input"
    placeholder="{{contactType === 'phone' ? '请输入手机号' : contactType === 'wechat' ? '请输入微信号' : '请输入QQ号'}}"
    value="{{contactValue}}"
    bindinput="onContactValueInput"
  />
</view>
```

- [ ] **Step 8: 在 productEdit.wxss 末尾追加样式**

打开 `mall-mini/miniprogram/pages/productEdit/productEdit.wxss`，在文件末尾追加（与 publish.wxss 相同的样式）：

```css
/* 联系方式 */
.contact-type-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.contact-type-btn {
  flex: 1;
  text-align: center;
  padding: 14rpx 0;
  border: 2rpx solid #e0e0e0;
  border-radius: 8rpx;
  font-size: 28rpx;
  color: #666;
}

.contact-type-btn.active {
  border-color: #fa5151;
  color: #fa5151;
}

.contact-input {
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
}
```

- [ ] **Step 9: 手动验证**

在 WeChat DevTools 中进入某个已发布商品的编辑页：
- 验证联系方式已回填（类型选中正确，号码正确）
- 修改联系方式后保存，重新进入编辑页验证更新成功

- [ ] **Step 10: Commit**

```bash
git add mall-mini/miniprogram/pages/productEdit/productEdit.ts \
        mall-mini/miniprogram/pages/productEdit/productEdit.wxml \
        mall-mini/miniprogram/pages/productEdit/productEdit.wxss
git commit -m "feat: add contact type selector and input to product edit page"
```

---

### Task 5: 前端详情页 — detail.ts & detail.wxml & detail.wxss

**Files:**
- Modify: `mall-mini/miniprogram/pages/detail/detail.ts`
- Modify: `mall-mini/miniprogram/pages/detail/detail.wxml`
- Modify: `mall-mini/miniprogram/pages/detail/detail.wxss`

- [ ] **Step 1: 在 Product interface 中新增两个字段**

打开 `mall-mini/miniprogram/pages/detail/detail.ts`，将 `interface Product` 替换为：

```typescript
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
  contactType: string
  contactValue: string
}
```

- [ ] **Step 2: 更新 data 中 product 的初始值**

在 `Page({ data: { product: { ... } } })` 中，在 `views: 0` 之后追加：

```typescript
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
  views: 0,
  contactType: '',
  contactValue: '',
},
```

- [ ] **Step 3: 在 loadProductDetail() 中映射新字段**

找到 `const product: Product = { ... }` 的构建部分，在 `views: ...` 之后追加两行：

```typescript
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
  views: Math.floor(Math.random() * 500) + 50,
  contactType: data.contact_type || '',
  contactValue: data.contact_value || '',
}
```

- [ ] **Step 4: 删除 contactSeller() 方法**

找到并删除以下方法（共 3 行）：

```typescript
contactSeller() {
  wx.showToast({ title: '功能开发中', icon: 'none' })
},
```

- [ ] **Step 5: 更新 detail.wxml — 在卖家区块后插入联系方式区块**

打开 `mall-mini/miniprogram/pages/detail/detail.wxml`。

首先，找到卖家区块末尾 `</view>（seller-section 的结束标签）`，在其后（`<!-- Time Info -->` 之前）插入：

```xml
<!-- 联系方式 -->
<view class="contact-section" wx:if="{{product.contactValue}}">
  <text class="section-title">联系卖家</text>
  <view class="contact-info-row">
    <text class="contact-type-tag">{{product.contactType === 'phone' ? '手机号' : product.contactType === 'wechat' ? '微信' : 'QQ'}}</text>
    <text class="contact-value-text">{{product.contactValue}}</text>
  </view>
</view>
```

然后，找到 `seller-card` 的 `bindtap="contactSeller"` 属性并删除（仅删除该属性，保留元素），将：

```xml
<view class="seller-card" bindtap="contactSeller">
```

改为：

```xml
<view class="seller-card">
```

最后，找到底部 Action Bar，将"联系卖家"按钮整行删除，只保留收藏按钮：

```xml
<!-- Bottom Action Bar -->
<view wx:if="{{!loading && !error}}" class="action-bar">
  <button class="action-btn primary" bindtap="toggleFavorite">
    {{isFavorite ? '已收藏' : '收藏'}}
  </button>
</view>
```

- [ ] **Step 6: 在 detail.wxss 末尾追加样式**

打开 `mall-mini/miniprogram/pages/detail/detail.wxss`，在文件末尾追加：

```css
/* 联系方式区块 */
.contact-section {
  margin: 24rpx 0;
  padding: 24rpx;
  background: #fff;
  border-radius: 16rpx;
}

.contact-info-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 12rpx;
}

.contact-type-tag {
  background: #fff3f0;
  color: #fa5151;
  font-size: 24rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}

.contact-value-text {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
}
```

- [ ] **Step 7: 手动验证**

在 WeChat DevTools 中打开一个已填联系方式的商品详情页：
- 验证联系方式区块出现在卖家信息下方
- 验证类型标签（手机号/微信/QQ）和号码正确显示
- 验证底部 Action Bar 只剩收藏按钮，点击收藏功能正常
- 打开一个老数据（contact_value 为空）的商品 → 联系方式区块不显示

- [ ] **Step 8: Commit**

```bash
git add mall-mini/miniprogram/pages/detail/detail.ts \
        mall-mini/miniprogram/pages/detail/detail.wxml \
        mall-mini/miniprogram/pages/detail/detail.wxss
git commit -m "feat: display contact info on product detail page, remove contactSeller placeholder"
```
