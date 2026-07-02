## Why

`mall-ios` 目前只是纯占位骨架：三个 Tab 各自渲染 `Text(viewModel.title)`，`APIClient` 只有一个返回 mock 数据的示例方法，没有真实网络请求、没有错误处理、没有账号体系。而 `mall-mini`（微信小程序）已经是功能完整的客户端，覆盖浏览、搜索、发布、编辑、收藏、私信、个人中心共 10 个页面，并有一套已上线的 `mall-server` JWT 鉴权 API。

要把 mall-ios 对齐 mall-mini 的全部功能，体量很大，因此拆分为 5 个独立 change 依次交付（本 change 是第 1 个）。所有后续功能页面都需要统一的网络请求层（响应信封解析、错误处理）和登录态管理，如果每个功能各自实现一套网络/鉴权逻辑，会导致重复代码和不一致的错误处理。本 change 优先把这套基础设施和四 Tab 壳搭好，为后续 4 个 change（浏览/搜索/详情、发布/编辑/我的发布、收藏/个人中心、私信聊天）打底。

## What Changes

- 重写 `Core/Network/APIClient.swift`：提供通用 `request<T: Decodable>(path:method:body:requiresAuth:)` 方法，解析后端统一的 `{code, msg, data}` 响应信封，映射 HTTP 401 / 网络错误 / 解码错误为明确的错误类型。
- 新增 `Core/Network/ApiResponse.swift`（信封解码结构）与 `Core/Network/APIError.swift`（错误类型）。
- 新增 `Core/Auth/TokenStore.swift`：封装 Keychain 读写 JWT。
- 新增 `Core/Auth/AppSession.swift`：`@Observable` 会话状态，提供 `login`/`register`/`logout`/`bootstrap`，对接 `/user/login`、`/user/save`。
- 将 `ContentView.swift` 从三 Tab 扩展为四 Tab（首页/发布/消息/我的），新增「消息」占位 Tab（`Features/Chat` 新模块，仅占位视图）。
- 改造 `Features/Profile`：从占位文字变为真实的登录/注册表单（未登录态）与个人信息卡片 + 退出登录（已登录态）。
- 新增 `MallAppTests` 单元测试 target，覆盖 `APIClient` 信封解析/错误路径与 `AppSession` 的登录/注册/退出/状态恢复。

## Capabilities

### New Capabilities
- `ios-network-client`: 通用 HTTP 请求层，统一处理响应信封解析、鉴权 header 注入、错误分类（服务端错误码 / 401 / 网络传输 / 解码失败）。
- `ios-account-session`: 账号会话能力——用户名密码注册/登录/退出、JWT 在 Keychain 中的持久化、启动时会话恢复。

### Modified Capabilities
（无——mall-ios 尚无既有 spec，本 change 是首批 capability）

## Impact

- **受影响代码**：`mall-ios/Core/Network/*`（重写）、`mall-ios/Core/Auth/*`（新增）、`mall-ios/ContentView.swift`（Tab 扩展）、`mall-ios/Features/Profile/*`（改造）、`mall-ios/Features/Chat/*`（新增占位）。
- **对接的后端 API**：`POST /user/save`（注册，非信封响应）、`POST /user/login`（登录，返回信封含 token）。不修改 `mall-server` 代码。
- **依赖关系**：无前置 change；`mall-ios-browse-search-detail`、`mall-ios-publish-edit-my-listings`、`mall-ios-favorites-profile`、`mall-ios-chat-messaging` 四个后续 change 均依赖本 change 提供的网络层与 `AppSession`。
- **不影响** `mall-mini`、`mall-server`、`mall-admin-web`。
