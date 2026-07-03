---
change: mall-ios-network-auth-foundation
design-doc: docs/superpowers/specs/2026-07-02-mall-ios-network-auth-foundation-design.md
base-ref: 4c3249c052c8974429be9ffd66ed5e1ab858a4aa
---

# mall-ios-network-auth-foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `mall-ios` 从占位骨架（三个 Tab 各显示 `Text(title)`，`APIClient` 只有 mock `get` 方法）升级为具备真实网络层、用户名密码账号认证与四 Tab 结构的基础版本，并补齐项目此前完全缺失的单元测试基础设施。

**Architecture:** 新增 `Core/Network/`（`ApiResponse`/`APIError`/`HTTPMethod`/`APIClient`）承载统一信封解析、鉴权注入、错误分类；新增 `Core/Auth/`（`TokenStore` 封装 Keychain、`AppSession` 作为 `@Observable` 单一会话真源）；`Features/Chat/` 新增消息 Tab 占位；`Features/Profile/` 从纯占位改造为登录/注册表单 + 已登录态卡片；`ContentView` 扩展为四 Tab。同时编辑 `project.yml` 新增 `MallAppTests` target 并用自定义 `URLProtocol` 拦截网络请求驱动单元测试。

**Tech Stack:** Swift 5.9, SwiftUI, iOS 17 部署目标, `Observation` 框架（`@Observable`，非 `ObservableObject`），`Security` 框架（Keychain），`URLSession` async/await，XCTest，XcodeGen 2.45（`project.yml` → `xcodegen generate` → `.xcodeproj`，绝不手工编辑 `.xcodeproj`）。

## Global Constraints

- iOS 部署目标固定为 `17.0`（`project.yml` 已设置，新 target 需保持一致）。
- `SWIFT_VERSION: "5.9"`（沿用现有 `MallApp` target 设置）。
- 所有响应式状态类型使用 `@Observable`（`import Observation`），不得引入 `ObservableObject`/`@Published`（与仓库现有 `HomeViewModel`/`PublishViewModel` 一致）。
- `APIClient` 基础网络层文件（`ApiResponse.swift`/`APIError.swift`/`HTTPMethod.swift`/`APIClient.swift`）一次性把 `HTTPMethod` 定义完整（`GET`/`POST`/`PUT`/`DELETE`），即使本 change 只用到 `.get`/`.post`，避免后续 change 再回头修改已测试过的基础设施文件。
- `TokenStore` 使用 `kSecClassGenericPassword`，`kSecAttrService = Bundle.main.bundleIdentifier`，`kSecAttrAccount = "jwt"`（固定字符串常量，不按用户名区分）。
- 鉴权头格式固定为 `Authorization: Bearer <token>`（与后端 `mall-server/internal/app/router/auth.go` 的 `AuthMiddleware` 解析格式一致）。
- 后端信封格式固定为 `{"code": Int, "msg": String, "data": T?}`，`code == 0` 视为成功；`POST /user/save`（注册）是例外，直接返回 `{"message": String}`，不套用信封。
- `/user/login` 成功响应的 `data` 字段：`{"user_id": Int, "user_name": String, "avatar": String, "token": String}`（字段名与后端 `mall-server/internal/app/service/login.go:74-83` 完全一致，注意 `user_id`/`user_name` 是下划线命名，需要 `CodingKeys` 映射）。
- 严禁手工编辑 `mall-ios/MallApp.xcodeproj/**` 下的任何文件；对 target/scheme 的改动一律先改 `project.yml` 再运行 `xcodegen generate`。
- 不测试纯 SwiftUI 视图布局（`ProfileView`/`ChatListView`/`ContentView` 等 View 文件不写 XCTest）；测试聚焦 `APIClient`、`TokenStore`、`AppSession` 的逻辑路径。
- 不修改 `mall-server/` 或 `mall-mini/` 下任何文件——它们只是本计划的行为参考。
- 提交信息使用英文祈使句（沿用仓库现有提交历史风格，如 `feat(ios): ...` / `fix(ios): ...` / `test(ios): ...`）。

---

## Task 1: 新增 `MallAppTests` 测试基础设施

**Files:**
- Modify: `mall-ios/project.yml`
- Create: `mall-ios/MallAppTests/SmokeTests.swift`

**Interfaces:**
- Produces: 一个可运行的 `MallAppTests` XCTest target，`@testable import MallApp` 可用，供后续所有任务追加测试文件。

- [x] **Step 1: 编辑 `project.yml` 新增 `MallAppTests` target**

在 `mall-ios/project.yml` 的 `targets:` 下、`MallApp:` 定义之后追加 `MallAppTests` target，并新增顶层 `schemes:` 块让同一个 scheme 同时构建 App 和 Tests：

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
          - "MallAppTests/**"
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
  MallAppTests:
    type: bundle.unit-test
    platform: iOS
    deploymentTarget: "17.0"
    sources:
      - path: MallAppTests
    dependencies:
      - target: MallApp
    settings:
      base:
        SWIFT_VERSION: "5.9"
        PRODUCT_BUNDLE_IDENTIFIER: com.secondhandmall.MallAppTests
        GENERATE_INFOPLIST_FILE: YES
schemes:
  MallApp:
    build:
      targets:
        MallApp: all
        MallAppTests: [test]
    test:
      targets:
        - MallAppTests
```

注意：`MallApp` target 的 `sources.excludes` 新增了 `"MallAppTests/**"`，防止 App target 把测试代码也编译进去（此前 `path: .` 会递归包含所有子目录）。

**验证要点（xcodegen 生成规则，已在隔离沙箱中验证过）：** `MallAppTests` 只需声明 `dependencies: [{target: MallApp}]`，XcodeGen 会自动为其生成 `TEST_HOST = "$(BUILT_PRODUCTS_DIR)/MallApp.app/MallApp"` 与 `BUNDLE_LOADER = "$(TEST_HOST)"`，不需要额外的 `host:` 字段（XcodeGen 的 Target Spec 里没有这个字段名）。顶层 `schemes:` 一旦手写，会替代自动生成的 per-target scheme，因此显式把 `MallAppTests` 纳入同一个 `MallApp` scheme 的 build/test 阶段。

- [x] **Step 2: 创建占位测试文件**

创建 `mall-ios/MallAppTests/SmokeTests.swift`：

```swift
import XCTest
@testable import MallApp

