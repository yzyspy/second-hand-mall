# mall-ios 基础架构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `mall-ios/` 目录下搭建 SwiftUI + MVVM 架构的 iOS 二手交易平台客户端骨架，包含三个 Tab 占位页。

**Architecture:** 使用 xcodegen 从 `project.yml` 生成 Xcode 工程，源文件按 MVVM 分层（Features/Feature/View + ViewModel），网络层以 `APIClient` 单例骨架预留，所有 ViewModel 使用 iOS 17 `@Observable` 宏。

**Tech Stack:** Swift 5.9、SwiftUI、iOS 17+、xcodegen（Homebrew）、xcodebuild

## Global Constraints

- iOS 最低版本：17.0
- Swift 版本：5.9
- UI 框架：SwiftUI（不使用 UIKit）
- ViewModel 响应式：`@Observable`（`import Observation`），不用 `ObservableObject`
- View 持有 ViewModel：`@State private var viewModel = XxxViewModel()`
- Bundle ID：`com.secondhandmall.MallApp`
- 后端地址：`http://localhost:8080`
- 工程目录：`mall-ios/`，与 `mall-server/`、`mall-mini/` 并列
- 当前阶段不添加单元测试

---

## File Map

| 文件 | 动作 | 职责 |
|------|------|------|
| `mall-ios/project.yml` | Create | xcodegen 工程描述文件 |
| `mall-ios/MallApp.swift` | Create | App 入口，`@main` |
| `mall-ios/ContentView.swift` | Create | TabView 根视图，组装三个 Tab |
| `mall-ios/Core/Network/APIClient.swift` | Create | 网络单例骨架 |
| `mall-ios/Features/Home/View/HomeView.swift` | Create | 首页占位 View |
| `mall-ios/Features/Home/ViewModel/HomeViewModel.swift` | Create | 首页 ViewModel |
| `mall-ios/Features/Publish/View/PublishView.swift` | Create | 发布占位 View |
| `mall-ios/Features/Publish/ViewModel/PublishViewModel.swift` | Create | 发布 ViewModel |
| `mall-ios/Features/Profile/View/ProfileView.swift` | Create | 我的占位 View |
| `mall-ios/Features/Profile/ViewModel/ProfileViewModel.swift` | Create | 我的 ViewModel |
| `mall-ios/Resources/Assets.xcassets/Contents.json` | Create | Asset Catalog 根 |
| `mall-ios/Resources/Assets.xcassets/AppIcon.appiconset/Contents.json` | Create | AppIcon 占位 |

---

## Task 1: 安装 xcodegen 并生成 Xcode 工程骨架

**Files:**
- Create: `mall-ios/project.yml`
- Create: `mall-ios/Resources/Assets.xcassets/Contents.json`
- Create: `mall-ios/Resources/Assets.xcassets/AppIcon.appiconset/Contents.json`
- Generated: `mall-ios/MallApp.xcodeproj/` （xcodegen 生成，不手动创建）

**Interfaces:**
- Produces: `MallApp.xcodeproj`，供后续所有任务的 `xcodebuild` 命令使用

- [ ] **Step 1: 安装 xcodegen**

```bash
brew install xcodegen
```

预期输出包含 `xcodegen` 版本信息。

- [ ] **Step 2: 创建 mall-ios 目录结构**

```bash
mkdir -p mall-ios/Core/Network
mkdir -p mall-ios/Features/Home/View
mkdir -p mall-ios/Features/Home/ViewModel
mkdir -p mall-ios/Features/Publish/View
mkdir -p mall-ios/Features/Publish/ViewModel
mkdir -p mall-ios/Features/Profile/View
mkdir -p mall-ios/Features/Profile/ViewModel
mkdir -p mall-ios/Resources/Assets.xcassets/AppIcon.appiconset
```

- [ ] **Step 3: 创建 project.yml**

写入 `mall-ios/project.yml`：

