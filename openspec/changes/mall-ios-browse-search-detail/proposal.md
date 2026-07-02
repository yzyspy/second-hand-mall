## Why

`mall-ios-network-auth-foundation`（change #1）已经打好网络层与账号会话的地基，但首页、发布、消息三个 Tab 仍是占位文字。mall-mini 的核心用户路径是「首页浏览 → 按分类/地区筛选 → 搜索 → 查看商品详情」，这是二手交易平台最基础也是使用频率最高的功能。本 change（5 个拆分 change 中的第 2 个）把这条浏览主链路对齐到 mall-ios，为后续发布、收藏、聊天 change 提供可浏览、可跳转的商品数据基础。

## What Changes

- 重写 `Features/Home`：对接 `GET /api/product/search`（`status=0` 仅在售）展示商品列表卡片（标题/价格/地区/卖家/图片），支持下拉刷新与上拉加载分页（`page`/`page_size`）。
- 首页新增分类筛选（`电子产品/服装鞋帽/图书文具/生活用品/数码配件/其他`）与地区筛选（省/市/区三级联动，复用 mall-mini 的省市区数据结构）面板，选择后重新拉取列表。
- 新增 `Features/Search` 模块：关键字搜索、排序（最新发布/最早发布）、状态筛选（全部/在售/已售出）、分类与地区筛选，分页加载，结果点击进入详情。
- 新增 `Features/ProductDetail` 模块：对接 `GET /api/product/detail`，展示商品图片轮播、标题、价格、描述、成色/分类、发布时间、卖家信息；收藏按钮与"联系卖家"按钮在本 change 中为**占位跳转**（真实收藏切换在 `mall-ios-favorites-profile`、真实私信在 `mall-ios-chat-messaging` 中实现）。
- 首页顶部新增搜索入口，点击进入搜索页。

## Capabilities

### New Capabilities
- `ios-product-browse`: 首页商品列表浏览，含分类/地区筛选、下拉刷新、分页加载。
- `ios-product-search`: 关键字搜索商品，含排序、状态与分类/地区筛选、分页加载。
- `ios-product-detail`: 商品详情展示（不含收藏/私信的真实业务逻辑，仅占位入口）。

### Modified Capabilities
（无——不修改 change #1 已定义的 `ios-network-client` / `ios-account-session` 需求，仅复用）

## Impact

- **受影响代码**：`mall-ios/Features/Home/*`（从占位改为真实实现）、新增 `mall-ios/Features/Search/*`、新增 `mall-ios/Features/ProductDetail/*`、可能新增共享的 `Core/Models/Product.swift` 与 `Core/Data/ChinaRegions.swift`（复用 mall-mini 的省市区数据，需转换为 Swift 静态数据）。
- **对接的后端 API**（均为公开接口，无需鉴权）：`GET /api/product/search`（列表+搜索共用）、`GET /api/product/detail`。
- **依赖关系**：依赖 `mall-ios-network-auth-foundation`（`APIClient.request`）。本 change 的 `ProductDetail` 页面为 `mall-ios-favorites-profile`（收藏按钮）与 `mall-ios-chat-messaging`（联系卖家按钮）预留挂载点，但不实现其行为。
- **不影响** `mall-server`、`mall-mini`、`mall-admin-web`。
