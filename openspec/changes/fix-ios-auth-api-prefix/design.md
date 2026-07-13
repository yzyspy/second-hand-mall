# Design — fix-ios-auth-api-prefix

## 修复方案（hotfix，单方案）

线上 nginx 只代理 `/api/*` 到 Go 后端，且 nginx 配置不在本仓库、也不希望为此增加代理规则。因此把注册/登录接口统一收敛到 `/api` 前缀下，与小程序 `/api/user/wx-login` 及其余业务接口保持同一约定。

### 服务端（mall-server/internal/app/router/router.go）

- `r.POST("/user/save", ...)` → `r.POST("/api/user/save", ...)`
- `r.POST("/user/login", ...)` → `r.POST("/api/user/login", ...)`
- 同步更新路由上方的 curl 注释
- 不保留旧路径别名：全仓库排查确认仅 mall-ios 调用这两个路径，且 iOS 包尚未发布

### iOS（mall-ios/MallApp/Models/Auth.swift）

- `AuthAPI.login`：`API.post("/user/login", ...)` → `API.post("/api/user/login", ...)`
- `AuthAPI.register`：`API.post("/user/save", ...)` → `API.post("/api/user/save", ...)`
- 同步更新相关注释

### 文档

- CLAUDE.md 的 API Endpoints 表：`POST /user/save`、`POST /user/login` 改为带 `/api` 前缀

## 验证方式

- `cd mall-server && go build && go test ./...`
- 本地起服务，curl `POST /api/user/save`、`POST /api/user/login` 走通注册→登录全链路；确认旧路径 `/user/save` 返回 404
- iOS 侧 xcodebuild 编译通过（路径为字符串常量，行为由服务端 e2e 覆盖）
