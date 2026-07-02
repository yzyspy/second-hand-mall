# Brainstorm Summary

- Change: mall-ios-network-auth-foundation
- Date: 2026-07-02

## 确认的技术方案

- `ApiResponse<T: Decodable>` 信封结构 `{code, msg, data: T?}`；无数据的接口统一使用新增的 `EmptyResponse`（空 `Decodable` 标记类型）作为 `T`，避免非可选类型解码歧义。
- `HTTPMethod` 一次性定义完整集合（`.get/.post/.put/.delete`），为后续 change #3（`PUT /api/product/update`）等预留，避免回头修改已完成的基础设施文件。
- `APIClient` 单例 `request<T: Decodable>(path:method:body:requiresAuth:)`：`requiresAuth == true` 时注入 `Authorization: Bearer <token>`；错误映射为 `.server(code:msg:)` / `.unauthorized` / `.transport(Error)` / `.decoding(Error)`；独立 `register(username:password:)` 方法处理 `/user/save` 的非信封 `{"message": String}` 响应。
- `TokenStore`：基于 Security 框架 `kSecClassGenericPassword`，不引入第三方依赖；`kSecAttrService = Bundle.main.bundleIdentifier`，`kSecAttrAccount = "jwt"`（固定字符串，单用户单 token 场景）。
- `AppSession`：`@Observable` 单例，暴露 `userId/username/avatar/isLoggedIn`；`login`/`register`/`logout`/`bootstrap`。`register` 成功后自动调用 `login` 完成"注册即登录"。`bootstrap()` 在 App 启动时从 `TokenStore` + `UserDefaults` 恢复登录态。
- 四 Tab 壳：`ContentView` 扩展为首页/发布/消息/我的；消息 Tab 使用占位 `ChatListView`/`ChatListViewModel`（真实实现留给 change #5）。
- Profile：未登录态提供登录/注册表单切换；已登录态展示个人信息卡片 + 退出登录（二次确认）。
- 错误反馈统一走 ViewModel 的 `errorMessage: String?` 驱动 SwiftUI 原生 `.alert()`，作为后续所有 change 的统一错误反馈模式。

## 关键取舍与风险

- `/user/save` 无用户名唯一性校验，重复注册会导致后端 500 → iOS 侧统一按"非 2xx 或缺少 message 字段"归类为注册失败，展示通用提示，不解析具体后端错误原因。
- 不做静默重登录（token 过期后需用户手动重新登录），可接受，因为 24 小时 JWT 有效期下是低频场景。
- 用户名密码账号体系与小程序微信登录账号体系不互通，是已知产品限制，不在本 change 解决范围。
- 项目当前只有单一 `MallApp` target（`project.yml` 无测试 target），本 change 的任务范围包含新增 `MallAppTests` target：编辑 `project.yml` + 运行 `xcodegen generate` 重建 `.xcodeproj`。

## 测试策略

- 新增 `MallAppTests` XCTest target（`project.yml` 改动 + `xcodegen generate`）。
- 自定义 `URLProtocol` 拦截网络请求进行 mock，通过 `URLSessionConfiguration.protocolClasses` 按测试用例注入，不做全局 swizzling。
- 覆盖 `APIClient`：信封解码成功路径、`code != 0` 抛错、HTTP 401 → `.unauthorized`、`/user/save` 非信封响应解析、`EmptyResponse` 场景。
- 覆盖 `AppSession`：`login`/`register`/`logout` 状态流转、`bootstrap()` 从持久化存储恢复登录态（含无已保存状态的场景）。
- 不测试纯 SwiftUI 视图布局（沿用 OpenSpec design.md 既定策略）。

## Spec Patch

无——本轮澄清未发现 `ios-network-client` / `ios-account-session` 的验收场景缺口，现有 delta spec 已覆盖上述行为决策。
