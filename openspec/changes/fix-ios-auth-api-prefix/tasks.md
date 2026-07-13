# Tasks — fix-ios-auth-api-prefix

- [x] 1. mall-server：`router.go` 注册/登录路由改为 `/api/user/save`、`/api/user/login`（含 curl 注释），`go build && go test ./...` 通过
- [ ] 2. mall-ios：`Auth.swift` 请求路径改为 `/api/user/save`、`/api/user/login`（含注释），CLAUDE.md API 表同步更新
- [ ] 3. 本地 e2e 验证：起服务后 curl 新路径完成注册→登录链路，旧路径 `/user/save` 返回 404
