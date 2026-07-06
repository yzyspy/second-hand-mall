# ios-network-client Specification

## Purpose
TBD - created by archiving change mall-ios-network-auth-foundation. Update Purpose after archive.
## Requirements
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

