# Delta — ios-account-session

## MODIFIED Requirements

### Requirement: 用户名密码登录
`AppSession` SHALL 提供 `login(username:password:)`，调用 `POST /api/user/login`，成功后把返回的 JWT 存入 `TokenStore`，并把 `userId`/`username`/`avatar` 持久化到 `UserDefaults`，同时更新已发布的会话属性。

#### Scenario: 登录成功
- **WHEN** 用户输入正确的用户名密码并调用 `login`
- **THEN** `isLoggedIn` 变为 `true`，token 保存到 Keychain，用户信息保存到 `UserDefaults`

#### Scenario: 登录失败
- **WHEN** 用户名或密码错误，`/api/user/login` 返回业务错误信封
- **THEN** `login` 向上抛出错误，`isLoggedIn` 保持 `false`，会话状态不变

### Requirement: 注册即登录
`AppSession` SHALL 提供 `register(username:password:)`，调用 `POST /api/user/save`；由于该接口不返回 token，注册成功后 SHALL 自动调用 `login` 完成登录态建立。

#### Scenario: 注册成功后自动登录
- **WHEN** 用户提交新用户名密码且 `/api/user/save` 返回成功
- **THEN** `register` 内部自动调用 `login`，完成后 `isLoggedIn` 变为 `true`

#### Scenario: 注册失败
- **WHEN** `/api/user/save` 返回非 2xx 或响应格式不符合预期
- **THEN** `register` 抛出错误，不触发登录，`isLoggedIn` 保持 `false`
