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
- 新增 `Features/Chat/View/ChatListView.swift` + `ViewModel/ChatListViewModel.swift` 作为占位（与 Home/Publish 现状一致），真实实现在 `mall-ios-chat-messaging` change 中完成。

### 4. 错误反馈模式
- 登录失败、注册失败、网络错误统一通过 ViewModel 的 `errorMessage: String?` 驱动 SwiftUI 原生 `.alert()` 展示。这是后续所有 change 统一遵循的错误反馈模式（替代小程序的 `wx.showToast`），在此 change 中确立。

## Risks / Trade-offs

- **[风险] `/user/save` 无唯一性校验，重复用户名会导致后端 panic → 500** → 缓解：iOS 侧把任何非 2xx 或缺少 `message` 字段的响应统一归类为注册失败，展示通用错误提示，不尝试解析具体后端错误原因。
- **[风险] 用户名密码认证与小程序的微信登录是两套独立账号体系，两端用户数据不互通** → 缓解：这是已知的产品层面限制，不在本 change 解决范围内；proposal 中已明确记录为已知限制。
- **[权衡] 不做静默重登录（不像小程序可以重新 `wx.login()` 换 token）** → 影响：用户 token 过期后需要手动重新输入密码登录；可接受，因为 24 小时 JWT 过期时间下这是低频场景。

## Migration Plan

无迁移概念——`mall-ios` 是全新客户端，本 change 内的改动全部是新增或替换占位代码，不涉及已发布版本的数据迁移或回滚策略。

## Open Questions

无遗留未决问题；已知限制（`/user/save` 校验缺失、账号体系不互通）已在 Risks 中记录并接受。
