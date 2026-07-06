# Comet Design Handoff

- Change: mall-ios-network-auth-foundation
- Phase: design
- Mode: compact
- Context hash: aaf33b15c286fabc6181f65380124045fc741a085b2526303955578f26c857a9

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/mall-ios-network-auth-foundation/proposal.md

- Source: openspec/changes/mall-ios-network-auth-foundation/proposal.md
- Lines: 1-31
- SHA256: e84050b9517ffd98d657adfb193bb4201f474ae8a15ff5eaaa1872fd61fc9f8c

```md
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
```

## openspec/changes/mall-ios-network-auth-foundation/design.md

- Source: openspec/changes/mall-ios-network-auth-foundation/design.md
- Lines: 1-98
- SHA256: a3f9b913c905e6892cd142c87205fb808112b6ff9f5b56da8aa69876073339ab

[TRUNCATED]

```md
## Context

`mall-ios` 是纯占位骨架（`Text(viewModel.title)`），`APIClient` 只有一个返回 mock 数据的 `get` 方法，没有真实网络层、错误处理或账号体系。`mall-mini` 已是功能完整的小程序，覆盖首页浏览、搜索、发布、编辑、收藏、私信、个人中心共 10 个页面。整体功能对齐工作量较大，已拆分为 5 个独立 change 依次交付，本 change 是第 1 个：网络层 + 认证 + 四 Tab 壳，为后续 4 个 change 打底。

后端 `mall-server` 已有的登录方式是微信小程序专用的 `wx.login()` code 换 token（`POST /api/user/wx-login`），iOS 上没有微信 SDK 依赖，无法复用。后端另外提供了未被小程序使用的用户名密码接口：`POST /user/save`（注册）与 `POST /user/login`（登录）。mall-ios 采用用户名密码方式，不改动后端。

`/user/save` 的响应格式是已知的后端限制：除该接口外所有接口统一返回 `{code, msg, data}` 信封（`code == 0` 成功），但 `/user/save` 直接返回 `{"message": "..."}`，不遵循信封，且后端没有用户名重复校验（保存失败会 panic，由 gin Recovery 兜底为 500）。本 change 不修改后端，iOS 按“非 2xx 或缺少预期字段即视为注册失败”处理。

## Goals / Non-Goals

**Goals:**
- 提供通用的 `APIClient.request<T>` 方法：注入 JWT、解析 `{code, msg, data}` 信封、把 HTTP 401 / 网络错误 / 解码错误映射为明确的 `APIError` case。
- 提供 `TokenStore`（Keychain 封装）与 `AppSession`（`@Observable` 会话状态：登录、注册、退出、启动恢复）。
- 把 `ContentView` 扩展为四 Tab（首页/发布/消息/我的），新增「消息」占位 Tab。
- 把「我的」Tab 从占位文字改造为真实登录/注册表单（未登录态）与个人信息卡片 + 退出登录（已登录态）。
- 补充单元测试覆盖 `APIClient` 与 `AppSession` 的核心路径。

**Non-Goals:**
- 不实现 Home / Publish / Chat 的真实业务逻辑（仍为占位，留给后续 change）。
- 不给 `/user/save` 增加用户名唯一性校验或修改响应格式（后端改动超出 mall-ios 范围）。
- 不支持 Sign in with Apple 或其他第三方认证方式。
- 不做多环境（dev/prod）baseURL 切换配置，沿用 `http://localhost:8080`。
- 不实现消息未读角标（延后到私信聊天 change 一起实现）。

## Decisions

