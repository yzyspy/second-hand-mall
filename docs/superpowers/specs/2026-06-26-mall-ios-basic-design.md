# mall-ios 基础架构设计

**日期**: 2026-06-26  
**目标**: 为二手交易平台创建 iOS 用户端，搭好 MVVM 分层结构与 TabBar 壳

---

## 概述

在现有项目（Go 后端 + 微信小程序）基础上，新增 `mall-ios/` 目录作为 iOS 原生客户端。技术栈：SwiftUI + iOS 17+，采用 MVVM 架构，三个 Tab 对应小程序的首页 / 发布 / 我的。

---

## 目录结构

```
mall-ios/
├── MallApp.swift
├── ContentView.swift
├── Core/
│   └── Network/
│       └── APIClient.swift
├── Features/
│   ├── Home/
│   │   ├── View/HomeView.swift
│   │   └── ViewModel/HomeViewModel.swift
│   ├── Publish/
│   │   ├── View/PublishView.swift
│   │   └── ViewModel/PublishViewModel.swift
│   └── Profile/
│       ├── View/ProfileView.swift
│       └── ViewModel/ProfileViewModel.swift
└── Resources/
    └── Assets.xcassets
```

---

## 架构设计

### MVVM 模式
- **ViewModel**: 使用 `@Observable`（iOS 17 宏），持有页面状态
- **View**: 通过 `@State` 持有 ViewModel 实例，读取状态渲染 UI
- **数据流**: View → ViewModel → APIClient（单向）

### TabBar
- 使用原生 `TabView`，三个 Tab：
  - 首页（`house` SF Symbol）
  - 发布（`plus.circle` SF Symbol）
  - 我的（`person` SF Symbol）

### 网络层骨架
- `APIClient` 为单例，方法签名使用 `async throws`
- 当前返回 mock 数据，后续对接 `mall-server`（`localhost:8080`）时只改此层
- 错误处理：`throw` 时打印日志，不 crash

---

## 占位页内容

各页面居中显示页面名称文字，无其他业务逻辑：

| Tab | View | 占位内容 |
|-----|------|---------|
| 首页 | HomeView | Text("首页") |
| 发布 | PublishView | Text("发布") |
| 我的 | ProfileView | Text("我的") |

---

## 测试策略

- 当前阶段不添加单元测试（占位页无逻辑）
- 首个真实功能完成后补充对应 ViewModel 测试

---

## 与现有项目的关系

- `mall-ios/` 与 `mall-server/`、`mall-mini/`、`mall-admin-web/` 并列
- 后端 API 地址：`localhost:8080`（开发），与小程序保持一致
- 认证方式：JWT，`Authorization: Bearer <token>` header，后续在 `APIClient` 统一注入