```yaml
name: MallApp
options:
  bundleIdPrefix: com.secondhandmall
  deploymentTarget:
    iOS: "17.0"
targets:
  MallApp:
    type: application
    platform: iOS
    deploymentTarget: "17.0"
    sources:
      - path: .
        excludes:
          - "project.yml"
          - "MallApp.xcodeproj/**"
          - "*.md"
    settings:
      base:
        SWIFT_VERSION: "5.9"
        PRODUCT_BUNDLE_IDENTIFIER: com.secondhandmall.MallApp
        DEVELOPMENT_TEAM: ""
        TARGETED_DEVICE_FAMILY: "1,2"
        GENERATE_INFOPLIST_FILE: YES
        MARKETING_VERSION: "1.0"
        CURRENT_PROJECT_VERSION: "1"
```

- [ ] **Step 4: 创建 Asset Catalog 最小内容**

写入 `mall-ios/Resources/Assets.xcassets/Contents.json`：

```json
{
  "info" : {
    "author" : "xcode",
    "version" : 1
  }
}
```

写入 `mall-ios/Resources/Assets.xcassets/AppIcon.appiconset/Contents.json`：

```json
{
  "images" : [],
  "info" : {
    "author" : "xcode",
    "version" : 1
  }
}
```

- [ ] **Step 5: 生成 Xcode 工程**

```bash
cd mall-ios && xcodegen generate
```

预期输出：
```
⚙️  Generating project MallApp
✅  Created project at mall-ios/MallApp.xcodeproj
```

- [ ] **Step 6: 提交**

```bash
git add mall-ios/project.yml mall-ios/Resources/
git commit -m "feat(ios): scaffold xcodegen project with asset catalog"
```

---

## Task 2: App 入口与 TabBar 根视图

**Files:**
- Create: `mall-ios/MallApp.swift`
- Create: `mall-ios/ContentView.swift`

**Interfaces:**
- Consumes: `HomeView`、`PublishView`、`ProfileView`（Task 4-6 创建，编译时依赖）
- Produces: `MallApp`（`@main`），`ContentView`（TabView 根）

- [ ] **Step 1: 创建 MallApp.swift**

写入 `mall-ios/MallApp.swift`：

```swift
import SwiftUI

@main
struct MallApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

- [ ] **Step 2: 创建 ContentView.swift**

写入 `mall-ios/ContentView.swift`：

```swift
import SwiftUI

struct ContentView: View {
    var body: some View {
        TabView {
            HomeView()
                .tabItem {
                    Label("首页", systemImage: "house")
                }
            PublishView()
                .tabItem {
                    Label("发布", systemImage: "plus.circle")
                }
            ProfileView()
                .tabItem {
                    Label("我的", systemImage: "person")
                }
        }
    }
}
```

- [ ] **Step 3: 提交**

```bash
git add mall-ios/MallApp.swift mall-ios/ContentView.swift
git commit -m "feat(ios): add app entry point and TabBar root view"
```

---

## Task 3: 网络层骨架

**Files:**
- Create: `mall-ios/Core/Network/APIClient.swift`

**Interfaces:**
- Produces: `APIClient.shared`（`@MainActor` 单例），各 ViewModel 后续通过此单例发请求

- [ ] **Step 1: 创建 APIClient.swift**

写入 `mall-ios/Core/Network/APIClient.swift`：

```swift
import Foundation

enum APIError: Error {
    case invalidResponse
    case serverError(Int)
}

@MainActor
final class APIClient {
    static let shared = APIClient()

    private let baseURL = "http://localhost:8080"

    private init() {}

    // 示例骨架方法——后续真实功能替换此处实现
    func get<T: Decodable>(_ path: String, as type: T.Type) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIError.invalidResponse
        }
        let (data, response) = try await URLSession.shared.data(from: url)
        guard let http = response as? HTTPURLResponse, http.statusCode == 200 else {
            let code = (response as? HTTPURLResponse)?.statusCode ?? -1
            print("[APIClient] error: HTTP \(code) for \(path)")
            throw APIError.serverError(code)
        }
        return try JSONDecoder().decode(type, from: data)
    }
}
```

- [ ] **Step 2: 重新生成工程（新增文件后需刷新）**

```bash
cd mall-ios && xcodegen generate
```

- [ ] **Step 3: 提交**

```bash
git add mall-ios/Core/Network/APIClient.swift
git commit -m "feat(ios): add APIClient network skeleton"
```

---

## Task 4: Home Feature（首页）

**Files:**
- Create: `mall-ios/Features/Home/ViewModel/HomeViewModel.swift`
- Create: `mall-ios/Features/Home/View/HomeView.swift`

**Interfaces:**
- Consumes: `Observation`（iOS 17 标准库）
- Produces: `HomeViewModel`（`@Observable`，`title: String`），`HomeView`（SwiftUI View）

- [ ] **Step 1: 创建 HomeViewModel.swift**

写入 `mall-ios/Features/Home/ViewModel/HomeViewModel.swift`：

```swift
import Observation