### 1. 网络层：单一 `request<T>` 方法 + 信封解码
```swift
struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String
    let data: T?
}

enum APIError: Error {
    case server(code: Int, msg: String)   // code != 0
    case unauthorized                      // HTTP 401
    case transport(Error)                  // URLSession 失败
    case decoding(Error)                   // JSON 解析失败
}

final class APIClient {
    static let shared = APIClient()
    func request<T: Decodable>(
        _ path: String,
        method: HTTPMethod = .get,
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T
}
```
- `requiresAuth == true` 时从 `TokenStore` 读取 token，注入 `Authorization: Bearer <token>`。
- HTTP 401 → 抛出 `.unauthorized`，不做静默重试（用户名密码模式没有小程序那种静默重登录能力，调用方收到 `.unauthorized` 后触发 `AppSession.logout()`）。
- `code != 0` → 抛出 `.server(code, msg)`；否则返回 `data`。
- 单独提供 `register(username:password:) async throws -> String`，直接解码 `{"message": String}`，不复用信封解码逻辑（因为 `/user/save` 不遵循信封）。
- **备选方案**：为每个功能模块各写一个专用 Client 方法。放弃：会导致信封解析、鉴权注入、错误映射逻辑在每个 Feature 里重复，且后续 4 个 change 都要用到这套网络层，集中在一处更利于复用和测试。

### 2. 会话状态：`AppSession` 单例 + Keychain 持久化
```swift
@Observable
final class AppSession {
    static let shared = AppSession()
    private(set) var userId: Int?
    private(set) var username: String?
    private(set) var avatar: String?
    var isLoggedIn: Bool { userId != nil }

    func bootstrap()
    func login(username: String, password: String) async throws
    func register(username: String, password: String) async throws
    func logout()
}
```
- `login` 调用 `/user/login`，成功后 token 存入 `TokenStore`（Keychain），`user_id`/`user_name`/`avatar` 存入 `UserDefaults`。
- `register` 调用 `/user/save`；该接口不返回 token，注册成功后自动调用 `login` 完成登录态建立（“注册即登录”）。
- `logout` 清空 `TokenStore` 与 `UserDefaults` 中的对应字段。
- **备选方案**：把 token 直接存 `UserDefaults`。放弃：JWT 属敏感凭证，Keychain 是 iOS 标准做法，且不引入第三方依赖，封装成本很低。

### 3. Tab 结构与消息占位
- `ContentView` 扩展为四 Tab，顺序与图标对齐 mall-mini 的 `app.json`：首页(`house`)、发布(`plus.circle`)、消息(`bubble.left.and.bubble.right`)、我的(`person`)。
```

Full source: openspec/changes/mall-ios-network-auth-foundation/design.md

## openspec/changes/mall-ios-network-auth-foundation/tasks.md

- Source: openspec/changes/mall-ios-network-auth-foundation/tasks.md
- Lines: 1-36
- SHA256: 9707c2ee93b4d7b66550a348f7250b0b1c39f057f702c13eb2585845728c79c7

```md
## 1. 网络层基础

- [ ] 1.1 新增 `Core/Network/ApiResponse.swift`：`ApiResponse<T: Decodable>` 信封结构
- [ ] 1.2 重写 `Core/Network/APIError.swift`：`.server(code:msg:)` / `.unauthorized` / `.transport(Error)` / `.decoding(Error)`
- [ ] 1.3 重写 `Core/Network/APIClient.swift`：`request<T: Decodable>(path:method:body:requiresAuth:)`，处理鉴权注入、信封解析、错误映射
- [ ] 1.4 在 `APIClient` 中新增独立的 `register(username:password:) async throws -> String` 方法（非信封解析）

## 2. 账号会话

- [ ] 2.1 新增 `Core/Auth/TokenStore.swift`：Keychain save/get/delete 封装
- [ ] 2.2 新增 `Core/Auth/AppSession.swift`：`@Observable` 会话状态，实现 `login`/`register`/`logout`/`bootstrap`
- [ ] 2.3 在 `MallApp.swift` 启动时调用 `AppSession.shared.bootstrap()`

## 3. 四 Tab 壳与消息占位

- [ ] 3.1 新增 `Features/Chat/View/ChatListView.swift` 与 `Features/Chat/ViewModel/ChatListViewModel.swift`（占位）
- [ ] 3.2 修改 `ContentView.swift`：扩展为四 Tab（首页/发布/消息/我的），图标对齐 mall-mini `app.json`

## 4. 我的（Profile）真实登录/注册

