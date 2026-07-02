## Context

mall-mini 的收藏能力由三处组成：详情页的收藏切换按钮（`POST /api/favorite/toggle`）、独立的「我的收藏」列表页（`myFavorite.ts`，对接 `GET /api/favorite/list`，支持分页和列表内取消收藏）、「我的」页面的菜单入口。`mall-ios-browse-search-detail`（change #2）已经在详情页放好了收藏按钮的占位（只读展示 `is_favorited`，点击提示"即将上线"），本 change 把它接上真实接口，并新增收藏列表页。

「我的」页面 mall-mini 的 `my.ts` 中还列出了「我卖出的」「我买到的」「收货地址」「设置」「联系客服」「关于」等菜单项，但这些在 mall-mini 里本身就是指向不存在页面的死链接（`app.json` 未注册对应页面）或仅是本地占位交互（`联系客服` 只是打电话）。这些不属于本 change 范围——不存在的功能没有必要在 iOS 上"忠实复刻"死链接。

## Goals / Non-Goals

**Goals:**
- 详情页收藏按钮：点击切换收藏状态，成功后立即更新按钮展示（已收藏/未收藏）。
- 收藏列表页：分页展示收藏商品、下拉刷新、点击进入详情、列表内直接取消收藏并从列表移除。
- 「我的」页面新增「我的收藏」菜单入口，未登录点击提示登录。

**Non-Goals:**
- 不实现「我卖出的」「我买到的」「收货地址」「设置」「联系客服」「关于」——这些在 mall-mini 中本就是不存在页面的死链接或与后端无关的本地交互，不在 mall-server 当前 API 范围内，复刻它们没有实际价值。
- 不做收藏数量的角标提示（mall-mini 也没有）。
- 不做收藏分类/排序功能（mall-mini 的收藏列表就是简单的时间倒序分页列表）。

## Decisions

### 1. 详情页收藏按钮从占位切换为真实交互
- `ProductDetailViewModel` 新增 `toggleFavorite()` 方法，调用 `APIClient.request` 访问 `POST /api/favorite/toggle`（`requiresAuth: true`），成功后用返回的 `is_favorited` 更新本地状态；未登录时点击直接提示登录（不发请求）。
- **备选方案**：沿用 change #2 的占位提示，只在收藏列表页支持真实收藏（用户只能取消收藏，不能从详情页新增收藏）。放弃：收藏的主要入口在浏览路径中（详情页），只做列表页管理没有实际收藏能力，不符合产品预期。

### 2. 收藏列表：服务端分页 + 本地移除
- `FavoriteViewModel` 对接 `GET /api/favorite/list`（`page`/`page_size`），分页与下拉刷新逻辑与 `MyPublishViewModel`（change #3）一致的模式：加载中防重入、`hasMore` 基于累计已加载数量与 `total` 比较。
- 列表内取消收藏（`onUnfavorite`）SHALL 直接从本地列表移除该项并 `total -= 1`，不重新拉取整个列表——这与 `mall-ios-publish-edit-my-listings` 中「我的发布」选择"重新拉取列表"的策略不同，因为取消收藏是幂等的单向操作（移除后不会再次出现），本地移除足以保证一致性，且与 mall-mini `myFavorite.ts` 的 `onUnfavorite` 实现一致（直接 `filter` 本地数组）。
- **备选方案**：取消收藏后也重新拉取整个列表（与我的发布保持一致的策略）。放弃：会引入不必要的网络请求，且 mall-mini 本身对这两个列表采用了不同策略（我的发布重新拉取，收藏列表本地移除），忠实对齐现状。

### 3. 「我的」菜单只新增确实存在的功能入口
- 仅新增「我的收藏」一项菜单（「我发布的」已在 change #3 加入）。不新增指向不存在功能的菜单项，避免死点击。

## Risks / Trade-offs

- **[风险] 收藏列表本地移除与服务端状态可能因并发操作（如另一设备同时取消收藏）产生短暂不一致** → 缓解：可接受，用户下次下拉刷新或重新进入页面会同步最新状态；这是 mall-mini 现有行为的延续，不在本 change 引入新风险。
- **[权衡] 不复刻 mall-mini 菜单中的死链接项** → 影响：iOS 版「我的」页面菜单项少于 mall-mini 视觉呈现，但功能完全对等（不存在的功能两端都不可用），符合"参照 mall-mini 实现全部功能"的真实意图（功能对等而非像素级复刻死链接）。

## Migration Plan

无迁移概念，全部为新增视图与网络对接，以及把 change #2 的占位交互替换为真实交互。

## Open Questions

无遗留未决问题。
