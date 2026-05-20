# 商品搜索分类与地区筛选设计

**Date:** 2026-05-20
**Status:** Approved

## Problem

搜索功能目前只支持关键词和状态过滤，无法按分类或地区缩小范围。
此外，分类字段虽然在发布/编辑页有 UI，但后端从未持久化；地区仅存为合并字符串，无法精确过滤。

## Goal

- 首页和搜索页均支持按**商品分类**和**发布地区**（精确到省/市/县区）筛选
- 修复 `category` 字段从未写入数据库的问题
- 发布/编辑时额外存储 `province`、`city`、`district` 三个独立字段

## Decisions

- 地区存储采用**独立三字段**（province / city / district），精确匹配，不依赖字符串 LIKE
- 历史商品四个新字段均为空字符串，不参与筛选结果（符合预期）
- 首页筛选栏：搜索框下方横排两个芯片（分类 + 地区），点击分别弹出选择面板
- 搜索页：同样的筛选芯片栏，与首页共用相同的交互模式
- 地区选择器复用发布页已有的 `china-regions.ts` 三级联动数据，用户可停在省/市级提交

## Architecture

### Files Changed

| 文件 | 变更 |
|------|------|
| `mall-server/internal/app/dao/product.entity.go` | 新增 `Category`、`Province`、`City`、`District` 四字段 |
| `mall-server/internal/app/dao/product.repo.go` | `ProductSearchResult` 加 `Category`；`SearchProducts` 新增四个可选过滤参数；SELECT 追加四列 |
| `mall-server/internal/app/service/types.go` | `SearchProductRequest` 加四参数；`PublishProductRequest` 加 `Province/City/District`；`UpdateProductRequest` 加 `Category/Province/City/District` |
| `mall-server/internal/app/service/product.go` | `PublishProduct` 写入四字段；`UpdateProduct` 更新四字段；`SearchProducts` handler 传四参数给 dao |
| `mall-mini/miniprogram/pages/home/home.ts` | 新增筛选状态、事件处理器、`loadProducts` 携带筛选参数 |
| `mall-mini/miniprogram/pages/home/home.wxml` | 新增横向筛选芯片栏 + 分类面板 + 地区选择器 |
| `mall-mini/miniprogram/pages/home/home.wxss` | 新增筛选栏样式 |
| `mall-mini/miniprogram/pages/search/search.ts` | 同首页：新增筛选状态、事件处理器、search() 携带筛选参数 |
| `mall-mini/miniprogram/pages/search/search.wxml` | 新增横向筛选芯片栏 |
| `mall-mini/miniprogram/pages/search/search.wxss` | 新增筛选栏样式 |
| `mall-mini/miniprogram/pages/publish/publish.ts` | `productData` 补充 `province`/`city`/`district` |
| `mall-mini/miniprogram/pages/productEdit/productEdit.ts` | `put` payload 补充四字段；`loadProduct` 回填 `category`/`province`/`city`/`district` |

### 数据模型

```go
// product.entity.go 新增字段（插在 ContactType 之前）
Category string `gorm:"column:category;type:varchar(50);not null;default:''"  json:"category"  comment:"商品分类"`
Province string `gorm:"column:province;type:varchar(50);not null;default:''"  json:"province"  comment:"省"`
City     string `gorm:"column:city;type:varchar(50);not null;default:''"      json:"city"      comment:"市"`
District string `gorm:"column:district;type:varchar(50);not null;default:''" json:"district"  comment:"县区"`
```

合法 `category` 值：`电子产品` / `服装鞋帽` / `图书文具` / `生活用品` / `数码配件` / `其他`（与前端发布页保持一致）。

GORM AutoMigrate 自动建列，无需手写迁移。

### 后端 DTO

**`SearchProductRequest` 新增（可选）：**
```go
Category string `form:"category"`
Province string `form:"province"`
City     string `form:"city"`
District string `form:"district"`
```

**`PublishProductRequest` 新增：**
```go
Province string `json:"province"`
City     string `json:"city"`
District string `json:"district"`
```
（`Category` 已存在，不再重复添加；handler 补写入 DB 即可）

