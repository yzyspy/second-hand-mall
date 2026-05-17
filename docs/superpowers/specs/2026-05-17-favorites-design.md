# 我的收藏功能设计文档

**日期:** 2026-05-17  
**状态:** 已批准，待实现

---

## 需求概述

为二手交易小程序新增"我的收藏"功能，允许用户在商品详情页收藏感兴趣的在售商品，并在个人中心的「我的收藏」页面统一管理。

---

## 需求确认

| 决策点 | 结论 |
|--------|------|
| 收藏入口 | 仅商品详情页（心形按钮） |
| 已售出/下架商品 | 自动从收藏列表消失，不展示 |
| 收藏人数展示 | 不展示 |
| 收藏列表布局 | 竖向列表（与「我的发布」页一致） |

---

## 数据库设计

### 新增表：`user_favorite`

```sql
CREATE TABLE user_favorite (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, product_id)
);
```

**设计要点：**
- `UNIQUE(user_id, product_id)` 防止重复收藏，toggle 逻辑依赖此约束判断插入或删除
- 不加 `favorite_count` 到 `product` 表（不展示收藏数）
- 不使用软删除（`deleted_at`），取消收藏即物理删除
- 查询收藏列表时 JOIN `product` 并过滤 `status = 0`，已售出/下架商品自动消失

### GORM 实体

文件：`mall-server/internal/app/dao/favorite.entity.go`

```go
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

---

## 后端 API 设计

### 新增接口

#### 1. 收藏/取消收藏（Toggle）

```
POST /api/favorite/toggle
Auth: 必须登录（AuthMiddleware）
```

**请求 Body:**
```json
{ "product_id": 123 }
```

**响应:**
```json
{ "code": 0, "msg": "success", "data": { "is_favorited": true } }
```

**逻辑：**
- 记录不存在 → INSERT → 返回 `is_favorited: true`
- 记录已存在 → DELETE → 返回 `is_favorited: false`

#### 2. 我的收藏列表

```
GET /api/favorite/list?page=1&page_size=10
Auth: 必须登录（AuthMiddleware）
```

**响应:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "iPhone 14 Pro",
        "price": 5999,
        "images": "url1,url2",
        "location": "北京市朝阳区",
        "seller": "张三",
        "avatar": "avatar_url"
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 10
  }
}
```

**逻辑：**
- JOIN `product` 表，WHERE `product.status = 0`（只返回在售商品）
- ORDER BY `user_favorite.created_at DESC`（最新收藏在最前）

### 修改现有接口

#### 3. 商品详情 `GET /api/product/detail` — 新增 `is_favorited` 字段

**问题：** 详情页是公开路由，但收藏状态需要知道当前用户。

**方案：** 新增 `OptionalAuthMiddleware`，有 token 时解析注入 `user_id`，无 token 时静默跳过（不返回 401）。

**路由变更：**
```go
// 原来
r.GET("/api/product/detail", service.GetProductDetail(svc))

// 改为
r.GET("/api/product/detail", OptionalAuthMiddleware(), service.GetProductDetail(svc))
```

**响应新增字段：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "title": "...",
    "is_favorited": true
  }
}
```

- 未登录用户 → `is_favorited: false`
- 已登录用户 → 查 `user_favorite` 表返回真实状态

### 接口汇总

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/favorite/toggle` | AuthMiddleware | 收藏/取消收藏 |
| GET | `/api/favorite/list` | AuthMiddleware | 我的收藏列表（分页） |
| GET | `/api/product/detail` | OptionalAuthMiddleware | 新增 `is_favorited` 字段 |

---

## 前端设计

### 新页面：`pages/myFavorite/myFavorite`

**布局：** 竖向列表，与 `myPublish` 页结构一致。

**每行结构：**
- 左：64×64 商品缩略图（取 images 第一张）
- 中：商品标题、价格（红色）、地点、卖家昵称
- 右：实心红色心形图标，点击直接取消收藏（无需进入详情页）

**页面行为：**
- `onShow` 触发刷新（重置到第 1 页），保证从详情页收藏后返回列表即时更新
- 下拉刷新重置到第 1 页
- 滚动到底自动加载下一页（无限滚动）
- 空状态：展示"还没有收藏的商品"提示

**导航：** 点击商品行（非心形图标区域）跳转到 `pages/detail/detail?id=xxx`

### 改造现有页面：`pages/detail/detail`

**改造点：**

1. `onLoad` 时：由详情接口返回的 `is_favorited` 初始化心形图标状态（`true` = 实心红，`false` = 空心灰）
2. 点击心形时：
   - 未登录 → 提示"请先登录"，不调用接口
   - 已登录 → 调用 `POST /api/favorite/toggle`，用返回的 `is_favorited` 更新图标状态
3. 移除当前详情页"收藏数"的 mock 数据展示

### `app.json` 页面注册

在 `pages` 数组中新增：
```json
"pages/myFavorite/myFavorite"
```

### `pages/my/my.ts` 导航修改

将「我的收藏」菜单项的跳转路径从占位改为：
```
/pages/myFavorite/myFavorite
```

---

## 文件变更清单

### 后端（mall-server）

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/app/dao/favorite.entity.go` | 新增 | UserFavorite GORM 实体 |
| `internal/app/dao/favorite.repo.go` | 新增 | ToggleFavorite、GetFavoriteList、IsFavorited 函数 |
| `internal/app/service/favorite.go` | 新增 | ToggleFavorite、GetFavoriteList handler |
| `internal/app/service/types.go` | 修改 | 新增 FavoriteToggleRequest、FavoriteListRequest；ProductDetail 新增 IsFavorited 字段 |
| `internal/app/router/auth.go` | 修改 | 新增 OptionalAuthMiddleware |
| `internal/app/router/router.go` | 修改 | 注册新路由，detail 路由改用 OptionalAuthMiddleware |
| `internal/app/models/init.go` | 修改 | AutoMigrate 新增 UserFavorite |

### 前端（mall-mini）

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `miniprogram/pages/myFavorite/myFavorite.ts` | 新增 | 收藏列表页逻辑 |
| `miniprogram/pages/myFavorite/myFavorite.wxml` | 新增 | 收藏列表页模板 |
| `miniprogram/pages/myFavorite/myFavorite.wxss` | 新增 | 收藏列表页样式 |
| `miniprogram/pages/myFavorite/myFavorite.json` | 新增 | 页面配置 |
| `miniprogram/pages/detail/detail.ts` | 修改 | 接入 is_favorited、toggle 逻辑 |
| `miniprogram/pages/detail/detail.wxml` | 修改 | 移除收藏数 mock，心形图标绑定状态 |
| `miniprogram/pages/my/my.ts` | 修改 | 我的收藏跳转路径 |
| `miniprogram/app.json` | 修改 | 注册 myFavorite 页面 |
