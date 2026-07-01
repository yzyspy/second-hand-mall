# mall-ios Phase 1 设计：网络层 / 认证 / 四 Tab 壳

**日期**: 2026-07-01
**目标**: 把 mall-ios 从占位骨架升级为具备真实网络请求、账号认证与四 Tab 结构的基础版本，为后续各功能阶段打底。

---

## 背景

mall-ios 当前是纯占位骨架（`Text(viewModel.title)`），`APIClient` 只有一个 `get` 方法。而 mall-mini 已经是功能完整的小程序，包含首页浏览、搜索、发布、编辑、收藏、消息（聊天）、个人中心共 10 个页面。整体功能对齐 mall-mini 的工作量较大，因此按阶段拆分：

1. **Phase 1（本设计）**: 网络层 + 认证 + 四 Tab 壳
2. Phase 2: 首页 / 搜索 / 商品详情（浏览为主）
3. Phase 3: 发布 / 编辑 / 我发布的
4. Phase 4: 我的收藏 / 个人中心完善
5. Phase 5: 消息（会话列表 + 聊天）

每个阶段独立走 设计 → 计划 → 实现 → review 流程。

### 认证方式的关键差异

mall-mini 通过 `wx.login()` 获取 code，再用 `/api/user/wx-login` 换取 token——这套机制在 iOS 上不存在（没有微信 SDK 依赖）。后端同时提供了未被小程序使用的用户名密码接口：`POST /user/save`（注册）与 `POST /user/login`（登录）。mall-ios 采用用户名密码方式，无需改动后端。

### `/user/save` 响应格式差异

除 `/user/save` 外，所有接口统一返回 `{code, msg, data}` 信封格式，`code == 0` 表示成功。但 `/user/save` 直接返回 `{"message": "save user success ..."}`，不遵循该信封，且后端没有用户名重复校验（`SaveUser` 内部对 DB 保存失败会 `panic`，由 gin 的 Recovery 中间件兜底为 500）。这是已知的后端限制，本阶段不修改后端，iOS 端按“非 2xx 或缺少预期字段即视为注册失败”处理，展示通用错误提示。

---

## 目录结构变更

```
mall-ios/
├── Core/
│   ├── Network/
│   │   ├── APIClient.swift        # 重写：通用 request 方法
│   │   ├── APIError.swift         # 新增
│   │   └── ApiResponse.swift      # 新增：{code, msg, data} 信封
│   └── Auth/
│       ├── TokenStore.swift       # 新增：Keychain 封装
│       └── AppSession.swift       # 新增：@Observable 会话状态
├── Features/
│   ├── Home/                      # 不变（仍为占位）
│   ├── Publish/                   # 不变（仍为占位）
│   ├── Chat/                      # 新增占位模块（消息 Tab）
│   │   ├── View/ChatListView.swift
│   │   └── ViewModel/ChatListViewModel.swift
│   └── Profile/                   # 改造为真实登录/注册/个人信息页
│       ├── View/ProfileView.swift
│       └── ViewModel/ProfileViewModel.swift
└── ContentView.swift               # 四 Tab
```

---

## 网络层设计（`Core/Network/`）

### `ApiResponse<T: Decodable>: Decodable`
```swift
struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String
    let data: T?
}
```

### `APIError`
```swift
enum APIError: Error {
    case server(code: Int, msg: String)   // code != 0
    case unauthorized                      // HTTP 401
    case transport(Error)                  // URLSession 失败
    case decoding(Error)                   // JSON 解析失败
}
```

### `APIClient`
- 单例，`baseURL` 沿用现状 `http://localhost:8080`（与小程序开发环境一致，暂不做多环境配置）。
- 核心方法：
  ```swift
  func request<T: Decodable>(
      _ path: String,
      method: HTTPMethod = .get,
      body: Encodable? = nil,
      requiresAuth: Bool = false
  ) async throws -> T
  ```
  - `requiresAuth == true` 时从 `TokenStore` 读取 token，注入 `Authorization: Bearer <token>`。
  - HTTP 401 → 抛出 `.unauthorized`（不做静默重试；用户名密码模式没有小程序那种静默重登录能力）。
  - 解码 `ApiResponse<T>`；`code != 0` → 抛出 `.server(code, msg)`；否则返回 `data`（`data` 为 nil 时对 `T == EmptyResponse` 之类的场景放行）。
