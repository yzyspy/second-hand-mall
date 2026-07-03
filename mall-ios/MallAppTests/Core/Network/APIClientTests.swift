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
