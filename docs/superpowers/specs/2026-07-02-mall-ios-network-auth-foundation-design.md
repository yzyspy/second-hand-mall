---
comet_change: mall-ios-network-auth-foundation
role: technical-design
canonical_spec: openspec
---

# mall-ios-network-auth-foundation 技术设计

需求与验收场景的权威来源是 OpenSpec：`openspec/changes/mall-ios-network-auth-foundation/proposal.md`、`design.md`、`specs/ios-network-client/spec.md`、`specs/ios-account-session/spec.md`。本文档是在此基础上做的实现层技术设计确认，补充 OpenSpec design.md 未完全钉死的具体决策。

## 架构概览

- `Core/Network/`：`ApiResponse<T>`（信封）、`APIError`、`HTTPMethod`、`APIClient`（单例）。
- `Core/Auth/`：`TokenStore`（Keychain 封装）、`AppSession`（`@Observable` 会话状态）。
- `Features/Chat/`：消息 Tab 占位（`ChatListView`/`ChatListViewModel`），真实实现留给后续 change。
- `Features/Profile/`：登录/注册/已登录态改造。
- `ContentView.swift`：四 Tab 壳。
- `MallAppTests/`：新增测试 target。

## 关键实现决策

### 1. 空数据响应：`EmptyResponse` 标记类型
后端某些接口只需要 `code == 0` 语义成功，`data` 为 `null`。为避免非可选 `T` 类型解码歧义，新增：
```swift
struct EmptyResponse: Decodable {}
```
所有此类接口统一以 `T = EmptyResponse` 调用 `request`，而非把 `T` 声明为 Optional（避免调用方多一层解包）。

### 2. `HTTPMethod` 一次性定义完整集合
```swift
enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}
```
本 change 只用到 `.get`/`.post`，但 `PUT`/`DELETE` 一并定义，为后续 change #3（`PUT /api/product/update`）等预留，避免回头修改本 change 已完成并测试过的基础设施文件。

### 3. `TokenStore`：Keychain 标识符
- `kSecClass = kSecClassGenericPassword`
- `kSecAttrService = Bundle.main.bundleIdentifier`（即 `com.secondhandmall.MallApp`）
- `kSecAttrAccount = "jwt"`（固定字符串——单用户单 token 场景，不需要按用户名区分账号项）
- 仅封装 save / get / delete 三个操作，不引入第三方依赖。

### 4. `APIClient.request<T: Decodable>`
```swift
func request<T: Decodable>(
    _ path: String,
    method: HTTPMethod = .get,
    body: Encodable? = nil,
    requiresAuth: Bool = false
) async throws -> T
```
- `requiresAuth == true` 时从 `TokenStore` 读取 token 注入 `Authorization: Bearer <token>`；`token` 缺失时直接抛 `.unauthorized`，不发起网络请求。
- 解码 `ApiResponse<T>`；`code != 0` → `.server(code, msg)`；HTTP 401 → `.unauthorized`（不做静默重试）；`URLSession` 失败 → `.transport`；JSON 解码失败 → `.decoding`。
- 独立 `register(username:password:) async throws -> String` 方法：直接解码 `{"message": String}`，任何非 2xx 响应或缺少 `message` 字段均抛出统一的注册失败错误（不复用信封解码逻辑，因为 `/user/save` 不遵循信封）。

### 5. `AppSession`
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
- `login` 成功后：token 存入 `TokenStore`，`userId`/`username`/`avatar` 存入 `UserDefaults`，更新已发布属性。
- `register` 成功后自动调用 `login`（"注册即登录"）。
- `logout` 清空 `TokenStore` 与 `UserDefaults` 对应字段，重置已发布属性。
- `bootstrap()` 在 `MallApp.swift` 的 `init` 中同步调用（只读本地存储，无网络请求，不存在启动竞态）。
- 调用方在收到 `.unauthorized` 时负责调用 `AppSession.shared.logout()`（`AppSession` 本身不感知具体请求上下文）。

### 6. Tab 结构与错误反馈模式
- `ContentView` 四 Tab：首页(`house`) / 发布(`plus.circle`) / 消息(`bubble.left.and.bubble.right`) / 我的(`person`)，消息 Tab 用占位 `ChatListView`。
- 统一错误反馈模式：ViewModel 暴露 `errorMessage: String?`，View 用 SwiftUI 原生 `.alert(item:)` 或 `.alert(isPresented:)` 展示；这是后续所有 change 共同遵循的模式，替代小程序的 `wx.showToast`。

## 测试策略

项目当前只有单一 `MallApp` target（`project.yml` 未定义测试 target）。本 change 的任务范围包含新增测试基础设施：

1. 编辑 `project.yml` 新增 `MallAppTests` target（`type: bundle.unit-test`，`host: MallApp`）。
2. 运行 `xcodegen generate` 重建 `MallApp.xcodeproj`。
3. 自定义 `URLProtocol` 子类拦截网络请求，通过 `URLSessionConfiguration.protocolClasses` 按测试用例注入（不做全局 method swizzling），驱动 `APIClient` 的单元测试。

覆盖范围：
- `APIClient`：信封解码成功路径、`code != 0` 抛错、HTTP 401 → `.unauthorized`、`/user/save` 非信封响应解析、`EmptyResponse` 场景。
- `AppSession`：`login`/`register`/`logout` 状态流转、`bootstrap()` 从持久化存储恢复登录态（含无已保存状态场景）。
- 不测试纯 SwiftUI 视图布局。

## 风险与取舍

（与 OpenSpec design.md 一致，此处不重复展开，仅确认无新增风险）
- `/user/save` 无用户名唯一性校验，重复注册会触发后端 500 → iOS 侧统一归类为注册失败，展示通用提示。
- 不做静默重登录，可接受（24 小时 JWT 有效期下是低频场景）。
- 用户名密码账号体系与小程序微信登录账号体系不互通，是已知产品限制。

## Spec Patch

无——本轮技术设计确认未发现 `ios-network-client` / `ios-account-session` 的验收场景缺口，不需要回写 delta spec。
