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