final class SmokeTests: XCTestCase {
    func testTrue() {
        XCTAssertTrue(true)
    }
}
```

- [x] **Step 3: 运行 `xcodegen generate` 重建 `.xcodeproj`**

```bash
cd mall-ios && xcodegen generate
```

预期输出包含 `Created project at .../MallApp.xcodeproj`（或 `Updated project`），无报错。

- [x] **Step 4: 构建并运行测试验证管线打通**

```bash
cd mall-ios && xcrun simctl list devices available | grep iPhone
```

从输出中选一台可用的 iPhone 模拟器（例如 `iPhone 17`），然后运行：

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test
```

预期看到 `Test Suite 'SmokeTests'` 通过，最终输出 `** TEST SUCCEEDED **`。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add project.yml MallAppTests/SmokeTests.swift MallApp.xcodeproj
git commit -m "test(ios): add MallAppTests target via xcodegen"
```

---

## Task 2: 自定义 `URLProtocol` 网络 Mock

**Files:**
- Create: `mall-ios/MallAppTests/Support/MockURLProtocol.swift`
- Create: `mall-ios/MallAppTests/Support/MockURLProtocolTests.swift`

**Interfaces:**
- Produces: `MockURLProtocol`，静态属性 `requestHandler: ((URLRequest) throws -> (HTTPURLResponse, Data))?`，供后续 `APIClientTests`/`AppSessionTests` 通过 `URLSessionConfiguration.protocolClasses = [MockURLProtocol.self]` 注入。

- [x] **Step 1: 创建 `MockURLProtocol`**

```swift
import Foundation

/// 自定义 URLProtocol，用于在单元测试中拦截 URLSession 请求并返回预设响应。
/// 通过 URLSessionConfiguration.protocolClasses 按测试用例注入，不做全局 method swizzling。
final class MockURLProtocol: URLProtocol {
    static var requestHandler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool {
        true
    }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    override func startLoading() {
        guard let handler = MockURLProtocol.requestHandler else {
            client?.urlProtocol(self, didFailWithError: URLError(.unknown))
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
```

- [x] **Step 2: 编写自测确认拦截生效**

```swift
import XCTest
@testable import MallApp

final class MockURLProtocolTests: XCTestCase {
    override func tearDown() {
        MockURLProtocol.requestHandler = nil
        super.tearDown()
    }

    func testRequestHandler_isInvokedAndReturnsStubbedResponse() async throws {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: config)

        MockURLProtocol.requestHandler = { request in
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, "hello".data(using: .utf8)!)
        }

        let (data, response) = try await session.data(from: URL(string: "https://mall.test/ping")!)

        XCTAssertEqual(String(data: data, encoding: .utf8), "hello")
        XCTAssertEqual((response as? HTTPURLResponse)?.statusCode, 200)
    }
}
```

- [x] **Step 3: 运行测试验证通过**

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/MockURLProtocolTests
```

预期：`Test Suite 'MockURLProtocolTests' passed`。

- [x] **Step 4: Commit**

```bash
cd mall-ios && git add MallAppTests/Support/MockURLProtocol.swift MallAppTests/Support/MockURLProtocolTests.swift
git commit -m "test(ios): add MockURLProtocol harness for network stubbing"
```

---

## Task 3: `TokenStore` Keychain 封装

**Files:**
- Create: `mall-ios/Core/Auth/TokenStore.swift`
- Create: `mall-ios/MallAppTests/Core/Auth/TokenStoreTests.swift`

**Interfaces:**
- Produces:
  ```swift
  final class TokenStore {
      static let shared: TokenStore
      func save(_ token: String)
      func getToken() -> String?
      func delete()
  }
  ```
- Consumed by: Task 5（`APIClient.request` 的 `requiresAuth` 分支）、Task 7（`AppSession`）。

- [x] **Step 1: 写失败的测试**

创建 `mall-ios/MallAppTests/Core/Auth/TokenStoreTests.swift`：

```swift
import XCTest
@testable import MallApp

final class TokenStoreTests: XCTestCase {
    override func tearDown() {
        TokenStore.shared.delete()
        super.tearDown()
    }

    func testSaveAndGetToken_returnsSavedValue() {
        TokenStore.shared.save("abc123")
        XCTAssertEqual(TokenStore.shared.getToken(), "abc123")
    }

    func testGetToken_returnsNilWhenNotSaved() {
        TokenStore.shared.delete()
        XCTAssertNil(TokenStore.shared.getToken())
    }

    func testSaveOverwritesPreviousToken() {
        TokenStore.shared.save("first")
        TokenStore.shared.save("second")
        XCTAssertEqual(TokenStore.shared.getToken(), "second")
    }

    func testDelete_removesToken() {
        TokenStore.shared.save("abc123")
        TokenStore.shared.delete()
        XCTAssertNil(TokenStore.shared.getToken())
    }
}
```

- [x] **Step 2: 运行测试确认编译失败（`TokenStore` 尚不存在）**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/TokenStoreTests
```

预期：编译失败，报 `cannot find 'TokenStore' in scope`。

- [x] **Step 3: 实现 `TokenStore`**

创建 `mall-ios/Core/Auth/TokenStore.swift`：

```swift
import Foundation
import Security

/// Keychain 封装，仅负责单一 JWT token 的存取。
/// service = Bundle.main.bundleIdentifier，account 固定为 "jwt"（单用户单 token 场景）。
final class TokenStore {
    static let shared = TokenStore()

    private let service: String
    private let account = "jwt"

    init(service: String = Bundle.main.bundleIdentifier ?? "com.secondhandmall.MallApp") {
        self.service = service
    }

    func save(_ token: String) {
        delete()
        guard let data = token.data(using: .utf8) else { return }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecValueData as String: data
        ]
        SecItemAdd(query as CFDictionary, nil)
    }

    func getToken() -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    func delete() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account
        ]
        SecItemDelete(query as CFDictionary)
    }
}
```

- [x] **Step 4: 运行测试确认通过**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/TokenStoreTests
```