- [ ] 4.1 重写 `Features/Profile/ViewModel/ProfileViewModel.swift`：未登录态表单状态、`errorMessage` 驱动的错误反馈
- [ ] 4.2 重写 `Features/Profile/View/ProfileView.swift`：未登录态登录/注册表单切换 UI
- [ ] 4.3 实现已登录态：个人信息卡片（占位头像 + 用户名）+ 退出登录按钮（含二次确认）

## 5. 测试

- [ ] 5.1 新增 `MallAppTests` XCTest target（写入 `project.yml`）
- [ ] 5.2 编写自定义 `URLProtocol` 用于拦截网络请求 mock
- [ ] 5.3 测试 `APIClient`：信封解码成功路径、`code != 0` 抛错、HTTP 401 → `.unauthorized`、`/user/save` 非信封响应解析
- [ ] 5.4 测试 `AppSession`：`login`/`register`/`logout` 状态流转，以及 `bootstrap()` 从持久化存储恢复登录态

## 6. 验证

- [ ] 6.1 `xcodebuild build` 或等效命令通过
- [ ] 6.2 运行 `MallAppTests` 全部通过
- [ ] 6.3 手动验证：注册新用户 → 自动登录 → 杀掉 App 重启 → 登录态保留 → 退出登录 → 回到登录表单
```

## openspec/changes/mall-ios-network-auth-foundation/specs/ios-account-session/spec.md

- Source: openspec/changes/mall-ios-network-auth-foundation/specs/ios-account-session/spec.md
- Lines: 1-48
- SHA256: 8ea92cbb8636f8ff7bdb1952e2de05d51dd22b6ee77c218716e7599b09acb480

```md
## ADDED Requirements

### Requirement: 用户名密码登录
`AppSession` SHALL 提供 `login(username:password:)`，调用 `POST /user/login`，成功后把返回的 JWT 存入 `TokenStore`，并把 `userId`/`username`/`avatar` 持久化到 `UserDefaults`，同时更新已发布的会话属性。

#### Scenario: 登录成功
- **WHEN** 用户输入正确的用户名密码并调用 `login`
- **THEN** `isLoggedIn` 变为 `true`，token 保存到 Keychain，用户信息保存到 `UserDefaults`

#### Scenario: 登录失败
- **WHEN** 用户名或密码错误，`/user/login` 返回业务错误信封
- **THEN** `login` 向上抛出错误，`isLoggedIn` 保持 `false`，会话状态不变

### Requirement: 注册即登录
`AppSession` SHALL 提供 `register(username:password:)`，调用 `POST /user/save`；由于该接口不返回 token，注册成功后 SHALL 自动调用 `login` 完成登录态建立。

#### Scenario: 注册成功后自动登录
- **WHEN** 用户提交新用户名密码且 `/user/save` 返回成功
- **THEN** `register` 内部自动调用 `login`，完成后 `isLoggedIn` 变为 `true`

#### Scenario: 注册失败
- **WHEN** `/user/save` 返回非 2xx 或响应格式不符合预期
- **THEN** `register` 抛出错误，不触发登录，`isLoggedIn` 保持 `false`

### Requirement: 退出登录
`AppSession` SHALL 提供 `logout()`，清空 `TokenStore` 中的 token 与 `UserDefaults` 中的用户信息字段，并重置已发布的会话属性为初始值。

#### Scenario: 用户主动退出
- **WHEN** 已登录用户调用 `logout()`
- **THEN** `isLoggedIn` 变为 `false`，`userId`/`username`/`avatar` 均重置为 `nil`，Keychain 中的 token 被删除

### Requirement: 启动时会话恢复
`AppSession` SHALL 提供 `bootstrap()`，在 App 启动时从 `TokenStore` 和 `UserDefaults` 恢复已保存的登录态，无需用户重新登录。

#### Scenario: 重启后存在已保存的登录态
- **WHEN** App 启动且 Keychain 中存在有效 token、`UserDefaults` 中存在用户信息
- **THEN** `bootstrap()` 后 `isLoggedIn` 为 `true` 且 `userId`/`username`/`avatar` 恢复为保存前的值

#### Scenario: 重启后无已保存的登录态
- **WHEN** App 启动且 Keychain 中不存在 token
- **THEN** `bootstrap()` 后 `isLoggedIn` 为 `false`

