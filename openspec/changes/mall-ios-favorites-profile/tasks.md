## 1. 详情页收藏真实交互（勘察确认已实现，验证阶段覆盖）

- [x] 1.1 `ProductDetailViewModel.toggleFavorite()`：调用 `POST /api/favorite/toggle`，未登录时提示登录不发请求（已存在于 `MallApp/Views/ProductDetailView.swift`，2026-07-17 勘察确认）
- [x] 1.2 `ProductDetailView` 收藏按钮展示真实 `is_favorited` 状态并响应点击切换（已存在，同上）

## 2. 收藏 API 封装

- [ ] 2.1 新增 `MallApp/Models/Favorite.swift`：迁入 `FavoriteAPI.toggle`（自 `ProductDetail.swift`，签名不变），新增 `FavoriteAPI.list(page:pageSize:) -> ProductPage`（GET /api/favorite/list，复用 ProductPage 解码）

## 3. 收藏列表页

- [ ] 3.1 新增 `MallApp/Views/FavoriteListView.swift` 的 `FavoriteListViewModel`：分页加载（`hasMore = 已加载 < total`、loading 防重入）、`reload()` 重置第 1 页、`unfavorite()` 成功本地移除且 `total -= 1`、失败 toast 保留原项
- [ ] 3.2 `FavoriteListView` UI：卡片复用 `ProductCardView` + 右上实心红心取消收藏按钮、下拉刷新、上拉加载更多、空状态、onAppear 重载第 1 页
- [ ] 3.3 列表项点击跳转商品详情页（NavigationLink + navigationDestination）

## 4. 「我的」页面入口

- [ ] 4.1 `ProfileView` 登录态新增菜单区「我的收藏」，跳转 `FavoriteListView`；未登录整页为登录表单，入口不可见（对齐 Spec Patch 后的场景）

## 5. 验证

- [ ] 5.1 `xcodebuild build`（iOS Simulator destination）通过
- [ ] 5.2 模拟器连真后端全链路验证：登录 → 详情页收藏 → 「我的收藏」列表可见 → 列表红心取消收藏移除 → 重进详情页状态同步 → 空状态展示 → 未登录时「我的」页为登录表单（无收藏入口）
