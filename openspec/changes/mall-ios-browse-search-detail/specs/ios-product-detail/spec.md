## ADDED Requirements

### Requirement: 商品详情加载与展示
详情页 SHALL 根据传入的商品 ID 调用 `GET /api/product/detail` 加载数据，展示图片轮播、标题、价格、描述、成色、地区、发布时间与卖家信息（昵称、头像）。

#### Scenario: 加载成功
- **WHEN** 详情页接收到有效商品 ID 且请求成功
- **THEN** 展示该商品的完整信息

#### Scenario: 加载失败
- **WHEN** 请求返回错误信封或网络错误
- **THEN** 展示错误提示，不展示空白或崩溃

#### Scenario: 下拉刷新详情
- **WHEN** 用户在详情页下拉
- **THEN** 重新请求该商品详情并更新展示内容

### Requirement: 收藏与联系卖家的占位入口
详情页 SHALL 展示收藏按钮（依据接口返回的 `is_favorited` 只读展示当前状态）与"联系卖家"按钮，但本 capability 范围内点击这两个按钮 SHALL 仅展示"即将上线"提示，不发起任何网络请求。

#### Scenario: 点击收藏按钮
- **WHEN** 用户点击收藏按钮
- **THEN** 展示"即将上线"提示，不调用收藏接口，按钮展示状态不变

#### Scenario: 点击联系卖家按钮
- **WHEN** 用户点击"联系卖家"按钮
- **THEN** 展示"即将上线"提示，不导航到聊天页
