# 重新编辑已发布商品 — 设计文档

## 概述

在"我的"页面新增"我的发布"入口，用户可浏览自己发布的所有商品，并对在售商品执行编辑、标记已售、下架操作；对任意状态商品执行删除。

---

## 后端设计

### 新增 API 接口（均需 JWT 认证）

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/product/mine` | 获取当前登录用户发布的商品列表，支持 `page`/`page_size` 分页 |
| PUT | `/api/product/update` | 更新商品内容，仅限在售（status=0）且属于本人的商品 |
| POST | `/api/product/change-status` | 变更商品状态（status=1 已售，status=2 下架），须本人校验 |

### 请求/响应结构

**GET /api/product/mine**
```
Query: page int, page_size int
Response: { list: [ { id, title, price, images, status, created_at } ], total, page, page_size }
```

**PUT /api/product/update**
```
Body: { id, description, price, location, category, images[] }
title 取 description 前 50 字（与 publish 保持一致）
Response: { code: 0, msg: "success" }
```

**POST /api/product/change-status**
```
Body: { id, status }  // status: 1=已售, 2=下架
Response: { code: 0, msg: "success" }
```

### 权限校验规则

- `update`：product.UserId == 当前用户 ID，且 product.Status == 0
- `change-status`：product.UserId == 当前用户 ID

### 涉及文件

- `mall-server/internal/app/dao/product.repo.go` — 新增 `GetMyProducts`、`UpdateProduct`、`ChangeProductStatus`
- `mall-server/internal/app/service/product.go` — 新增三个 handler
- `mall-server/internal/app/service/types.go` — 新增 Request/Response struct
- `mall-server/internal/app/router/router.go` — 在认证路由组注册三条路由

---

## 前端设计

### 新增页面

**1. `pages/myPublish/myPublish`（我的发布列表）**

布局：
- 商品卡片：左侧第一张缩略图 + 右侧标题、价格、状态标签
  - 在售：绿色标签
  - 已下架：灰色标签
  - 已售出：红色标签
- 操作栏（卡片底部）：
  - 在售 → 「编辑」「标记已售」「下架」
  - 已售/已下架 → 「删除」（调 change-status status=2，或直接软删除）
- 交互：下拉刷新；滚动到底触发加载更多（分页）
- 空状态：「还没有发布的商品」提示

**2. `pages/productEdit/productEdit`（编辑商品）**

字段：与 publish 页面完全相同（图片、描述、价格、分类、省市区级联）

数据回填：
- `onLoad(options)` 取 `options.id`，调 `GET /api/product/detail?id=xxx` 拉取数据
- 图片回填：远程 URL 图片直接显示，不重新上传；用户新增的本地图片才走 Qiniu 上传
- 省市区回填：解析 location 字符串，匹配 china-regions 数据设置 `regionIndexes`

提交：
- 调 `PUT /api/product/update`
- 成功后 `wx.navigateBack()` 返回列表，列表页 `onShow` 触发刷新

### 导航关系

```
pages/my/my  →（已有链接）→  pages/myPublish/myPublish
pages/myPublish/myPublish  →  pages/productEdit/productEdit?id=xxx
pages/productEdit/productEdit  →（navigateBack）→  pages/myPublish/myPublish
```

### app.json 变更

在 `pages` 数组中新增：
```json
"pages/myPublish/myPublish",
"pages/productEdit/productEdit"
```

---

## 状态机

```
在售(0) --[标记已售]--> 已售(1)
在售(0) --[下架]------> 已下架(2)
已售(1) --[删除]------> 已下架(2)   （前端用"删除"按钮，实际调 change-status status=2）
已下架(2) --[删除]----> 已下架(2)   （幂等，无变化）
```

---

## 约束与边界

- 已售/已下架商品不可编辑，编辑按钮不显示
- `update` 接口后端二次校验 status==0，防止绕过前端直接调用
- 图片最多 9 张，与 publish 保持一致
- 删除为软删除（status=2），不物理删除数据
