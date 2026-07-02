## Context

mall-mini 的私信功能由三部分组成：`chat-list.ts`（会话列表，`GET /api/chat/conversations`）、`chat.ts`（聊天详情，`GET /api/chat/messages?conv_id=&last_id=` 增量拉取 + `POST /api/chat/send` 发送 + `PUT /api/chat/read/:conv_id` 已读 + 3 秒轮询）、`app.ts`（启动时查询 `GET /api/chat/unread-count` 设置 Tab 角标）。会话可以通过两种方式进入：从会话列表点击已有会话（带 `conversation_id`），或从商品详情页"联系卖家"直接发起新会话（只带 `product_id`/`receiver_id`，首次发送消息时后端会创建新会话并返回 `conversation_id`）。

`mall-ios-network-auth-foundation` 已经预留了消息 Tab 占位（`ChatListView`/`ChatListViewModel`），本 change 替换为真实实现；`mall-ios-browse-search-detail` 的详情页"联系卖家"按钮占位在本 change 中接上真实跳转。

## Goals / Non-Goals

**Goals:**
- 会话列表：加载并展示会话（商品信息、对方信息、最后消息、未读数），点击进入聊天详情。
- 聊天详情：加载历史消息、发送消息、进入会话自动标记已读、轮询增量拉取新消息，离开页面停止轮询。
- 从商品详情页"联系卖家"发起新会话（无已有 `conversation_id` 时，首次发送消息由后端创建会话）。
- 消息 Tab 角标：展示未读会话总数，为 0 时不展示角标。

**Non-Goals:**
- 不实现推送通知（后端没有 WebSocket/APNs 推送机制，mall-mini 本身也是轮询实现）。
- 不做消息撤回、已读回执细粒度展示、图片/表情消息（mall-mini 仅支持纯文本消息）。
- 不做会话删除/免打扰等管理功能（mall-mini 没有）。
- 不优化轮询为长连接/WebSocket（后端接口本身是轮询设计，超出 iOS 端范围）。

## Decisions

### 1. 轮询策略与生命周期对齐 SwiftUI 视图生命周期
- `ChatDetailViewModel` 在聊天详情视图 `.task` 修饰符启动时开始 3 秒间隔的 `Task` 轮询循环（对齐 mall-mini `setInterval(3000)`），视图消失（`.onDisappear` / `.task` 被取消）时取消轮询任务。
- 增量拉取用 `last_id` 记录已加载的最后一条消息 ID，每次轮询携带该 `last_id`，只追加新消息，不重复渲染已有消息（对齐 `chat.ts` 的 `loadMessages` 逻辑）。
- **备选方案**：用 `Timer` 而非 Swift 结构化并发 `Task`。放弃：SwiftUI 生命周期下用 `.task` 自动取消更符合项目已确立的 `async/await` 网络层风格，避免手动管理 `Timer` 失效。

### 2. 发起新会话与已有会话的统一处理
- `ChatDetailViewModel` 接受三种可能的入参组合：`(conversationId)` 从会话列表进入；`(productId, receiverId)` 从商品详情"联系卖家"进入（无 `conversationId`）。
- 有 `conversationId` 时：立即加载历史消息、标记已读、启动轮询。
- 无 `conversationId` 时：不加载历史消息，等待用户发送第一条消息；`POST /api/chat/send` 返回的 `conversation_id` 用于后续加载消息、标记已读、启动轮询（对齐 `chat.ts` 的 `onSend` 逻辑：`if (!this.pollTimer) { ... }` 首次发送后才开始轮询）。
- **备选方案**：详情页"联系卖家"先调用一个"创建会话"接口拿到 `conversation_id` 再进入聊天页。放弃：后端没有独立的创建会话接口，会话是发送首条消息时隐式创建的，遵循后端已有设计。

### 3. 未读角标更新时机
- `AppSession`（change #1 已定义）新增 `unreadChatCount: Int` 已发布属性；在 App 进入前台（`scenePhase == .active`）与会话列表加载完成后调用 `GET /api/chat/unread-count` 刷新该属性，`ContentView` 的消息 `Tab` 用 `.badge(count > 0 ? "\(count)" : nil)` 展示（SwiftUI 原生 API，对齐 `wx.setTabBarBadge`/`removeTabBarBadge` 的"有则显示、无则移除"语义）。
- **备选方案**：轮询未读数（额外定时器）。放弃：mall-mini 本身也只在启动时查询一次，未做持续轮询未读数；本 change 对齐现状，在关键时机（前台激活、会话列表刷新）刷新即可，不引入额外常驻定时器。

## Risks / Trade-offs

- **[风险] 轮询而非长连接，最坏情况下消息延迟可达 3 秒** → 缓解：这是后端接口设计的既有限制，与 mall-mini 表现一致，不在本 change 解决范围。
- **[风险] 多个聊天详情页同时轮询（如用户快速进出多个会话）可能产生并发请求堆积** → 缓解：`.task` 生命周期保证前一个视图消失时轮询任务被取消，SwiftUI 同一时间只会展示一个聊天详情页（导航栈内非当前页面不会持续轮询）。
- **[权衡] 未读角标只在前台激活和会话列表刷新时更新，不做实时性保证** → 影响：角标数字可能短暂滞后于实际未读数；可接受，因为 mall-mini 本身也只在启动时更新一次，本 change 已经比现状更及时。

## Migration Plan

无迁移概念，全部为新增视图与网络对接，以及把 change #2 的"联系卖家"占位替换为真实交互。

## Open Questions

无遗留未决问题。
