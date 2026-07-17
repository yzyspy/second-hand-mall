# Brainstorm Summary

- Change: mall-ios-favorites-profile
- Date: 2026-07-17
- 状态: 用户已于 2026-07-17 确认设计方案（Step 1c 通过），已定稿

## 已确认事实（代码勘察）

1. **详情页收藏切换已实现**：`MallApp/Views/ProductDetailView.swift` 的 `ProductDetailViewModel.toggleFavorite()` 已调用 `FavoriteAPI.toggle`（`MallApp/Models/ProductDetail.swift:89`），含未登录拦截（`API.token` 为空时 toast「请先登录」）。proposal 中"占位改真实交互"的假设已过时——tasks 1.1/1.2 实际已完成。
2. **实际目录结构与 proposal 不符**：当前 iOS 代码是扁平结构 `MallApp/Views/*.swift` + `MallApp/Models/*.swift`（commit 693a2cc "iOS done" 重写），不存在 `Features/...` 结构，也无「我发布的」入口。
3. **无测试目标**：MallApp.xcodeproj 无任何 test target。
4. **后端接口确认**：`GET /api/favorite/list`（`page`/`page_size`，需登录）返回 `{list: []ProductSearchResult, total, page, page_size}`，与 `/api/product/search` 同构 → iOS 现有 `Product`/`ProductPage` 模型直接复用解码。`POST /api/favorite/toggle` 返回 `{is_favorited}`。
5. **ProfileView 现状**：登录态只有用户卡片+退出按钮，无菜单区；未登录态整页是登录表单 → spec「未登录点击入口提示登录」场景在 iOS 交互下不可达（未登录看不到菜单）。
6. mall-mini `myFavorite.ts` 基准：`hasMore = 已加载 < total`、下拉刷新重置、onShow 全量重载、取消收藏成功本地 filter 移除且 total-1、失败保留原项。

## 用户已确认的决策（AskUserQuestion）

- **Q1 范围**：只做缺失部分（FavoriteAPI.list + 收藏列表页 + 「我的」入口）；详情页收藏按"已实现"处理，验证阶段一并覆盖；Spec Patch 说明现状。
- **Q2 取消收藏交互**：卡片右上实心红心按钮，点击取消收藏并本地移除（对齐 mall-mini）。
- **Q3 测试策略**：不新建 test target；xcodebuild 构建通过 + 模拟器真后端手动/脚本验证（DEBUG_* 环境变量直达）；修订 tasks.md 去掉单测项。

## 确认的技术方案（待用户最终确认）

- 新增 `MallApp/Models/Favorite.swift`：`FavoriteAPI.list(page:pageSize:) -> ProductPage`（GET /api/favorite/list，复用 ProductPage 解码）；现有 `FavoriteAPI.toggle` 从 ProductDetail.swift 迁入或保留（倾向迁入集中）。
- 新增 `MallApp/Views/FavoriteListView.swift`：`FavoriteListViewModel`（@Observable，分页/hasMore=已加载<total/防重入/下拉刷新/unfavorite 本地移除 total-1、失败 toast 保留原项）+ 列表 UI（复用 ProductCardView 布局模式 + 右上红心按钮覆盖）、空状态、点击 NavigationLink 进 ProductDetailView。
- `ProfileView` 登录态增加菜单区：「我的收藏」行（NavigationLink → FavoriteListView）。未登录整页即登录表单，入口不可见——以 Spec Patch 调整「未登录点击提示登录」场景为「未登录不可见入口」。
- 列表在每次 onAppear 重载第 1 页（对齐 mall-mini onShow 行为），保证从详情页取消收藏返回后状态一致。

## 关键取舍与风险

- 取消收藏本地移除 vs 重新拉取：沿用 mall-mini 本地移除；并发不一致可接受，下拉刷新兜底。
- onAppear 全量重载会重置滚动位置：与 mall-mini 行为一致，列表页数据量小，可接受。
- 无单测：与仓库现状一致，验证依赖构建+模拟器全链路。

## 测试策略

xcodebuild 构建通过 + 模拟器连真后端验证：收藏 → 列表可见 → 列表取消 → 移除且详情页状态同步（重进详情）→ 空状态 → 分页（如数据足够）。

## Spec Patch（候选，待确认后回写）

1. `specs/ios-product-detail/spec.md`：MODIFIED requirement 保持（现状已满足，无需改文本，验证覆盖即可）——不改。
2. `specs/ios-product-favorite/spec.md`：「我的」页面收藏入口 requirement 的未登录场景改为「未登录时『我的』页面展示登录表单，收藏入口不可见，无法进入收藏列表」。
3. `tasks.md`：按已确认范围重写（详情页任务标记已完成/验证项，去掉单测任务，目录路径改为实际扁平结构）。