预期：`Test Suite 'TokenStoreTests' passed`，4 个测试全部通过。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add Core/Auth/TokenStore.swift MallAppTests/Core/Auth/TokenStoreTests.swift MallApp.xcodeproj
git commit -m "feat(ios): add TokenStore Keychain wrapper"
```

---

## Task 4: 网络基础类型 —— `ApiResponse` / `EmptyResponse` / `HTTPMethod` / `APIError`

**Files:**
- Create: `mall-ios/Core/Network/ApiResponse.swift`
- Create: `mall-ios/Core/Network/HTTPMethod.swift`
- Create: `mall-ios/Core/Network/APIError.swift`
- Modify: `mall-ios/Core/Network/APIClient.swift`（先删除旧的内联 `enum APIError` 定义，`APIClient` 类主体留到 Task 5 重写）
- Create: `mall-ios/MallAppTests/Core/Network/ApiResponseTests.swift`

**Interfaces:**
- Produces:
  ```swift
  struct ApiResponse<T: Decodable>: Decodable {
      let code: Int
      let msg: String
      let data: T?
  }
  struct EmptyResponse: Decodable {}
  enum HTTPMethod: String {
      case get = "GET"
      case post = "POST"
      case put = "PUT"
      case delete = "DELETE"
  }
  enum APIError: Error {
      case server(code: Int, msg: String)
      case unauthorized
      case transport(Error)
      case decoding(Error)
  }
  ```
- Consumed by: Task 5（`APIClient.request`/`register`）。

- [x] **Step 1: 写失败的测试**

创建 `mall-ios/MallAppTests/Core/Network/ApiResponseTests.swift`：

```swift
import XCTest
@testable import MallApp

final class ApiResponseTests: XCTestCase {
    private struct Payload: Decodable, Equatable {
        let message: String
    }

    func testDecode_successEnvelopeWithData() throws {
        let json = """
        {"code":0,"msg":"ok","data":{"message":"pong"}}
        """.data(using: .utf8)!

        let decoded = try JSONDecoder().decode(ApiResponse<Payload>.self, from: json)

        XCTAssertEqual(decoded.code, 0)
        XCTAssertEqual(decoded.msg, "ok")
        XCTAssertEqual(decoded.data, Payload(message: "pong"))
    }

    func testDecode_nullDataBecomesNil() throws {
        let json = """
        {"code":0,"msg":"ok","data":null}
        """.data(using: .utf8)!

        let decoded = try JSONDecoder().decode(ApiResponse<Payload>.self, from: json)

        XCTAssertNil(decoded.data)
    }

    func testHTTPMethod_rawValuesMatchHTTPVerbs() {
        XCTAssertEqual(HTTPMethod.get.rawValue, "GET")
        XCTAssertEqual(HTTPMethod.post.rawValue, "POST")
        XCTAssertEqual(HTTPMethod.put.rawValue, "PUT")
        XCTAssertEqual(HTTPMethod.delete.rawValue, "DELETE")
    }
}
```

- [x] **Step 2: 运行测试确认编译失败**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/ApiResponseTests
```

预期：编译失败，报 `cannot find type 'ApiResponse' in scope`（及 `HTTPMethod`）。

- [x] **Step 3: 创建 `ApiResponse.swift`**

```swift
import Foundation

/// mall-server 统一响应信封：{code, msg, data}。
/// code == 0 表示成功；data 为 nil 表示该接口本次响应无业务数据（见 EmptyResponse 用法）。
struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String
    let data: T?
}

/// 标记类型：用于只需要 code == 0 语义成功、data 恒为 null 的接口。
/// 让调用方以 `T = EmptyResponse` 显式声明"无数据"，而不必把 T 声明为 Optional 多一层解包。
struct EmptyResponse: Decodable {}
```

- [x] **Step 4: 创建 `HTTPMethod.swift`**

```swift
/// HTTP 方法集合。本 change 只用到 .get/.post；.put/.delete 一并定义，
/// 为后续 change（如 PUT /api/product/update）预留，避免回头修改本文件。
enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}
```

- [x] **Step 5: 创建 `APIError.swift`**

```swift
import Foundation

/// APIClient 统一错误分类。
enum APIError: Error {
    /// 信封解析成功但 code != 0：服务端业务错误。
    case server(code: Int, msg: String)
    /// HTTP 401，或 requiresAuth == true 但本地无已保存 token。
    case unauthorized
    /// URLSession 层失败（无网络连接等）。
    case transport(Error)
    /// JSON 解码失败，或响应体结构与期望类型不匹配。
    case decoding(Error)
}
```

- [x] **Step 6: 从 `APIClient.swift` 移除旧的内联 `APIError` 定义与占位 `get` 方法**

把 `mall-ios/Core/Network/APIClient.swift` 整个文件内容替换为最小占位（`APIClient` 主体在 Task 5 重写，这一步只是清掉与新 `APIError.swift` 冲突的重复类型定义）：

```swift
import Foundation

final class APIClient {
    static let shared = APIClient()
}
```

- [x] **Step 7: 运行测试确认通过**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/ApiResponseTests
```

预期：`Test Suite 'ApiResponseTests' passed`，3 个测试全部通过。同时确认整个项目仍能构建（`APIClient.shared` 目前是最小占位，`HomeView`/`PublishView`/`ProfileView` 均未引用旧的 `get` 方法，不会破坏编译）：

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' build
```

预期：`** BUILD SUCCEEDED **`。

- [x] **Step 8: Commit**

```bash
cd mall-ios && git add Core/Network/ApiResponse.swift Core/Network/HTTPMethod.swift Core/Network/APIError.swift Core/Network/APIClient.swift MallAppTests/Core/Network/ApiResponseTests.swift MallApp.xcodeproj
git commit -m "feat(ios): add ApiResponse envelope, HTTPMethod, and APIError types"
```

---

## Task 5: `APIClient.request<T>` 通用请求方法

**Files:**
- Modify: `mall-ios/Core/Network/APIClient.swift`
- Create: `mall-ios/MallAppTests/Core/Network/APIClientTests.swift`