**`UpdateProductRequest` 新增：**
```go
Category string `json:"category"`
Province string `json:"province"`
City     string `json:"city"`
District string `json:"district"`
```

**`ProductSearchResult` 新增：**
```go
Category string `json:"category"`
```

**`ProductDetail` 新增（供编辑页回填）：**
```go
Category string `json:"category"`
Province string `json:"province"`
City     string `json:"city"`
District string `json:"district"`
```

`GetProductByID` 的 SELECT 追加 `product.category, product.province, product.city, product.district`。

### 后端搜索查询

`dao.SearchProducts` 签名变更：
```go
func SearchProducts(db *gorm.DB, keyword, sort string, status *int,
    category, province, city, district string,
    page, pageSize int) ([]ProductSearchResult, int64, error)
```

WHERE 条件按传入参数有条件追加：
```go
if category != "" { query = query.Where("product.category = ?", category) }
if province != "" { query = query.Where("product.province = ?", province) }
if city     != "" { query = query.Where("product.city = ?", city) }
if district != "" { query = query.Where("product.district = ?", district) }
```

SELECT 追加：`product.category`（`province`/`city`/`district` 不需要出现在列表结果里，只用于过滤）。

### 前端状态（home.ts & search.ts）

两页均新增以下 data 字段：

```typescript
// 筛选状态
selectedCategory: '',          // '' = 全部
selectedProvince: '',
selectedCity: '',
selectedDistrict: '',
// 面板显示控制
showCategoryPanel: false,
showRegionPanel: false,
// 地区选择器中间状态（三级联动）
regionStep: 0,                 // 0=选省 1=选市 2=选县区
regionProvince: '',
regionCity: '',
regionCities: [] as string[],
regionDistricts: [] as string[],
// 分类列表
categories: ['全部', '电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
```

事件处理器：

| 方法 | 说明 |
|------|------|
| `onOpenCategoryPanel()` | 打开分类面板 |
| `onSelectCategory(e)` | 选择分类，关闭面板，重置列表，触发加载 |
| `onOpenRegionPanel()` | 打开地区面板，重置到省级 |
| `onSelectProvince(e)` | 选省，进入市级列表 |
| `onSelectCity(e)` | 选市，进入县区列表 |
| `onSelectDistrict(e)` | 选县区，关闭面板，触发加载 |
| `onConfirmRegion()` | 用户在省/市级直接确认，关闭面板，触发加载 |
| `onClearCategory()` | 清除分类筛选，重置列表 |
| `onClearRegion()` | 清除地区筛选，重置列表 |
| `onCloseCategoryPanel()` | 点击遮罩关闭分类面板（不改变已选值） |
| `onCloseRegionPanel()` | 点击遮罩关闭地区面板（不改变已选值） |

`loadProducts()` / `search()` 在请求参数中追加：
```typescript
if (this.data.selectedCategory) params.category = this.data.selectedCategory
if (this.data.selectedProvince) params.province = this.data.selectedProvince
if (this.data.selectedCity)     params.city = this.data.selectedCity
if (this.data.selectedDistrict) params.district = this.data.selectedDistrict
```

### 前端筛选芯片 UI（home.wxml / search.wxml）

搜索框下方插入：