- 单独提供 `register(username:password:) async throws -> String`，直接解码 `{"message": String}`，不复用信封解码逻辑。

---

## 会话与持久化（`Core/Auth/`）

### `TokenStore`
- 仅封装 Keychain 的 save / get / delete，操作一个字符串（JWT）。不引入第三方依赖。

### `AppSession`
```swift
@Observable
final class AppSession {
    static let shared = AppSession()

    private(set) var userId: Int?
    private(set) var username: String?
    private(set) var avatar: String?

    var isLoggedIn: Bool { userId != nil }

    func bootstrap()  // 启动时从 TokenStore + UserDefaults 恢复状态
    func login(username: String, password: String) async throws
    func register(username: String, password: String) async throws
    func logout()
}
```
- `login` 调用 `/user/login`，成功后把 token 存入 `TokenStore`，`user_id`/`user_name`/`avatar` 存入 `UserDefaults`，更新已发布属性。
- `register` 调用 `/user/save`；由于该接口不返回 token，注册成功后自动调用 `login` 完成登录态建立（对用户表现为“注册即登录”）。
- `logout` 清空 `TokenStore` 与 `UserDefaults` 中的对应字段，并重置发布属性。
- 任意请求收到 `.unauthorized` 时，调用方触发 `AppSession.shared.logout()`，UI 因 `isLoggedIn` 变化自动回退到登录表单。

---

## 我的（Profile）Tab

### 未登录态
- 用户名 / 密码输入框 + 登录按钮。
- “没有账号？去注册”切换链接，切换后显示用户名 / 密码 / 确认密码 + 注册按钮。
- 前端校验：用户名非空、密码非空、注册时两次密码一致；不做用户名格式/密码强度校验（YAGNI，后端本身也未做）。

### 已登录态
- 个人信息卡片：占位头像、用户名（后端当前不支持注册时自定义昵称/头像，属已知限制，不在本阶段解决）。
- 退出登录按钮，点击后二次确认（对齐 `my.ts` 的 `wx.showModal` 行为），确认后调用 `AppSession.logout()`。
- **不包含** `my.ts` 中的“我发布的 / 我卖出的 / 我的收藏 / 收货地址”等菜单项——这些功能对应的页面在后续阶段才会存在，本阶段加入会造成死点击，等对应阶段落地后再加回对应入口。

### 错误反馈
- 登录失败、注册失败、网络错误统一通过 ViewModel 的 `errorMessage: String?` 驱动 SwiftUI 原生 `.alert()` 展示。这是后续所有阶段统一遵循的错误反馈模式（替代小程序的 `wx.showToast`）。

---

## Tab 栏

`ContentView` 从三 Tab 扩展为四 Tab，顺序与图标对齐 mall-mini 的 `app.json`：

| 顺序 | Tab | 图标 (SF Symbol) | 本阶段状态 |
|---|---|---|---|
| 1 | 首页 | `house` | 占位（不变） |
| 2 | 发布 | `plus.circle` | 占位（不变） |
| 3 | 消息 | `bubble.left.and.bubble.right` | 新增占位 `ChatListView` |
| 4 | 我的 | `person` | **本阶段实现的真实页面** |

消息未读角标（对应小程序 `app.ts` 里 `wx.setTabBarBadge`）延后到 Phase 5 聊天功能一起实现。

---

## 测试策略

新增 `MallAppTests` XCTest target（写入 `project.yml`）。使用自定义 `URLProtocol` 拦截网络请求进行 mock，覆盖：

- `APIClient`：信封解码成功路径、`code != 0` 抛错路径、HTTP 401 → `.unauthorized` 映射、`/user/save` 非信封响应解析。
- `AppSession`：`login` / `register` / `logout` 状态流转，以及重启后从 `TokenStore` + `UserDefaults` 恢复登录态（`bootstrap()`）。

不测试纯 SwiftUI 视图布局。

---

## 本阶段不做的事

- 不实现 Home / Publish / Chat 的真实业务逻辑（仍为占位）。
- 不给 `/user/save` 增加用户名唯一性校验或修改响应格式（后端改动超出 mall-ios 范围）。
- 不支持 Sign in with Apple 或其他认证方式。
- 不做多环境（dev/prod）baseURL 切换配置。
- 不实现消息未读角标。
