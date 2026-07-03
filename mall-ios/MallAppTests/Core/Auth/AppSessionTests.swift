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
