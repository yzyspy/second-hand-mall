# Comet Design Handoff

- Change: mall-ios-favorites-profile
- Phase: design
- Mode: compact
- Context hash: 45e976fb0f8442fbd07e156f64499ef8965cb0b5cc7da297179690c0b7edae47

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/mall-ios-favorites-profile/proposal.md

- Source: openspec/changes/mall-ios-favorites-profile/proposal.md
- Lines: 1-24
- SHA256: b3a55744049a4e6e19e7e63419b00e05e06ff9b642a4ffd57ecd3b7c49297e2f

```md
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
```

## openspec/changes/mall-ios-favorites-profile/design.md

- Source: openspec/changes/mall-ios-favorites-profile/design.md
- Lines: 1-44
- SHA256: 031b0d05d16cd16b50fedf879b8967269c83f04acfdcf8df5cffdc37efb2092d

```md
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
```

## openspec/changes/mall-ios-favorites-profile/tasks.md

- Source: openspec/changes/mall-ios-favorites-profile/tasks.md
- Lines: 1-23
- SHA256: 7300492fef679eddf726bbe5b748f73536e0c94197dc9b40dba1eb27d465825d

```md
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
```

## openspec/changes/mall-ios-favorites-profile/specs/ios-product-detail/spec.md

- Source: openspec/changes/mall-ios-favorites-profile/specs/ios-product-detail/spec.md
- Lines: 1-16
- SHA256: f615ce9f41d792ae87a238339967b89d2105644d81b6672ca5d3640631520e72

```md
## MODIFIED Requirements

### Requirement: 收藏与联系卖家的占位入口
详情页 SHALL 展示收藏按钮与"联系卖家"按钮。收藏按钮 SHALL 展示真实收藏状态并支持切换：点击后调用收藏切换接口，成功后按钮态更新为最新的收藏状态；未登录用户点击收藏按钮 SHALL 提示登录，不发起请求。"联系卖家"按钮在本 capability 范围内仍为占位：点击仅展示"即将上线"提示，不发起任何网络请求或页面跳转（真实私信能力由后续的私信聊天 capability 提供）。

#### Scenario: 已登录用户点击收藏按钮
- **WHEN** 已登录用户点击收藏按钮
- **THEN** 调用收藏切换接口，成功后按钮展示状态在"已收藏"与"未收藏"之间切换

#### Scenario: 未登录用户点击收藏按钮
- **WHEN** 未登录用户点击收藏按钮
- **THEN** 展示登录提示，不发起收藏请求，按钮状态不变

#### Scenario: 点击联系卖家按钮
- **WHEN** 用户点击"联系卖家"按钮
- **THEN** 展示"即将上线"提示，不导航到聊天页
```

## openspec/changes/mall-ios-favorites-profile/specs/ios-product-favorite/spec.md

- Source: openspec/changes/mall-ios-favorites-profile/specs/ios-product-favorite/spec.md
- Lines: 1-49
- SHA256: 0c414d6e68e993e6c1e89539aedb682358c937b6f51c40160e1c9fe9c2bd3892

```md
## ADDED Requirements

### Requirement: 收藏列表加载
「我的收藏」页面 SHALL 调用 `GET /api/favorite/list` 分页加载当前登录用户收藏的商品，展示标题、价格、地区、卖家、封面图。

#### Scenario: 首次进入收藏列表
- **WHEN** 用户打开「我的收藏」页面
- **THEN** 展示第一页收藏商品列表

#### Scenario: 下拉刷新
- **WHEN** 用户下拉刷新收藏列表
- **THEN** 重置为第 1 页并重新加载

#### Scenario: 滚动到底部加载更多
- **WHEN** 已加载数量小于服务端返回的 `total` 且用户滚动到底部
- **THEN** 加载下一页并追加到列表

#### Scenario: 收藏列表为空
- **WHEN** 用户没有任何收藏商品
- **THEN** 展示空状态提示

### Requirement: 列表内取消收藏
「我的收藏」列表每一项 SHALL 提供取消收藏操作，点击后调用 `POST /api/favorite/toggle`，成功后直接从本地列表移除该项，不重新拉取整个列表。

#### Scenario: 取消收藏成功
- **WHEN** 用户在收藏列表中点击某项的取消收藏操作且请求成功
- **THEN** 该项从列表中移除，列表总数减一

#### Scenario: 取消收藏请求失败
- **WHEN** 取消收藏请求返回错误或网络失败
- **THEN** 该项保留在列表中，展示失败提示

### Requirement: 进入商品详情
收藏列表 SHALL 支持点击任意商品跳转到该商品的详情页。

#### Scenario: 点击收藏项
- **WHEN** 用户点击收藏列表中的某个商品
- **THEN** 导航到该商品 ID 对应的详情页

### Requirement: 「我的」页面收藏入口
「我的」页面 SHALL 在已登录状态下提供「我的收藏」菜单入口；未登录状态下「我的」页面整页展示登录表单，收藏入口 SHALL 不可见，用户无法进入收藏列表。

#### Scenario: 已登录点击入口
- **WHEN** 已登录用户点击「我的收藏」菜单项
- **THEN** 导航到收藏列表页

#### Scenario: 未登录访问「我的」页面
- **WHEN** 未登录用户打开「我的」Tab
- **THEN** 页面展示登录表单，不展示「我的收藏」入口，无法进入收藏列表
```

