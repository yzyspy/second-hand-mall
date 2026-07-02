## ADDED Requirements

### Requirement: 会话列表加载
消息 Tab SHALL 在展示时调用 `GET /api/chat/conversations` 加载当前登录用户的会话列表，展示商品缩略图/标题、对方昵称/头像、最后一条消息内容与时间、未读数。

#### Scenario: 加载成功
- **WHEN** 用户切换到消息 Tab 且请求成功
- **THEN** 展示会话列表，按最后消息时间排列

#### Scenario: 加载失败
- **WHEN** 请求返回错误或网络失败
- **THEN** 展示加载失败提示

#### Scenario: 会话列表为空
- **WHEN** 用户没有任何会话
- **THEN** 展示空状态提示

### Requirement: 进入聊天详情
会话列表 SHALL 支持点击任意会话跳转到聊天详情页，并传入该会话的 `conversation_id`、`product_id`、`receiver_id`。

#### Scenario: 点击某个会话
- **WHEN** 用户点击会话列表中的一项
- **THEN** 导航到聊天详情页并加载该会话的历史消息
