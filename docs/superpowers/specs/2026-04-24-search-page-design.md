# 搜索结果页面功能设计

## 概述

为二手商城小程序首页的搜索功能增加搜索结果页面，支持关键词搜索、按发布时间排序、按商品状态筛选。

## 后端设计

### 商品实体 (Product)

路径：`mall-server/internal/app/dao/product.entity.go`

```
Product
├── ID          uint           GORM 主键
├── Title       string         商品标题 (varchar 200)
├── Description string         商品描述 (text)
├── Price       float64        价格
├── Images      string         图片URL列表，逗号分隔 (varchar 1000)
├── Location    string         交易地点 (varchar 100)
├── Status      int            状态：0=在售，1=已售出，2=已下架
├── UserId      uint           发布者ID，外键关联 SysUser
├── CreatedAt   time.Time      发布时间 (GORM 自动)
├── UpdatedAt   time.Time      更新时间 (GORM 自动)
```

### 搜索 API

```
GET /api/product/search
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 标题模糊匹配，为空则查全部 |
| sort | string | 否 | 排序方式：time_desc(默认)、time_asc |
| status | int | 否 | 状态筛选：不传查全部，0=在售，1=已售出 |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 10，最大 50 |

**返回结构：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "商品标题",
        "price": 99.00,
        "images": "url1,url2",
        "location": "广州天河区",
        "status": 0,
        "seller": "卖家昵称",
        "create_time": "2026-04-20T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### Repository 函数

路径：`mall-server/internal/app/dao/product.repo.go`

```go
func SearchProducts(db *gorm.DB, keyword string, sort string, status *int, page, pageSize int) ([]Product, int64, error)
```

实现要点：
- keyword 使用 `LIKE %keyword%` 模糊匹配 title 字段
- sort 控制 `ORDER BY created_at DESC/ASC`
- status 为 nil 时不筛选，否则 `WHERE status = ?`
- 使用 `COUNT(*) OVER()` 或单独查询获取 total

## 前端设计

### 页面结构

路径：`miniprogram/pages/search/`

**文件列表：**

- `search.ts` — 页面逻辑
- `search.wxml` — 页面结构
- `search.wxss` — 页面样式
- `search.json` — 页面配置

### UI 布局

```
┌─────────────────────────────────┐
│  [输入框] 搜索商品      [搜索按钮] │  ← 搜索栏
├─────────────────────────────────┤
│  排序: [最新发布 ▼] [状态 ▼]      │  ← 排序筛选栏
├─────────────────────────────────┤
│  商品卡片列表                    │  ← 复用首页卡片样式
│  ...                            │
├─────────────────────────────────┤
│  加载中... / 没有更多 / 暂无商品  │  ← 状态提示
└─────────────────────────────────┘
```

### 页面数据 (search.ts)

```typescript
interface SearchData {
  keyword: string          // 搜索关键词
  sort: string             // 排序方式：time_desc | time_asc
  status: number | null    // 状态筛选：null=全部
  products: ProductItem[]  // 商品列表
  loading: boolean        // 加载状态
  page: number            // 当前页码
  hasMore: boolean        // 是否有更多
  sortOptions: string[]   // 排序选项
  statusOptions: string[] // 状态选项
  sortIndex: number       // 当前排序索引
  statusIndex: number     // 当前状态索引
}
```

### 交互流程

1. **页面加载**：从 URL 参数获取初始 keyword，如有则自动搜索
2. **搜索按钮**：点击后调用 API，重置 page=1
3. **排序变更**：picker 变更后重新搜索，重置 page=1
4. **触底加载**：滚动到底部加载下一页
5. **点击商品**：跳转详情页 `/pages/detail/detail?id=xxx`

### API 调用

在 `utils/request.ts` 基础上调用：

```typescript
get('/api/product/search', {
  keyword: this.data.keyword,
  sort: this.data.sort,
  status: this.data.status,
  page: this.data.page,
  page_size: 10
})
```

### 样式

复用首页 `home.wxss` 的商品卡片样式，搜索栏和排序筛选栏单独编写。

## 文件清单

### 后端 新增

- `internal/app/dao/product.entity.go` — 商品实体
- `internal/app/dao/product.repo.go` — 商品仓库函数
- `internal/app/service/product.go` — 商品处理器

### 后端 修改

- `internal/app/router/router.go` — 注册搜索路由
- `internal/app/models/init.go` — 自动迁移 Product 表（如需要）

### 前端 新增

- `miniprogram/pages/search/search.ts`
- `miniprogram/pages/search/search.wxml`
- `miniprogram/pages/search/search.wxss`
- `miniprogram/pages/search/search.json`

### 前端 修改

- `miniprogram/app.json` — 注册 search 页面路由

## 技术要点

1. **后端**：使用 GORM 的 `Where("title LIKE ?", "%"+keyword+"%")` 实现模糊搜索
2. **后端**：status 参数使用指针类型 `*int` 区分"不筛选"和"筛选状态0"
3. **前端**：使用 `wx.createSelectorQuery` 或 onReachBottom 实现触底加载
4. **样式**：使用 rpx 单位适配不同屏幕
