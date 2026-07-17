---
comet_change: mall-ios-favorites-profile
role: technical-design
canonical_spec: openspec
---

# iOS 收藏功能技术设计（mall-ios-favorites-profile）

## 背景与现状勘察

OpenSpec proposal 写于 2026-07-02，当时假设详情页收藏按钮是占位。2026-07-17 代码勘察确认现状已漂移：

1. **详情页收藏切换已实现**：`MallApp/Views/ProductDetailView.swift` 中 `ProductDetailViewModel.toggleFavorite()` 已调用 `FavoriteAPI.toggle`（定义于 `MallApp/Models/ProductDetail.swift`），含未登录拦截（`API.token` 为空时 toast「请先登录」，不发请求）。tasks 1.x 实际已完成。
2. **目录结构**：当前 iOS 代码为扁平结构 `MallApp/Views/*.swift` + `MallApp/Models/*.swift`（commit 693a2cc 重写），proposal 所述 `Features/...` 结构不存在。本设计遵循实际扁平结构。
3. **无 test target**：MallApp.xcodeproj 没有任何测试目标，原 tasks 4.x 单测项不可执行，已按用户确认改为构建+模拟器验证。
4. **后端接口**（mall-server，已存在）：
   - `POST /api/favorite/toggle`（需登录）body `{product_id}` → `{is_favorited}`
   - `GET /api/favorite/list`（需登录）query `page`/`page_size` → `{list, total, page, page_size}`，list 项为 `ProductSearchResult`，与 `/api/product/search` 完全同构

## 用户确认的决策（2026-07-17 brainstorming）

1. **范围**：只做缺失部分——`FavoriteAPI.list` 封装、收藏列表页、「我的」入口。详情页收藏按"已实现"处理，验证阶段一并覆盖。
2. **取消收藏交互**：卡片右上实心红心按钮，点击取消并本地移除（对齐 mall-mini）。
3. **测试策略**：不新建 test target；`xcodebuild build` 通过 + 模拟器连真后端全链路验证。

## 架构设计

### 1. API 封装 — 新增 `MallApp/Models/Favorite.swift`

- 将现有 `FavoriteAPI`（enum）从 `ProductDetail.swift` 迁入本文件集中管理，`toggle(productID:) -> Bool` 签名不变（`ProductDetailViewModel` 调用方无感知）。
- 新增 `FavoriteAPI.list(page:pageSize:) async throws -> ProductPage`：`GET /api/favorite/list`，直接复用现有 `ProductPage`/`Product` Decodable（服务端返回同构），不新增模型。

### 2. 收藏列表页 — 新增 `MallApp/Views/FavoriteListView.swift`

**FavoriteListViewModel**（`@MainActor @Observable`，模式与 `HomeViewModel` 一致）：

- 状态：`products: [Product]`、`total: Int`、`loading`、`initialLoaded`、`errorMessage`、`toast`
- `reload()`：重置 page=1 加载并替换列表（下拉刷新 / onAppear 调用）
- `loadMoreIfNeeded(current:)`：滚动到最后一项且 `hasMore` 时加载下一页追加；`hasMore = products.count < total`（对齐 mall-mini，比 HomeView 的 `count >= pageSize` 判断更准确）；`loading` 防重入
- `unfavorite(_ product: Product)`：调 `FavoriteAPI.toggle(productID:)`，成功后 `products.removeAll { $0.id == product.id }` 且 `total -= 1`；失败 toast 错误信息、保留原项

**FavoriteListView**：

- `ScrollView + LazyVStack`（与 HomeView 同模式），`.refreshable` 下拉刷新
- 卡片复用 `ProductCardView`，`ZStack`/`overlay` 右上角叠加实心红心按钮（`heart.fill`，红色），点击调 `unfavorite`；点击卡片本体 `NavigationLink(value: product.id)` 推入 `ProductDetailView`（复用 ProfileView 所在 NavigationStack 的 `navigationDestination`）
- 空状态：`initialLoaded && products.isEmpty` 时展示「暂无收藏商品」占位（同 HomeView 的 statusPlaceholder 样式）
- **数据一致性**：`onAppear` 时 `reload()` 第 1 页（对齐 mall-mini `onShow` 全量重载），保证从详情页取消收藏返回后列表状态一致

### 3. 「我的」入口 — 修改 `MallApp/Views/ProfileView.swift`

- 登录态 `loggedInContent` 在用户卡片与退出按钮之间新增菜单区：「我的收藏」行（heart 图标 + 标题 + chevron），`NavigationLink` → `FavoriteListView`
- `ProfileView` 已有 `NavigationStack`，需补 `navigationDestination(for: Int.self)` 以支持收藏列表 → 商品详情的跳转
- 未登录态整页为 `AuthFormView`，收藏入口不可见（Spec Patch 已同步该行为）

## 错误处理

- 列表加载失败：`errorMessage` + 重试按钮（同 HomeView 模式）
- 取消收藏失败：toast 展示 `error.localizedDescription`，列表项保留
- 401/未登录：接口层返回业务错误提示；入口仅登录态可见，正常路径不会未登录访问

## Spec Patch（已回写 delta spec）

- `specs/ios-product-favorite/spec.md`：「我的」页面收藏入口 requirement 的未登录场景，由「点击提示登录」改为「未登录时整页为登录表单，入口不可见」——iOS 现有交互下原场景不可达，功能意图（未登录无法进入收藏列表）不变。
- `tasks.md`：详情页任务标注已实现；去掉单测任务；路径改为实际扁平结构。

## 风险与取舍

- **取消收藏本地移除 vs 重新拉取**：沿用 mall-mini 本地移除策略；多设备并发取消的短暂不一致可接受，下拉刷新兜底。
- **onAppear 全量重载**重置滚动位置：与 mall-mini 行为一致，收藏列表数据量小，可接受。
- **无单元测试**：与仓库现状一致（无 test target）；验证依赖构建 + 模拟器真后端全链路。

## 验证策略

1. `xcodebuild build`（iOS Simulator destination）通过
2. 模拟器连真后端（https://yangzhongyu.site）全链路：登录 → 详情页收藏 → 「我的收藏」列表可见该商品 → 列表红心取消 → 项移除且 total 减一 → 重进详情页确认未收藏 → 全部取消后空状态 → 未登录时「我的」页为登录表单（无收藏入口）
3. 可用现有 DEBUG 环境变量直达：`DEBUG_SELECTED_TAB=3`（我的 Tab）、`DEBUG_AUTO_AUTH`（自动登录）、`DEBUG_OPEN_PRODUCT_ID`（直达详情页）
