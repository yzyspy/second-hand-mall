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
