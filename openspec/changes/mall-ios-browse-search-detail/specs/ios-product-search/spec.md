## ADDED Requirements

### Requirement: 关键字搜索
搜索页 SHALL 提供关键字输入框，提交后调用 `GET /api/product/search`（携带 `keyword` 参数）并展示匹配的商品列表，支持分页加载。

#### Scenario: 输入关键字搜索
- **WHEN** 用户输入关键字并触发搜索
- **THEN** 列表重置为第 1 页并展示按关键字匹配的商品

#### Scenario: 从首页带关键字跳转进入
- **WHEN** 页面携带初始 `keyword` 参数打开
- **THEN** 自动使用该关键字执行一次搜索

### Requirement: 排序切换
搜索页 SHALL 提供排序选项（最新发布 `time_desc` / 最早发布 `time_asc`），切换后 SHALL 重置分页并按新排序重新查询。

#### Scenario: 切换为最早发布
- **WHEN** 用户选择"最早发布"
- **THEN** 列表重置为第 1 页并按 `sort=time_asc` 重新查询

### Requirement: 状态与分类/地区筛选
搜索页 SHALL 提供状态筛选（全部 / 在售 `status=0` / 已售出 `status=1`）与分类、省市区地区筛选，均可与关键字组合使用；任一筛选条件变化后 SHALL 重置分页并重新查询。

#### Scenario: 组合关键字与状态筛选
- **WHEN** 用户已输入关键字，再选择"已售出"状态
- **THEN** 查询同时携带 `keyword` 与 `status=1`，列表重置为第 1 页展示结果

### Requirement: 搜索结果分页加载
搜索页 SHALL 支持滚动到底部时基于当前累计已加载数量与服务端返回的 `total` 判断是否继续加载下一页。

#### Scenario: 已加载数量小于总数
- **WHEN** 当前已加载商品数量小于 `total`
- **THEN** 滚动到底部触发加载下一页并追加结果

#### Scenario: 已加载数量达到总数
- **WHEN** 当前已加载商品数量已达到 `total`
- **THEN** 标记没有更多数据，不再发起请求

### Requirement: 进入商品详情
搜索结果 SHALL 支持点击任意商品跳转到该商品的详情页。

#### Scenario: 点击搜索结果
- **WHEN** 用户点击搜索结果列表中的某个商品
- **THEN** 导航到该商品 ID 对应的详情页
