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