### Requirement: 401 触发自动登出
调用方在收到 `APIError.unauthorized` 时 SHALL 调用 `AppSession.shared.logout()`，使 UI 因 `isLoggedIn` 变化自动回退到登录表单。

#### Scenario: 已登录期间 token 过期
- **WHEN** 已登录用户发起需要鉴权的请求，`APIClient` 返回 `APIError.unauthorized`
- **THEN** 调用方触发 `logout()`，`isLoggedIn` 变为 `false`，界面回退到登录表单
```

## openspec/changes/mall-ios-network-auth-foundation/specs/ios-network-client/spec.md

- Source: openspec/changes/mall-ios-network-auth-foundation/specs/ios-network-client/spec.md
- Lines: 1-49
- SHA256: f46464bb413b5bd222af962a97bcb0158b62c401f35cda42a2ea0f8fa9ebab1e

```md
## ADDED Requirements

### Requirement: 统一请求信封解析
`APIClient` SHALL 对所有 `mall-server` 响应统一解析为 `{code, msg, data}` 信封结构，`code == 0` 时返回 `data`，否则抛出携带该 `code`/`msg` 的错误。

#### Scenario: 服务端返回成功信封
- **WHEN** 请求收到 HTTP 2xx 且响应体为 `{"code": 0, "msg": "ok", "data": {...}}`
- **THEN** `request<T>` 返回解码后的 `data`

#### Scenario: 服务端返回业务错误信封
- **WHEN** 请求收到 HTTP 2xx 但响应体为 `{"code": 1001, "msg": "参数错误", "data": null}`
- **THEN** `request<T>` 抛出 `APIError.server(code: 1001, msg: "参数错误")`

### Requirement: 鉴权 Header 注入
`APIClient` SHALL 在 `requiresAuth == true` 时从 `TokenStore` 读取 JWT 并注入 `Authorization: Bearer <token>` 请求头；`requiresAuth == false` 时不注入。

#### Scenario: 需要鉴权的请求携带 token
- **WHEN** 调用 `request(path:, requiresAuth: true)` 且 `TokenStore` 中存在已保存的 token
- **THEN** 发出的 HTTP 请求头包含 `Authorization: Bearer <token>`

#### Scenario: 公开接口不注入鉴权头
- **WHEN** 调用 `request(path:, requiresAuth: false)`
- **THEN** 发出的 HTTP 请求不包含 `Authorization` 头

### Requirement: 错误分类映射
`APIClient` SHALL 将网络层各类失败映射为明确的 `APIError` case：HTTP 401 映射为 `.unauthorized`，URLSession 传输失败映射为 `.transport`，JSON 解码失败映射为 `.decoding`。

#### Scenario: 收到 401 响应
- **WHEN** 请求收到 HTTP 401
- **THEN** `request<T>` 抛出 `APIError.unauthorized`，不进行任何自动重试

#### Scenario: 网络连接失败
- **WHEN** `URLSession` 因无网络连接等原因抛出错误
- **THEN** `request<T>` 抛出 `APIError.transport(underlyingError)`

#### Scenario: 响应体无法解码为期望类型
- **WHEN** 响应体的 JSON 结构与期望的 `T` 类型不匹配
- **THEN** `request<T>` 抛出 `APIError.decoding(underlyingError)`

### Requirement: 非信封响应的注册接口
`APIClient` SHALL 为 `/user/save` 提供单独的 `register` 方法，直接解码 `{"message": String}` 格式，不套用信封解析逻辑；任何非 2xx 响应或缺少 `message` 字段的响应均视为注册失败。

#### Scenario: 注册成功
- **WHEN** `POST /user/save` 返回 HTTP 2xx 且响应体包含 `message` 字段
- **THEN** `register` 返回该 `message` 字符串

#### Scenario: 注册失败（用户名重复导致后端 500）
- **WHEN** `POST /user/save` 返回非 2xx 状态码
- **THEN** `register` 抛出错误，调用方展示通用注册失败提示，不解析具体后端错误原因
```

