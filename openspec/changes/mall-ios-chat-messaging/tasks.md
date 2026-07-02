## 1. 会话列表

- [ ] 1.1 重写 `Features/Chat/ViewModel/ChatListViewModel.swift`：对接 `GET /api/chat/conversations`
- [ ] 1.2 重写 `Features/Chat/View/ChatListView.swift`：会话列表展示（商品/对方信息/最后消息/未读数）、空状态、加载失败提示

## 2. 聊天详情

- [ ] 2.1 新增 `Features/ChatDetail/ViewModel/ChatDetailViewModel.swift`：处理"已有会话"与"新会话"两种入参、加载消息、发送消息、已读标记
- [ ] 2.2 实现 3 秒轮询增量拉取（基于 `last_id`），视图消失时取消轮询任务
- [ ] 2.3 新增 `Features/ChatDetail/View/ChatDetailView.swift`：消息列表、输入框、发送按钮

## 3. 联系卖家接入

- [ ] 3.1 `ProductDetailViewModel`/`ProductDetailView`「联系卖家」按钮改为导航到 `ChatDetailView`（传入 `product_id`/`receiver_id`），未登录提示登录

## 4. 未读角标

- [ ] 4.1 `AppSession` 新增 `unreadChatCount` 已发布属性与刷新方法（对接 `GET /api/chat/unread-count`）
- [ ] 4.2 App 进入前台（`scenePhase == .active`）与会话列表加载完成后触发未读数刷新
- [ ] 4.3 `ContentView` 消息 Tab 绑定 `.badge()`，未读数为 0 时不展示

## 5. 测试

- [ ] 5.1 测试 `ChatListViewModel`：加载成功/失败/空列表
- [ ] 5.2 测试 `ChatDetailViewModel`：已有会话加载消息+已读、无会话延迟创建、发送成功后建立会话并开始轮询、轮询去重追加
- [ ] 5.3 测试未读角标刷新逻辑（0 → 无角标，>0 → 展示数字）

## 6. 验证

- [ ] 6.1 `xcodebuild build` 或等效命令通过
- [ ] 6.2 运行单元测试全部通过
- [ ] 6.3 手动验证：详情页"联系卖家" → 发送首条消息建立会话 → 会话列表出现该会话 → 进入聊天收发消息 → 消息 Tab 角标随未读数变化 → 离开聊天页轮询停止
