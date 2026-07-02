## ADDED Requirements

### Requirement: 已有会话的历史消息加载与已读标记
聊天详情页在接收到有效 `conversation_id` 时 SHALL 调用 `GET /api/chat/messages` 加载历史消息，并调用 `PUT /api/chat/read/:conv_id` 标记该会话已读。

#### Scenario: 进入已有会话
- **WHEN** 用户从会话列表进入聊天详情页（携带 `conversation_id`）
- **THEN** 加载该会话的历史消息并标记为已读

### Requirement: 新会话的延迟创建
聊天详情页在只接收到 `product_id`/`receiver_id`（无 `conversation_id`）时 SHALL 不加载历史消息，直到用户发送第一条消息；发送成功后 SHALL 使用响应返回的 `conversation_id` 开始加载消息与标记已读。

#### Scenario: 从商品详情"联系卖家"进入且尚无历史会话
- **WHEN** 用户从商品详情页发起联系卖家，且该商品与卖家之间尚无会话
- **THEN** 聊天详情页展示空消息列表，等待用户输入并发送第一条消息

#### Scenario: 发送首条消息后建立会话
- **WHEN** 用户在无 `conversation_id` 的聊天详情页发送消息且请求成功
- **THEN** 使用响应中的 `conversation_id` 加载消息、标记已读并开始轮询

### Requirement: 发送消息
聊天详情页 SHALL 提供文本输入框与发送按钮；点击发送时 SHALL 调用 `POST /api/chat/send`（携带 `product_id`/`receiver_id`/`content`），成功后清空输入框并将新消息追加到列表。

#### Scenario: 发送非空消息
- **WHEN** 用户输入非空文本并点击发送
- **THEN** 调用发送接口，成功后输入框清空，消息出现在列表末尾

#### Scenario: 发送空白消息
- **WHEN** 用户在输入框为空或全为空白字符时点击发送
- **THEN** 不发起请求

#### Scenario: 发送失败
- **WHEN** 发送请求返回错误或网络失败
- **THEN** 展示"发送失败"提示，输入框内容保留

### Requirement: 轮询增量拉取新消息
聊天详情页在会话已建立（存在 `conversation_id`）期间 SHALL 每 3 秒调用一次 `GET /api/chat/messages`（携带当前已加载的最后一条消息 `last_id`），仅追加新返回的消息，不重复渲染已有消息；页面离开时 SHALL 停止轮询。

#### Scenario: 轮询获取到新消息
- **WHEN** 轮询请求返回非空的新消息列表
- **THEN** 将新消息追加到列表末尾并更新 `last_id` 为最新一条消息的 ID

#### Scenario: 轮询未获取到新消息
- **WHEN** 轮询请求返回空列表
- **THEN** 消息列表保持不变

#### Scenario: 离开聊天详情页
- **WHEN** 用户导航离开聊天详情页
- **THEN** 停止该会话的轮询任务，不再发起后续轮询请求
