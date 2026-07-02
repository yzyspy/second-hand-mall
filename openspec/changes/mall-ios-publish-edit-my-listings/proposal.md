## Why

`mall-ios-network-auth-foundation`（change #1）提供了登录态，`mall-ios-browse-search-detail`（change #2）让用户能浏览商品，但作为「二手交易平台」，用户还需要能把自己的闲置物品发布出去、管理已发布的商品。本 change 是 5 个拆分 change 中的第 3 个，覆盖 mall-mini 的发布、编辑、我的发布三个页面，是平台双边交易能力（卖家侧）的核心闭环。

## What Changes

- 重写 `Features/Publish`：多图选择（`PHPickerViewController`，最多 9 张）+ 上传到七牛云（对接 `POST /api/upload/qiniu-token` 获取凭证后 multipart 上传）、描述/价格/分类/省市区三级联动地区/联系方式（手机号/微信/QQ）表单，提交调用 `POST /api/product/publish`；进入该 Tab 时若未登录则提示登录。
- 新增 `Features/ProductEdit`：复用发布表单组件，通过 `GET /api/product/detail?id=` 预填已有商品数据，提交调用 `PUT /api/product/update`。
- 新增 `Features/MyPublish`：对接 `GET /api/product/mine` 展示当前用户发布的商品列表（含状态标签：在售/已售出/已下架），支持分页；每项提供「编辑」「标记售出」「下架」「删除」操作，均调用 `POST /api/product/change-status`（后端删除即下架，`status=2`，与 mall-mini 行为一致）。
- 「我的」Tab 增加「我发布的」入口，跳转到 `MyPublish` 列表（此前在 change #1 中因页面不存在而未加入）。

## Capabilities

### New Capabilities
- `ios-product-publish`: 发布新商品，含多图上传、表单校验、提交。
- `ios-product-edit`: 编辑已发布商品，复用发布表单并预填数据。
- `ios-my-listings`: 我发布的商品列表管理（标记售出/下架/删除）。

### Modified Capabilities
（无——不修改 change #1/#2 已有 capability 的需求，仅新增「我的」页面入口，属于 UI 挂载，不改变 change #1 中 `ios-account-session` 的行为约定）

## Impact

- **受影响代码**：`mall-ios/Features/Publish/*`（从占位改为真实实现）、新增 `mall-ios/Features/ProductEdit/*`、新增 `mall-ios/Features/MyPublish/*`、新增 `Core/Network/QiniuUploader.swift`、修改 `Features/Profile/View/ProfileView.swift`（加入「我发布的」入口）。
- **对接的后端 API**：`POST /api/upload/qiniu-token`（公开，需带 token 已由请求头处理）、`POST /api/product/publish`（需登录）、`GET /api/product/detail`（复用 change #2 已用接口）、`PUT /api/product/update`（需登录）、`GET /api/product/mine`（需登录）、`POST /api/product/change-status`（需登录）。
- **依赖关系**：依赖 `mall-ios-network-auth-foundation`（网络层+登录态）；`ProductEdit` 复用 `mall-ios-browse-search-detail` 中定义的 `Product` 模型与地区/分类静态数据。
- **不影响** `mall-server`、`mall-mini`、`mall-admin-web`。
