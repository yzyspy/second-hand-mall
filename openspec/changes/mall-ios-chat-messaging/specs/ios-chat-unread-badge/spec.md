## ADDED Requirements

### Requirement: 消息 Tab 未读角标
App SHALL 在进入前台时以及会话列表刷新完成后调用 `GET /api/chat/unread-count`，把返回的未读数展示在消息 Tab 的角标上；未读数为 0 时 SHALL 不展示角标。

#### Scenario: 存在未读消息
- **WHEN** 未读数查询返回大于 0 的数值
- **THEN** 消息 Tab 展示对应数字角标

#### Scenario: 无未读消息
- **WHEN** 未读数查询返回 0
- **THEN** 消息 Tab 不展示角标（若之前展示过则移除）

#### Scenario: App 从后台回到前台
- **WHEN** App 状态从后台变为前台激活
- **THEN** 重新查询未读数并更新角标
