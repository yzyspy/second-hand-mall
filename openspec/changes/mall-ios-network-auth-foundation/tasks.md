## 1. 网络层基础

- [x] 1.1 新增 `Core/Network/ApiResponse.swift`：`ApiResponse<T: Decodable>` 信封结构
- [x] 1.2 重写 `Core/Network/APIError.swift`：`.server(code:msg:)` / `.unauthorized` / `.transport(Error)` / `.decoding(Error)`
- [x] 1.3 重写 `Core/Network/APIClient.swift`：`request<T: Decodable>(path:method:body:requiresAuth:)`，处理鉴权注入、信封解析、错误映射
- [x] 1.4 在 `APIClient` 中新增独立的 `register(username:password:) async throws -> String` 方法（非信封解析）

## 2. 账号会话

- [x] 2.1 新增 `Core/Auth/TokenStore.swift`：Keychain save/get/delete 封装
- [x] 2.2 新增 `Core/Auth/AppSession.swift`：`@Observable` 会话状态，实现 `login`/`register`/`logout`/`bootstrap`
- [x] 2.3 在 `MallApp.swift` 启动时调用 `AppSession.shared.bootstrap()`

## 3. 四 Tab 壳与消息占位

- [ ] 3.1 新增 `Features/Chat/View/ChatListView.swift` 与 `Features/Chat/ViewModel/ChatListViewModel.swift`（占位）
- [ ] 3.2 修改 `ContentView.swift`：扩展为四 Tab（首页/发布/消息/我的），图标对齐 mall-mini `app.json`

## 4. 我的（Profile）真实登录/注册

- [ ] 4.1 重写 `Features/Profile/ViewModel/ProfileViewModel.swift`：未登录态表单状态、`errorMessage` 驱动的错误反馈
- [ ] 4.2 重写 `Features/Profile/View/ProfileView.swift`：未登录态登录/注册表单切换 UI
- [ ] 4.3 实现已登录态：个人信息卡片（占位头像 + 用户名）+ 退出登录按钮（含二次确认）

## 5. 测试

- [x] 5.1 新增 `MallAppTests` XCTest target（写入 `project.yml`）
- [ ] 5.2 编写自定义 `URLProtocol` 用于拦截网络请求 mock
- [ ] 5.3 测试 `APIClient`：信封解码成功路径、`code != 0` 抛错、HTTP 401 → `.unauthorized`、`/user/save` 非信封响应解析
- [ ] 5.4 测试 `AppSession`：`login`/`register`/`logout` 状态流转，以及 `bootstrap()` 从持久化存储恢复登录态

## 6. 验证

- [ ] 6.1 `xcodebuild build` 或等效命令通过
- [ ] 6.2 运行 `MallAppTests` 全部通过
- [ ] 6.3 手动验证：注册新用户 → 自动登录 → 杀掉 App 重启 → 登录态保留 → 退出登录 → 回到登录表单