**Interfaces:**
- Consumes: `TokenStore.shared.getToken()`（Task 3），`ApiResponse<T>`/`EmptyResponse`/`HTTPMethod`/`APIError`（Task 4）。
- Produces:
  ```swift
  final class APIClient {
      static let shared: APIClient
      init(session: URLSession = .shared, baseURL: String = "http://localhost:8080")
      func request<T: Decodable>(
          _ path: String,
          method: HTTPMethod = .get,
          body: Encodable? = nil,
          requiresAuth: Bool = false
      ) async throws -> T
  }
  ```
  （`init` 非 private，允许测试注入自定义 `URLSession`/`baseURL`；`register` 方法在 Task 6 追加。）
- Consumed by: Task 6（`register` 复用同一个类）、Task 7（`AppSession`）。

- [x] **Step 1: 写失败的测试**

创建 `mall-ios/MallAppTests/Core/Network/APIClientTests.swift`：

```swift
import XCTest
@testable import MallApp

final class APIClientTests: XCTestCase {
    private var apiClient: APIClient!

    override func setUp() {
        super.setUp()
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        apiClient = APIClient(session: URLSession(configuration: config), baseURL: "https://mall.test")
    }

    override func tearDown() {
        MockURLProtocol.requestHandler = nil
        TokenStore.shared.delete()
        super.tearDown()
    }

    private struct Ping: Decodable, Equatable {
        let message: String
    }

    func testRequest_decodesSuccessEnvelope() async throws {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":0,"msg":"ok","data":{"message":"pong"}}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        let result: Ping = try await apiClient.request("/ping")

        XCTAssertEqual(result, Ping(message: "pong"))
    }

    func testRequest_throwsServerErrorForNonZeroCode() async {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":1001,"msg":"参数错误","data":null}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        do {
            let _: EmptyResponse = try await apiClient.request("/x")
            XCTFail("expected throw")
        } catch APIError.server(let code, let msg) {
            XCTAssertEqual(code, 1001)
            XCTAssertEqual(msg, "参数错误")
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }

    func testRequest_throwsUnauthorizedFor401Response() async {
        MockURLProtocol.requestHandler = { request in
            let response = HTTPURLResponse(url: request.url!, statusCode: 401, httpVersion: nil, headerFields: nil)!
            return (response, Data())
        }

        do {
            let _: EmptyResponse = try await apiClient.request("/x")
            XCTFail("expected throw")
        } catch APIError.unauthorized {
            // expected
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }

    func testRequest_throwsDecodingErrorForMalformedBody() async {
        MockURLProtocol.requestHandler = { request in
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, "not json".data(using: .utf8)!)
        }

        do {
            let _: EmptyResponse = try await apiClient.request("/x")
            XCTFail("expected throw")
        } catch APIError.decoding {
            // expected
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }

    func testRequest_decodesEmptyResponseWhenDataIsNull() async throws {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":0,"msg":"ok","data":null}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        let _: EmptyResponse = try await apiClient.request("/x")
    }

    func testRequest_injectsAuthorizationHeaderWhenTokenPresent() async throws {
        TokenStore.shared.save("secret-token")
        var capturedAuthHeader: String?
        MockURLProtocol.requestHandler = { request in
            capturedAuthHeader = request.value(forHTTPHeaderField: "Authorization")
            let json = """
            {"code":0,"msg":"ok","data":null}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        let _: EmptyResponse = try await apiClient.request("/ping", requiresAuth: true)

        XCTAssertEqual(capturedAuthHeader, "Bearer secret-token")
    }

    func testRequest_doesNotInjectAuthorizationHeaderWhenNotRequired() async throws {
        var capturedAuthHeader: String? = "unset"
        MockURLProtocol.requestHandler = { request in
            capturedAuthHeader = request.value(forHTTPHeaderField: "Authorization")
            let json = """
            {"code":0,"msg":"ok","data":null}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        let _: EmptyResponse = try await apiClient.request("/x", requiresAuth: false)

        XCTAssertNil(capturedAuthHeader)
    }

    func testRequest_throwsUnauthorizedWhenAuthRequiredAndNoTokenSaved() async {
        TokenStore.shared.delete()
        var handlerCalled = false
        MockURLProtocol.requestHandler = { request in
            handlerCalled = true
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, Data())
        }

        do {
            let _: EmptyResponse = try await apiClient.request("/x", requiresAuth: true)
            XCTFail("expected throw")
        } catch APIError.unauthorized {
            XCTAssertFalse(handlerCalled, "should not hit network when token missing")
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }
}
```

- [x] **Step 2: 运行测试确认编译失败**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/APIClientTests
```

预期：编译失败，报 `APIClient` 没有 `init(session:baseURL:)` 或 `request` 方法。

- [x] **Step 3: 实现 `APIClient.request<T>`**

把 `mall-ios/Core/Network/APIClient.swift` 替换为：

```swift
import Foundation

final class APIClient {
    static let shared = APIClient()

    private let session: URLSession
    private let baseURL: String

    init(session: URLSession = .shared, baseURL: String = "http://localhost:8080") {
        self.session = session
        self.baseURL = baseURL
    }

    /// 通用请求方法：解析统一信封 {code, msg, data}。
    /// - requiresAuth == true 时从 TokenStore 读取 token 注入 Authorization 头；
    ///   token 缺失时直接抛 .unauthorized，不发起网络请求。
    func request<T: Decodable>(
        _ path: String,
        method: HTTPMethod = .get,
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIError.transport(URLError(.badURL))
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body {
            urlRequest.httpBody = try JSONEncoder().encode(body)
        }

        if requiresAuth {
            guard let token = TokenStore.shared.getToken() else {
                throw APIError.unauthorized
            }
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.transport(URLError(.badServerResponse))
        }

        if http.statusCode == 401 {
            throw APIError.unauthorized
        }

        let envelope: ApiResponse<T>
        do {
            envelope = try JSONDecoder().decode(ApiResponse<T>.self, from: data)
        } catch {
            throw APIError.decoding(error)
        }

        guard envelope.code == 0 else {
            throw APIError.server(code: envelope.code, msg: envelope.msg)
        }

        if let payload = envelope.data {
            return payload
        }
        if let empty = EmptyResponse() as? T {
            return empty
        }
        throw APIError.decoding(
            DecodingError.valueNotFound(T.self, DecodingError.Context(codingPath: [], debugDescription: "data 字段缺失"))
        )
    }
}
```

- [x] **Step 4: 运行测试确认通过**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/APIClientTests
```

