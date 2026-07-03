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
