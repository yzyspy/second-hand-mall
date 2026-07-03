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