预期：`Test Suite 'APIClientTests' passed`，8 个测试全部通过。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add Core/Network/APIClient.swift MallAppTests/Core/Network/APIClientTests.swift MallApp.xcodeproj
git commit -m "feat(ios): implement APIClient.request generic envelope handling"
```

---

## Task 6: `APIClient.register` 非信封注册方法

**Files:**
- Modify: `mall-ios/Core/Network/APIClient.swift`
- Modify: `mall-ios/MallAppTests/Core/Network/APIClientTests.swift`（追加测试方法）

**Interfaces:**
- Produces:
  ```swift
  extension APIClient {
      func register(username: String, password: String) async throws -> String
  }
  ```
- Consumed by: Task 7（`AppSession.register`）。

- [x] **Step 1: 写失败的测试**

在 `mall-ios/MallAppTests/Core/Network/APIClientTests.swift` 的 `APIClientTests` 类末尾（`}` 之前）追加：

```swift
    func testRegister_returnsMessageOnSuccess() async throws {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"message":"save user success kane"}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        let message = try await apiClient.register(username: "kane", password: "111")

        XCTAssertEqual(message, "save user success kane")
    }

    func testRegister_throwsOnNon2xxResponse() async {
        MockURLProtocol.requestHandler = { request in
            let response = HTTPURLResponse(url: request.url!, statusCode: 500, httpVersion: nil, headerFields: nil)!
            return (response, Data())
        }

        do {
            _ = try await apiClient.register(username: "kane", password: "111")
            XCTFail("expected throw")
        } catch APIError.server {
            // expected
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }
```

- [x] **Step 2: 运行测试确认失败（编译错误：`register` 不存在）**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/APIClientTests
```

- [x] **Step 3: 实现 `register`**

在 `mall-ios/Core/Network/APIClient.swift` 的 `APIClient` 类内、`request<T>` 方法之后追加：

```swift

    /// /user/save 不遵循信封格式，直接返回 {"message": String}。
    /// 任何非 2xx 响应或缺少 message 字段均视为注册失败，不解析具体后端错误原因。
    func register(username: String, password: String) async throws -> String {
        guard let url = URL(string: baseURL + "/user/save") else {
            throw APIError.transport(URLError(.badURL))
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = HTTPMethod.post.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.httpBody = try JSONEncoder().encode([
            "username": username,
            "password": password
        ])

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            let code = (response as? HTTPURLResponse)?.statusCode ?? -1
            throw APIError.server(code: code, msg: "注册失败")
        }

        struct RegisterResponse: Decodable {
            let message: String
        }
        do {
            return try JSONDecoder().decode(RegisterResponse.self, from: data).message
        } catch {
            throw APIError.decoding(error)
        }
    }
```

- [x] **Step 4: 运行测试确认通过**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/APIClientTests
```

预期：`Test Suite 'APIClientTests' passed`，10 个测试全部通过。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add Core/Network/APIClient.swift MallAppTests/Core/Network/APIClientTests.swift MallApp.xcodeproj
git commit -m "feat(ios): add APIClient.register for non-envelope /user/save endpoint"
```

---

## Task 7: `AppSession` 会话状态

**Files:**
- Create: `mall-ios/Core/Auth/AppSession.swift`
- Create: `mall-ios/MallAppTests/Core/Auth/AppSessionTests.swift`

**Interfaces:**
- Consumes: `APIClient.request<T>`/`APIClient.register`（Task 5/6），`TokenStore.shared`（Task 3）。
- Produces:
  ```swift
  @Observable
  final class AppSession {
      static let shared: AppSession
      private(set) var userId: Int?
      private(set) var username: String?
      private(set) var avatar: String?
      var isLoggedIn: Bool { get }
      init(apiClient: APIClient = .shared, tokenStore: TokenStore = .shared, defaults: UserDefaults = .standard)
      func bootstrap()
      func login(username: String, password: String) async throws
      func register(username: String, password: String) async throws
      func logout()
  }
  ```
  `UserDefaults` 键名固定为 `"session.userId"` / `"session.username"` / `"session.avatar"`（后续任务和测试直接引用这三个字符串常量）。
- Consumed by: Task 8（`MallApp.swift` 调用 `bootstrap()`）、Task 10（`ProfileViewModel`）。

- [x] **Step 1: 写失败的测试**

创建 `mall-ios/MallAppTests/Core/Auth/AppSessionTests.swift`：

```swift
import XCTest
@testable import MallApp

final class AppSessionTests: XCTestCase {
    private var apiClient: APIClient!
    private var defaults: UserDefaults!
    private var session: AppSession!

    override func setUp() {
        super.setUp()
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        apiClient = APIClient(session: URLSession(configuration: config), baseURL: "https://mall.test")
        defaults = UserDefaults(suiteName: "AppSessionTests")!
        defaults.removePersistentDomain(forName: "AppSessionTests")
        TokenStore.shared.delete()
        session = AppSession(apiClient: apiClient, tokenStore: .shared, defaults: defaults)
    }

    override func tearDown() {
        MockURLProtocol.requestHandler = nil
        TokenStore.shared.delete()
        defaults.removePersistentDomain(forName: "AppSessionTests")
        super.tearDown()
    }

    func testLogin_success_updatesSessionState() async throws {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":0,"msg":"登录成功","data":{"user_id":1,"user_name":"kane","avatar":"https://a.png","token":"jwt-token"}}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        try await session.login(username: "kane", password: "111")

        XCTAssertTrue(session.isLoggedIn)
        XCTAssertEqual(session.userId, 1)
        XCTAssertEqual(session.username, "kane")
        XCTAssertEqual(session.avatar, "https://a.png")
        XCTAssertEqual(TokenStore.shared.getToken(), "jwt-token")
    }

    func testLogin_failure_keepsSessionLoggedOut() async {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":-1,"msg":"密码错误","data":null}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        do {
            try await session.login(username: "kane", password: "wrong")
            XCTFail("expected login to throw")
        } catch {
            // expected
        }

        XCTAssertFalse(session.isLoggedIn)
        XCTAssertNil(TokenStore.shared.getToken())
    }

    func testRegister_success_autoLogsIn() async throws {
        MockURLProtocol.requestHandler = { request in
            if request.url?.path == "/user/save" {
                let json = """
                {"message":"save user success kane"}
                """.data(using: .utf8)!
                let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
                return (response, json)
            }
            let json = """
            {"code":0,"msg":"登录成功","data":{"user_id":2,"user_name":"kane","avatar":"","token":"jwt-token-2"}}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }

        try await session.register(username: "kane", password: "111")

        XCTAssertTrue(session.isLoggedIn)
        XCTAssertEqual(session.userId, 2)
        XCTAssertEqual(TokenStore.shared.getToken(), "jwt-token-2")
    }

    func testRegister_failure_doesNotLogIn() async {
        MockURLProtocol.requestHandler = { request in
            let response = HTTPURLResponse(url: request.url!, statusCode: 500, httpVersion: nil, headerFields: nil)!
            return (response, Data())
        }

        do {
            try await session.register(username: "kane", password: "111")
            XCTFail("expected register to throw")
        } catch {
            // expected
        }

        XCTAssertFalse(session.isLoggedIn)
    }

    func testLogout_clearsSessionAndStorage() async throws {
        MockURLProtocol.requestHandler = { request in
            let json = """
            {"code":0,"msg":"登录成功","data":{"user_id":1,"user_name":"kane","avatar":"","token":"jwt-token"}}
            """.data(using: .utf8)!
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, json)
        }
        try await session.login(username: "kane", password: "111")

        session.logout()

        XCTAssertFalse(session.isLoggedIn)
        XCTAssertNil(session.userId)
        XCTAssertNil(session.username)
        XCTAssertNil(session.avatar)
        XCTAssertNil(TokenStore.shared.getToken())
    }

    func testBootstrap_restoresSessionWhenTokenAndDefaultsExist() {
        TokenStore.shared.save("saved-token")
        defaults.set(7, forKey: "session.userId")
        defaults.set("kane", forKey: "session.username")
        defaults.set("https://a.png", forKey: "session.avatar")

        session.bootstrap()

        XCTAssertTrue(session.isLoggedIn)
        XCTAssertEqual(session.userId, 7)
        XCTAssertEqual(session.username, "kane")
        XCTAssertEqual(session.avatar, "https://a.png")
    }

    func testBootstrap_leavesLoggedOutWhenNoSavedToken() {
        session.bootstrap()
        XCTAssertFalse(session.isLoggedIn)
    }
}
```

- [x] **Step 2: 运行测试确认编译失败**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/AppSessionTests
```

预期：编译失败，报 `cannot find 'AppSession' in scope`。

- [x] **Step 3: 实现 `AppSession`**

创建 `mall-ios/Core/Auth/AppSession.swift`：

```swift
import Foundation
import Observation

/// App 唯一会话真源。login/register 成功后把 token 存入 Keychain，
/// 用户信息存入 UserDefaults；bootstrap() 在启动时同步恢复，不发起网络请求。
@Observable
final class AppSession {
    static let shared = AppSession()

    private(set) var userId: Int?
    private(set) var username: String?
    private(set) var avatar: String?

    var isLoggedIn: Bool { userId != nil }

    private let apiClient: APIClient
    private let tokenStore: TokenStore
    private let defaults: UserDefaults

    private enum DefaultsKey {
        static let userId = "session.userId"
        static let username = "session.username"
        static let avatar = "session.avatar"
    }

    private struct LoginRequestBody: Encodable {
        let username: String
        let password: String
    }

    private struct LoginResponseData: Decodable {
        let userId: Int
        let userName: String
        let avatar: String
        let token: String

        enum CodingKeys: String, CodingKey {
            case userId = "user_id"
            case userName = "user_name"
            case avatar
            case token
        }
    }

    init(apiClient: APIClient = .shared, tokenStore: TokenStore = .shared, defaults: UserDefaults = .standard) {
        self.apiClient = apiClient
        self.tokenStore = tokenStore
        self.defaults = defaults
    }

    /// 启动时从本地存储恢复登录态。只读本地存储，无网络请求，不存在启动竞态。
    func bootstrap() {
        guard tokenStore.getToken() != nil else { return }
        guard defaults.object(forKey: DefaultsKey.userId) != nil else { return }
        userId = defaults.integer(forKey: DefaultsKey.userId)
        username = defaults.string(forKey: DefaultsKey.username)
        avatar = defaults.string(forKey: DefaultsKey.avatar)
    }

    func login(username: String, password: String) async throws {
        let data: LoginResponseData = try await apiClient.request(
            "/user/login",
            method: .post,
            body: LoginRequestBody(username: username, password: password),
            requiresAuth: false
        )
        tokenStore.save(data.token)
        defaults.set(data.userId, forKey: DefaultsKey.userId)
        defaults.set(data.userName, forKey: DefaultsKey.username)
        defaults.set(data.avatar, forKey: DefaultsKey.avatar)
        self.userId = data.userId
        self.username = data.userName
        self.avatar = data.avatar
    }

    /// 注册即登录：/user/save 不返回 token，注册成功后自动调用 login。
    func register(username: String, password: String) async throws {
        _ = try await apiClient.register(username: username, password: password)
        try await login(username: username, password: password)
    }

    func logout() {
        tokenStore.delete()
        defaults.removeObject(forKey: DefaultsKey.userId)
        defaults.removeObject(forKey: DefaultsKey.username)
        defaults.removeObject(forKey: DefaultsKey.avatar)
        userId = nil
        username = nil
        avatar = nil
    }
}
```

- [x] **Step 4: 运行测试确认通过**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test -only-testing:MallAppTests/AppSessionTests
```

预期：`Test Suite 'AppSessionTests' passed`，7 个测试全部通过。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add Core/Auth/AppSession.swift MallAppTests/Core/Auth/AppSessionTests.swift MallApp.xcodeproj
git commit -m "feat(ios): add AppSession with login/register/logout/bootstrap"
```

---

## Task 8: `MallApp.swift` 启动时恢复会话

**Files:**
- Modify: `mall-ios/MallApp.swift`

**Interfaces:**
- Consumes: `AppSession.shared.bootstrap()`（Task 7）。

- [x] **Step 1: 修改 `MallApp.swift`**

```swift
import SwiftUI

@main
struct MallApp: App {
    init() {
        AppSession.shared.bootstrap()
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

- [x] **Step 2: 构建验证**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' build
```

预期：`** BUILD SUCCEEDED **`。

- [x] **Step 3: Commit**

```bash
cd mall-ios && git add MallApp.swift MallApp.xcodeproj
git commit -m "feat(ios): call AppSession.bootstrap() on app launch"
```

---

## Task 9: 四 Tab 壳与消息占位

**Files:**
- Create: `mall-ios/Features/Chat/ViewModel/ChatListViewModel.swift`
- Create: `mall-ios/Features/Chat/View/ChatListView.swift`
- Modify: `mall-ios/ContentView.swift`

**Interfaces:**
- Produces: `ChatListView`（供 `ContentView` 引用）、`ChatListViewModel`（占位，真实实现留给后续 change）。
- 不写单元测试（纯 SwiftUI 视图布局 + 占位 ViewModel，无逻辑分支）。

- [x] **Step 1: 创建 `ChatListViewModel`**

创建 `mall-ios/Features/Chat/ViewModel/ChatListViewModel.swift`（与现有 `HomeViewModel`/`PublishViewModel` 占位模式一致）：

```swift
import Observation

@Observable
final class ChatListViewModel {
    var title = "消息"
}
```

- [x] **Step 2: 创建 `ChatListView`**

创建 `mall-ios/Features/Chat/View/ChatListView.swift`：

```swift
import SwiftUI

struct ChatListView: View {
    @State private var viewModel = ChatListViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
```

- [x] **Step 3: 修改 `ContentView.swift` 扩展为四 Tab**

图标对齐 mall-mini `app.json` 的 Tab 顺序（首页/发布/消息/我的）：

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
            ChatListView()
                .tabItem {
                    Label("消息", systemImage: "bubble.left.and.bubble.right")
                }
            ProfileView()
                .tabItem {
                    Label("我的", systemImage: "person")
                }
        }
    }
}
```

- [x] **Step 4: 构建验证**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' build
```

预期：`** BUILD SUCCEEDED **`。

- [x] **Step 5: Commit**

```bash
cd mall-ios && git add Features/Chat ContentView.swift MallApp.xcodeproj
git commit -m "feat(ios): add Chat tab placeholder, expand TabView to four tabs"
```

---

## Task 10: `ProfileViewModel` 重写 —— 表单状态与错误反馈

**Files:**
- Modify: `mall-ios/Features/Profile/ViewModel/ProfileViewModel.swift`

**Interfaces:**
- Consumes: `AppSession`（Task 7）：`session.isLoggedIn` / `session.username` / `session.avatar` / `session.login` / `session.register` / `session.logout`。
- Produces:
  ```swift
  @Observable
  final class ProfileViewModel {
      var username: String
      var password: String
      var isRegisterMode: Bool
      var errorMessage: String?
      private(set) var isSubmitting: Bool
      var isLoggedIn: Bool { get }
      var currentUsername: String? { get }
      var currentAvatar: String? { get }
      init(session: AppSession = .shared)
      func toggleMode()
      func submit() async
      func logout()
  }
  ```
- Consumed by: Task 11（`ProfileView`）。
- 不写单元测试（表单校验逻辑是 `AppSession` 已测试路径的薄封装；覆盖范围按设计文档限定在 `APIClient`/`AppSession`，不含 ViewModel/View）。

- [x] **Step 1: 重写 `ProfileViewModel.swift`**

```swift
import Observation

@Observable
final class ProfileViewModel {
    var username = ""
    var password = ""
    var isRegisterMode = false
    var errorMessage: String?
    private(set) var isSubmitting = false

    private let session: AppSession

    init(session: AppSession = .shared) {
        self.session = session
    }

    var isLoggedIn: Bool { session.isLoggedIn }
    var currentUsername: String? { session.username }
    var currentAvatar: String? { session.avatar }

    func toggleMode() {
        isRegisterMode.toggle()
        errorMessage = nil
    }

    func submit() async {
        guard !username.isEmpty, !password.isEmpty else {
            errorMessage = "用户名和密码不能为空"
            return
        }
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }
        do {
            if isRegisterMode {
                try await session.register(username: username, password: password)
            } else {
                try await session.login(username: username, password: password)
            }
            username = ""
            password = ""
        } catch {
            errorMessage = isRegisterMode ? "注册失败，请更换用户名后重试" : "登录失败，请检查用户名或密码"
        }
    }

    func logout() {
        session.logout()
    }
}
```

- [x] **Step 2: 构建验证（`ProfileView` 尚未更新，暂时会因引用旧 `title` 属性报错，属预期，留到 Task 11 一并解决）**

跳过独立构建，Task 11 完成后一起验证（`ProfileViewModel`/`ProfileView` 是同一功能的两半，中间态不具备独立可编译性——参照 writing-plans "Task Right-Sizing" 原则，这两个文件在同一次 commit 中一起提交）。

- [x] **Step 3: 不单独 commit，与 Task 11 合并提交**

---

## Task 11: `ProfileView` 重写 —— 登录/注册表单 + 已登录态卡片

**Files:**
- Modify: `mall-ios/Features/Profile/View/ProfileView.swift`

**Interfaces:**
- Consumes: `ProfileViewModel`（Task 10）的全部属性与方法。

- [x] **Step 1: 重写 `ProfileView.swift`**

```swift
import SwiftUI

struct ProfileView: View {
    @State private var viewModel = ProfileViewModel()
    @State private var showLogoutConfirm = false

    var body: some View {
        Group {
            if viewModel.isLoggedIn {
                loggedInView
            } else {
                authFormView
            }
        }
        .alert(
            "提示",
            isPresented: Binding(
                get: { viewModel.errorMessage != nil },
                set: { isPresented in
                    if !isPresented {
                        viewModel.errorMessage = nil
                    }
                }
            )
        ) {
            Button("确定", role: .cancel) {}
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
    }

    private var loggedInView: some View {
        VStack(spacing: 16) {
            Image(systemName: "person.circle.fill")
                .resizable()
                .frame(width: 72, height: 72)
                .foregroundStyle(.gray)
            Text(viewModel.currentUsername ?? "")
                .font(.title3)
            Button("退出登录", role: .destructive) {
                showLogoutConfirm = true
            }
        }
        .padding()
        .confirmationDialog(
            "确认退出登录吗？",
            isPresented: $showLogoutConfirm,
            titleVisibility: .visible
        ) {
            Button("退出登录", role: .destructive) {
                viewModel.logout()
            }
            Button("取消", role: .cancel) {}
        }
    }

    private var authFormView: some View {
        VStack(spacing: 16) {
            Picker("", selection: $viewModel.isRegisterMode) {
                Text("登录").tag(false)
                Text("注册").tag(true)
            }
            .pickerStyle(.segmented)

            TextField("用户名", text: $viewModel.username)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .textFieldStyle(.roundedBorder)

            SecureField("密码", text: $viewModel.password)
                .textFieldStyle(.roundedBorder)

            Button(viewModel.isRegisterMode ? "注册" : "登录") {
                Task { await viewModel.submit() }
            }
            .disabled(viewModel.isSubmitting)
            .buttonStyle(.borderedProminent)
        }
        .padding()
    }
}
```

- [x] **Step 2: 运行完整测试套件验证无回归**

```bash
cd mall-ios && xcodegen generate && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test
```

预期：全部测试通过（`SmokeTests`、`MockURLProtocolTests`、`TokenStoreTests`、`ApiResponseTests`、`APIClientTests`、`AppSessionTests`）。

- [x] **Step 3: 构建验证**

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' build
```

预期：`** BUILD SUCCEEDED **`。

- [x] **Step 4: Commit（Task 10 + Task 11 一并提交）**

```bash
cd mall-ios && git add Features/Profile/ViewModel/ProfileViewModel.swift Features/Profile/View/ProfileView.swift MallApp.xcodeproj
git commit -m "feat(ios): implement Profile login/register form and logged-in state"
```

---

## Task 12: 全量验证

**Files:** 无新增/修改文件，仅验证。

- [x] **Step 1: 全量构建**

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' build
```

预期：`** BUILD SUCCEEDED **`。（对应 tasks.md 6.1）

- [x] **Step 2: 运行全部单元测试**

```bash
cd mall-ios && xcodebuild -project MallApp.xcodeproj -scheme MallApp -destination 'platform=iOS Simulator,name=iPhone 17' test
```

预期：`** TEST SUCCEEDED **`，涵盖 `SmokeTests`（1）、`MockURLProtocolTests`（1）、`TokenStoreTests`（4）、`ApiResponseTests`（3）、`APIClientTests`（10）、`AppSessionTests`（7），共 26 个测试全部通过。（对应 tasks.md 6.2）

- [x] **Step 3: 手动验证 —— 启动后端**

```bash
cd mall-server && go build -o mall-server && ./mall-server web -config configs/config.yaml
```

确认服务监听 `http://localhost:8080`。

- [x] **Step 4: 手动验证 —— 完整登录态生命周期（对应 tasks.md 6.3）**

> **接受基于自动化证据的验证，非人工点击走查**：详见 `.superpowers/sdd/task-12-manual-verification-note.md`。

在模拟器中运行 App（Xcode 打开 `mall-ios/MallApp.xcodeproj` 或 `xcodebuild ... -destination '...' run`，也可用 `xcrun simctl` 配合已构建的 `.app`），依次验证：

1. 首次启动，"我的" Tab 显示登录/注册表单（默认登录模式）。
2. 切换到"注册"模式，输入一个新用户名（例如 `manual-test-user`）与密码，点击"注册"。
3. 预期：注册成功后自动登录，界面切换为已登录态卡片，显示刚才输入的用户名。
4. 杀掉 App（模拟器长按或 `xcrun simctl terminate booted com.secondhandmall.MallApp`），重新启动。
5. 预期：登录态保留（`bootstrap()` 从 Keychain + UserDefaults 恢复），无需重新登录，"我的" Tab 直接显示已登录态卡片。
6. 点击"退出登录"，在二次确认弹窗中点击"退出登录"。
7. 预期：界面回退到登录/注册表单；再次杀掉重启 App，确认登录态未被误恢复（Keychain 中 token 已清除）。

- [x] **Step 5: 记录验证结果**

若手动验证全部通过，此计划的实现阶段完成，可以进入 Comet `verify` 阶段（`comet-verify`）做正式验证收尾；若发现偏差，参照 `superpowers:systematic-debugging` 定位根因后回到对应 Task 修复并重新提交。

---

## Self-Review 备注（供执行者参考，无需重复劳动）

- **Spec 覆盖**：`ios-network-client` 的 5 个 Requirement（信封解析、鉴权注入、错误分类、非信封注册）由 Task 4/5/6 覆盖；`ios-account-session` 的 5 个 Requirement（登录、注册即登录、登出、启动恢复、401 触发登出契约）由 Task 7 覆盖，401→logout 的调用方责任在 Task 10 的 `submit()` 错误分支体现（当前范围内登录/注册请求本身不会收到需要鉴权的 401，该契约留给后续引入鉴权接口调用的 change 落地为具体调用点，`AppSession.logout()` 已就绪可供其调用）。
- **命名一致性**：`APIClient.request<T>`、`APIError` 四个 case、`HTTPMethod` 四个 case、`TokenStore.save/getToken/delete`、`AppSession.login/register/logout/bootstrap`、`ProfileViewModel` 属性名在所有任务中保持完全一致的拼写与签名。
- **依赖顺序**：`TokenStore`（Task 3）先于 `APIClient.request`（Task 5）实现，因为后者的 `requiresAuth` 分支直接引用 `TokenStore.shared`。