@Observable
final class HomeViewModel {
    var title = "首页"
}
```

- [ ] **Step 2: 创建 HomeView.swift**

写入 `mall-ios/Features/Home/View/HomeView.swift`：

```swift
import SwiftUI

struct HomeView: View {
    @State private var viewModel = HomeViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
```

- [ ] **Step 3: 重新生成工程**

```bash
cd mall-ios && xcodegen generate
```

- [ ] **Step 4: 提交**

```bash
git add mall-ios/Features/Home/
git commit -m "feat(ios): add Home feature placeholder"
```

---

## Task 5: Publish Feature（发布）

**Files:**
- Create: `mall-ios/Features/Publish/ViewModel/PublishViewModel.swift`
- Create: `mall-ios/Features/Publish/View/PublishView.swift`

**Interfaces:**
- Consumes: `Observation`（iOS 17 标准库）
- Produces: `PublishViewModel`（`@Observable`，`title: String`），`PublishView`（SwiftUI View）

- [ ] **Step 1: 创建 PublishViewModel.swift**

写入 `mall-ios/Features/Publish/ViewModel/PublishViewModel.swift`：

```swift
import Observation

@Observable
final class PublishViewModel {
    var title = "发布"
}
```

- [ ] **Step 2: 创建 PublishView.swift**

写入 `mall-ios/Features/Publish/View/PublishView.swift`：

```swift
import SwiftUI

struct PublishView: View {
    @State private var viewModel = PublishViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
```

- [ ] **Step 3: 重新生成工程**

```bash
cd mall-ios && xcodegen generate
```

- [ ] **Step 4: 提交**

```bash
git add mall-ios/Features/Publish/
git commit -m "feat(ios): add Publish feature placeholder"
```

---

## Task 6: Profile Feature（我的）+ 最终编译验证

**Files:**
- Create: `mall-ios/Features/Profile/ViewModel/ProfileViewModel.swift`
- Create: `mall-ios/Features/Profile/View/ProfileView.swift`

**Interfaces:**
- Consumes: `Observation`（iOS 17 标准库）
- Produces: `ProfileViewModel`（`@Observable`，`title: String`），`ProfileView`（SwiftUI View）

- [ ] **Step 1: 创建 ProfileViewModel.swift**

写入 `mall-ios/Features/Profile/ViewModel/ProfileViewModel.swift`：

```swift
import Observation

@Observable
final class ProfileViewModel {
    var title = "我的"
}
```

- [ ] **Step 2: 创建 ProfileView.swift**

写入 `mall-ios/Features/Profile/View/ProfileView.swift`：

```swift
import SwiftUI

struct ProfileView: View {
    @State private var viewModel = ProfileViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
```

- [ ] **Step 3: 重新生成工程**

```bash
cd mall-ios && xcodegen generate
```

- [ ] **Step 4: 编译验证（需安装 Xcode 命令行工具）**

```bash
cd mall-ios && xcodebuild \
  -project MallApp.xcodeproj \
  -scheme MallApp \
  -destination 'generic/platform=iOS Simulator' \
  build 2>&1 | grep -E '(BUILD SUCCEEDED|BUILD FAILED|error:)'
```

预期输出：
```
BUILD SUCCEEDED
```

- [ ] **Step 5: 提交**

```bash
git add mall-ios/Features/Profile/
git commit -m "feat(ios): add Profile feature placeholder, complete basic TabBar shell"
```
