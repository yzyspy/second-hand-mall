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