```xml
<!-- 筛选栏 -->
<view class="filter-bar">
  <view class="filter-chip {{selectedCategory ? 'active' : ''}}"
        bindtap="onOpenCategoryPanel">
    <text>{{selectedCategory || '全部分类'}}</text>
    <text wx:if="{{selectedCategory}}" bindtap="onClearCategory" catchtap="onClearCategory"> ✕</text>
    <text wx:else> ▾</text>
  </view>
  <view class="filter-chip {{selectedProvince ? 'active' : ''}}"
        bindtap="onOpenRegionPanel">
    <text>{{selectedDistrict || selectedCity || selectedProvince || '全部地区'}}</text>
    <text wx:if="{{selectedProvince}}" bindtap="onClearRegion" catchtap="onClearRegion"> ✕</text>
    <text wx:else> ▾</text>
  </view>
</view>

<!-- 分类面板 -->
<view class="panel-mask" wx:if="{{showCategoryPanel}}" bindtap="onCloseCategoryPanel" />
<view class="category-panel" wx:if="{{showCategoryPanel}}">
  <view wx:for="{{categories}}" wx:key="index"
        class="category-item {{item === (selectedCategory || '全部') ? 'selected' : ''}}"
        bindtap="onSelectCategory" data-value="{{item === '全部' ? '' : item}}">
    {{item}}
  </view>
</view>

<!-- 地区面板 -->
<view class="panel-mask" wx:if="{{showRegionPanel}}" bindtap="onCloseRegionPanel" />
<view class="region-panel" wx:if="{{showRegionPanel}}">
  <!-- 面包屑 + 确认按钮 -->
  <view class="region-header">
    <text>{{regionProvince || '请选择省份'}}</text>
    <text wx:if="{{regionCity}}"> / {{regionCity}}</text>
    <button wx:if="{{regionStep > 0}}" bindtap="onConfirmRegion">确认</button>
  </view>
  <!-- 省列表（step=0） -->
  <scroll-view wx:if="{{regionStep === 0}}" scroll-y class="region-list">
    <view wx:for="{{provinces}}" wx:key="index"
          class="region-item" bindtap="onSelectProvince" data-province="{{item}}">
      {{item}}
    </view>
  </scroll-view>
  <!-- 市列表（step=1） -->
  <scroll-view wx:if="{{regionStep === 1}}" scroll-y class="region-list">
    <view wx:for="{{regionCities}}" wx:key="index"
          class="region-item" bindtap="onSelectCity" data-city="{{item}}">
      {{item}}
    </view>
  </scroll-view>
  <!-- 县区列表（step=2） -->
  <scroll-view wx:if="{{regionStep === 2}}" scroll-y class="region-list">
    <view wx:for="{{regionDistricts}}" wx:key="index"
          class="region-item" bindtap="onSelectDistrict" data-district="{{item}}">
      {{item}}
    </view>
  </scroll-view>
</view>
```

`provinces` 从 `china-regions.ts` 中取省名列表（`Object.keys(regions)`）。

### 发布页 & 编辑页微调

**`publish.ts`** — `productData` 补充三字段（`province`/`city`/`district`，从现有 `regionNames` state 取值）：
```typescript
province: this.data.regionNames[0]?.[this.data.regionIndexes[0]] || '',
city: this.data.regionNames[1]?.[this.data.regionIndexes[1]] || '',
district: this.data.regionNames[2]?.[this.data.regionIndexes[2]] || '',
```

**`productEdit.ts`** — `put` payload 补充四字段；`loadProduct` 从详情接口回填（详情接口需同步在 `ProductDetail` 中加 `Category`/`Province`/`City`/`District`）。

## Error Handling

| 场景 | 行为 |
|------|------|
| 只传 province 不传 city/district | 筛选该省全部商品 |
| 传 province + city | 筛选该市全部商品 |
| 传三级 | 精确筛选该县区 |
| 历史商品（字段为空） | 被地区/分类筛选自动排除，符合预期 |
| 分类 + 地区 + 关键词同时传 | 后端 AND 组合，全部生效 |
| 清除筛选 | 前端置空对应 state，重新加载列表 |

## Testing

1. 首页：分类芯片选"电子产品" → 仅展示电子产品分类商品
2. 首页：地区芯片选到市级点确认 → 仅展示该市商品
3. 首页：同时选分类+地区 → 两条件 AND 过滤
4. 首页：点 ✕ 清除单个筛选 → 列表还原
5. 搜索页：同上四项
6. 发布商品后进详情页 → `category`/`province`/`city`/`district` 显示正确
7. 编辑页进入 → `category` 已回填，`province`/`city`/`district` 正确选中
8. 搜索页 keyword + 分类筛选同时使用 → 结果同时满足两条件
