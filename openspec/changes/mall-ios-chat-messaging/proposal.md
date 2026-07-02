## Why

前四个 change 完成了浏览、发布、收藏三大能力，但二手交易的最后一环——买卖双方就商品细节沟通——还未实现。`mall-ios-network-auth-foundation`（change #1）预留了「消息」占位 Tab，`mall-ios-browse-search-detail`（change #2）在详情页留了"联系卖家"占位按钮。本 change 是 5 个拆分 change 中的最后一个，实现真正的私信聊天能力，闭环整个 mall-mini 功能对齐工作。

## What Changes

- 重写 `Features/Chat`（会话列表）：对接 `GET /api/chat/conversations` 展示会话列表（商品缩略图/标题、对方昵称/头像、最后一条消息、未读数）。
- 新增 `Features/ChatDetail`（聊天详情）：对接 `GET /api/chat/messages`（按 `last_id` 增量拉取）、`POST /api/chat/send`（发送消息）、`PUT /api/chat/read/:conv_id`（标记已读），进入会话后启动轮询（3 秒间隔，对齐 mall-mini）拉取新消息，离开页面停止轮询。
- 商品详情页"联系卖家"按钮从占位改为真实交互：调用发送消息接口发起新会话（或跳转到已有会话），进入聊天详情页。
- 消息 Tab 增加未读消息角标：对接 `GET /api/chat/unread-count`，在 App 前台/会话列表刷新时更新角标数字，无未读时不显示角标。

## Capabilities

### New Capabilities
- `ios-chat-conversations`: 会话列表加载与展示。
- `ios-chat-messaging`: 单个会话内的消息收发、已读标记、轮询增量拉取。
- `ios-chat-unread-badge`: 消息 Tab 未读数角标。

### Modified Capabilities
- `ios-product-detail`: "联系卖家"按钮从"点击展示即将上线"改为真实发起会话/跳转聊天——这是 `mall-ios-browse-search-detail` 中 `ios-product-detail` capability 的既有需求变更。

## Impact

- **受影响代码**：重写 `mall-ios/Features/Chat/*`（会话列表）、新增 `mall-ios/Features/ChatDetail/*`、修改 `mall-ios/Features/ProductDetail/*`（联系卖家按钮）、修改 `mall-ios/ContentView.swift`（消息 Tab 角标）。
- **对接的后端 API**：`GET /api/chat/conversations`、`GET /api/chat/messages`、`POST /api/chat/send`、`PUT /api/chat/read/:conv_id`、`GET /api/chat/unread-count`（均需登录）。
- **依赖关系**：依赖 `mall-ios-network-auth-foundation`（登录态、四 Tab 壳中的消息占位 Tab）、`mall-ios-browse-search-detail`（详情页"联系卖家"按钮挂载点）。
- **不影响** `mall-server`、`mall-mini`、`mall-admin-web`。本 change 完成后，5 个拆分 change 累计覆盖 mall-mini 的全部页面功能。
