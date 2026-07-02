## Why

`mall-ios-browse-search-detail`（change #2）在商品详情页留了收藏按钮的占位（点击仅提示"即将上线"），`mall-ios-publish-edit-my-listings`（change #3）补上了「我发布的」入口。本 change 是 5 个拆分 change 中的第 4 个，把「收藏」这一常见的二手交易平台留资/复购行为落地：用户可以在浏览时收藏心仪商品，并在「我的」页面统一查看和管理。

## What Changes

- 商品详情页（`ProductDetailView`）的收藏按钮从占位改为真实交互：点击调用 `POST /api/favorite/toggle`，成功后更新按钮态（已收藏/未收藏）。
- 新增 `Features/Favorite` 模块：对接 `GET /api/favorite/list` 展示收藏商品列表（标题/价格/地区/卖家/封面图），支持分页、下拉刷新、取消收藏（列表内直接调用 `POST /api/favorite/toggle` 并从列表移除）、点击进入详情。
- 「我的」Tab 菜单加入「我的收藏」入口，跳转到 `Features/Favorite` 列表；未登录时点击提示登录（对齐 mall-mini 行为）。

## Capabilities

### New Capabilities
- `ios-product-favorite`: 商品收藏/取消收藏（详情页切换 + 收藏列表管理）。

### Modified Capabilities
- `ios-product-detail`: 收藏按钮从"点击展示即将上线"改为真实收藏切换行为——这是 `mall-ios-browse-search-detail` 中 `ios-product-detail` capability 的既有需求变更（该占位需求被替换为真实交互）。

## Impact

- **受影响代码**：`mall-ios/Features/ProductDetail/*`（收藏按钮改为真实交互）、新增 `mall-ios/Features/Favorite/*`、修改 `mall-ios/Features/Profile/View/ProfileView.swift`（加入「我的收藏」菜单项）。
- **对接的后端 API**：`POST /api/favorite/toggle`（需登录）、`GET /api/favorite/list`（需登录）。
- **依赖关系**：依赖 `mall-ios-network-auth-foundation`（登录态）、`mall-ios-browse-search-detail`（详情页收藏按钮挂载点、`Product` 相关展示组件复用）。
- **不影响** `mall-server`、`mall-mini`、`mall-admin-web`。
