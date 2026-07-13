# fix-ios-auth-api-prefix

## Why

iOS App 部署指向线上环境后，注册（及登录）报错"数据格式不正确"。根因：线上 nginx 只把 `/api/*` 前缀代理到 Go 后端，根路径的 `POST /user/save`、`POST /user/login` 被 nginx 直接返回 404 HTML 页面；iOS 客户端对 HTML 做 JSON 解码抛出 `DecodingError`，其系统中文文案即"数据格式不正确"。已通过 curl 线上直接复现确认（`/api/product/search` 正常 200，`/user/save`、`/user/login`、`/admin/login` 均为 nginx 404）。

## What Changes

- mall-server：注册/登录路由从 `POST /user/save`、`POST /user/login` 改为 `POST /api/user/save`、`POST /api/user/login`（**BREAKING** 对旧路径调用方；经排查仅 iOS 使用这两个路径，且 iOS 包尚未发布，无兼容负担，不保留旧路径）
- mall-ios：`Auth.swift` 中 `AuthAPI.login` / `AuthAPI.register` 请求路径同步改为 `/api/user/login`、`/api/user/save`
- 文档：CLAUDE.md 中 API Endpoints 表同步更新

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ios-account-session`：登录/注册 Requirement 中引用的后端路径 `/user/login`、`/user/save` 改为 `/api/user/login`、`/api/user/save`
- `ios-network-client`：注册接口 Requirement 中引用的 `/user/save` 改为 `/api/user/save`

## Impact

- 受影响代码：`mall-server/internal/app/router/router.go`（两条路由）、`mall-ios/MallApp/Models/Auth.swift`（两个请求路径）
- 受影响系统：线上 nginx 无需改动（这正是本方案目的）；小程序使用 `/api/user/wx-login`，不受影响；mall-admin-web 使用 `/admin/*`，不在本次范围
- 部署：mall-server 需重新构建部署；iOS 包未发布，直接带上此修复
