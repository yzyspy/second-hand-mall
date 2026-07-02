## ADDED Requirements

### Requirement: 我发布的商品列表加载
「我发布的」页面 SHALL 调用 `GET /api/product/mine` 分页加载当前登录用户发布的商品，展示标题、价格、封面图与状态标签（在售/已售出/已下架）。

#### Scenario: 首次进入列表
- **WHEN** 用户打开「我发布的」页面
- **THEN** 展示第一页商品列表及各自状态标签

#### Scenario: 下拉刷新
- **WHEN** 用户下拉刷新列表
- **THEN** 重置为第 1 页并重新加载

#### Scenario: 滚动到底部加载更多
- **WHEN** 已加载数量小于服务端返回的 `total` 且用户滚动到底部
- **THEN** 加载下一页并追加到列表

### Requirement: 跳转编辑
列表每一项 SHALL 提供「编辑」操作，点击后跳转到编辑页并传入该商品 ID。

#### Scenario: 点击编辑
- **WHEN** 用户点击某商品的「编辑」按钮
- **THEN** 导航到编辑页并加载该商品数据

### Requirement: 标记售出
列表每一项 SHALL 提供「标记售出」操作，点击后二次确认，确认后调用 `POST /api/product/change-status`（`status=1`），成功后重新加载列表。

#### Scenario: 确认标记售出
- **WHEN** 用户点击「标记售出」并在确认弹窗中确认
- **THEN** 调用状态变更接口，成功后展示"已标记为售出"提示并刷新列表

#### Scenario: 取消确认
- **WHEN** 用户点击「标记售出」但在确认弹窗中取消
- **THEN** 不发起任何请求，列表保持不变

### Requirement: 下架商品
列表每一项 SHALL 提供「下架」操作，点击后二次确认，确认后调用 `POST /api/product/change-status`（`status=2`），成功后重新加载列表。

#### Scenario: 确认下架
- **WHEN** 用户点击「下架」并确认
- **THEN** 调用状态变更接口，成功后展示"已下架"提示并刷新列表

### Requirement: 删除商品（语义为下架）
列表每一项 SHALL 提供「删除」操作，点击后二次确认（确认按钮标红），确认后调用 `POST /api/product/change-status`（`status=2`，与下架相同的后端接口），成功后重新加载列表。

#### Scenario: 确认删除
- **WHEN** 用户点击「删除」并在标红确认按钮的弹窗中确认
- **THEN** 调用状态变更接口，成功后展示"已删除"提示并刷新列表，该商品的状态在服务端变为已下架
