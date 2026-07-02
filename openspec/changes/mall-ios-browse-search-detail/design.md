## Context

mall-mini 的浏览主链路由三个页面组成：`home`（首页列表 + 分类/地区筛选入口）、`search`（关键字搜索 + 排序/状态/分类/地区筛选）、`detail`（商品详情）。三者共享同一个后端搜索接口 `GET /api/product/search`（`home` 固定 `status=0`，`search` 额外支持 `keyword`/`sort`/`status`），以及同一份省市区级联数据 `data/china-regions.ts`。`detail` 页面还包含收藏按钮和"联系卖家"按钮，但这两个功能的真实实现分别属于后续的 `mall-ios-favorites-profile` 和 `mall-ios-chat-messaging` change。

`APIClient.request<T>` 与信封解析已在 `mall-ios-network-auth-foundation` 中落地，本 change 直接复用，不重新设计网络层。

## Goals / Non-Goals

**Goals:**
- 首页展示真实商品列表，支持分类/地区筛选、下拉刷新、上拉加载分页。
- 独立的搜索页：关键字、排序、状态、分类、地区筛选组合查询，分页加载。
- 商品详情页：完整展示商品信息（图片轮播、价格、描述、卖家、发布时间等）。
- 为收藏与私信预留 UI 挂载点（按钮存在，但本 change 中点击行为为禁用态或提示"即将上线"，不调用后端）。

**Non-Goals:**
- 不实现收藏切换的真实逻辑（`POST /api/favorite/toggle`），留给 `mall-ios-favorites-profile`。
- 不实现私信发起的真实逻辑（跳转到聊天页），留给 `mall-ios-chat-messaging`。
- 不实现商品举报功能（mall-mini 的 `reportProduct` 只是本地 UI 反馈，不调用后端，本 change 同样不做）。
- 不做搜索历史记录、热门搜索推荐等增强功能（mall-mini 本身也没有）。

## Decisions

### 1. 首页与搜索页复用同一套筛选组件与数据模型
- **共享 `Product` 模型**：新增 `Core/Models/Product.swift`，字段对齐后端 `/api/product/search` 返回的列表项（`id/title/price/location/images/seller/avatar/buy_uid/create_time`）。`images` 字段后端以逗号分隔字符串返回，Model 层负责 split 成 `[String]`（对齐 mall-mini `home.ts`/`search.ts` 的处理方式）。
- **共享分类常量**：`['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他']` 定义为 `Core/Models/ProductCategory.swift` 的静态数组，首页与搜索页共用，避免重复定义。
- **共享省市区数据**：新增 `Core/Data/ChinaRegions.swift`，把 `mall-mini/miniprogram/data/china-regions.ts` 的省市区三级数据转换为 Swift 静态结构体数组（一次性数据迁移，非运行时依赖小程序数据）。
- **备选方案**：首页和搜索页各自定义一套筛选逻辑。放弃：两者的筛选参数（分类/省/市/区）完全一致，且未来若分类/地区数据变化，双份定义会不同步。

### 2. 列表分页与筛选状态管理
- `HomeViewModel` 与 `SearchViewModel` 各自持有 `page`/`hasMore`/`loading` 与筛选条件，筛选变化时重置 `page = 1` 并清空当前列表，与 mall-mini 的 `home.ts`/`search.ts` 行为一致。
- 用 SwiftUI `List` 的 `.onAppear` 触发最后一行时加载下一页（对齐小程序 `onReachBottom`），下拉刷新用 `.refreshable`（对齐 `onPullDownRefresh`）。
- **备选方案**：用第三方分页库。放弃：分页逻辑简单（`page`/`page_size`/`hasMore`），手写更符合项目当前无第三方依赖的现状。

### 3. 商品详情页的收藏/联系卖家占位策略
- `ProductDetailView` 渲染收藏按钮（心形图标）与"联系卖家"按钮，但点击后仅展示 `.alert("即将上线")`，不发起网络请求。`isFavorite` 初始状态直接使用详情接口返回的 `is_favorited` 字段展示（只读展示，不可切换）。
- **备选方案**：本 change 直接隐藏收藏/联系卖家按钮，等对应 change 落地后再加回。放弃：mall-mini 的详情页布局中两者是核心操作区，直接隐藏会导致后续 change 需要改动布局结构；保留占位、禁用交互的成本更低，且与 `mall-ios-network-auth-foundation` 中"Profile 菜单项延后加入"的先例思路一致（这里选择"位置占好、交互禁用"而非"完全不出现"，因为收藏按钮是详情页视觉核心，缺失会显得页面不完整）。

## Risks / Trade-offs

- **[风险] 省市区数据从 TypeScript 手动转换为 Swift，存在转换出错的可能** → 缓解：转换后编写一个简单的完整性检查（省份数量、每省至少一个城市），在测试中校验数据结构完整。
- **[风险] `/api/product/detail` 是 `OptionalAuthMiddleware`（登录/未登录均可访问），未登录用户看到的 `is_favorited` 应为 `false`** → 缓解：直接按接口返回值展示，未登录时后端本就不会返回 `true`，iOS 侧不需要额外判断登录态。
- **[权衡] 收藏/联系卖家按钮做成禁用态而非隐藏** → 影响：用户在本 change 上线期间点击会看到"即将上线"提示；可接受，因为这是分阶段交付的过渡态，不影响核心浏览体验。

## Migration Plan

无迁移概念，全部为新增/替换占位视图，不涉及已发布数据迁移。

## Open Questions

无遗留未决问题。
