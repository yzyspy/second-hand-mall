# Delta — ios-network-client

## MODIFIED Requirements

### Requirement: 非信封响应的注册接口
`APIClient` SHALL 为 `/api/user/save` 提供单独的 `register` 方法，直接解码 `{"message": String}` 格式，不套用信封解析逻辑；任何非 2xx 响应或缺少 `message` 字段的响应均视为注册失败。

#### Scenario: 注册成功
- **WHEN** `POST /api/user/save` 返回 HTTP 2xx 且响应体包含 `message` 字段
- **THEN** `register` 返回该 `message` 字符串

#### Scenario: 注册失败（用户名重复导致后端 500）
- **WHEN** `POST /api/user/save` 返回非 2xx 状态码
- **THEN** `register` 抛出错误，调用方展示通用注册失败提示，不解析具体后端错误原因
