import Foundation

/// POST /user/login、/user/save 的响应 data
struct AuthResult: Decodable {
    let userID: Int
    let userName: String
    let avatar: String
    let token: String

    enum CodingKeys: String, CodingKey {
        case avatar, token
        case userID = "user_id"
        case userName = "user_name"
    }
}

enum AuthAPI {
    private struct Credentials: Encodable {
        let username: String
        let password: String
    }

    /// POST /user/login 用户名密码登录
    static func login(username: String, password: String) async throws -> AuthResult {
        try await API.post("/user/login", body: Credentials(username: username, password: password))
    }

    /// POST /user/save 注册，成功直接返回 token
    static func register(username: String, password: String) async throws -> AuthResult {
        try await API.post("/user/save", body: Credentials(username: username, password: password))
    }
}

/// 登录态，token 与用户信息持久化在 UserDefaults
/// （与小程序 wx.setStorageSync('token'/'userInfo'/'userId') 的角色一致）
@MainActor
@Observable
final class UserSession {
    static let shared = UserSession()

    private(set) var userID: Int
    private(set) var userName: String
    private(set) var avatar: String
    /// 作为可观察的存储属性保存一份 token，登录/登出时同步写 UserDefaults；
    /// 若直接读 API.token（UserDefaults），SwiftUI 无法观察其变化，登录后页面不会刷新
    private(set) var token: String

    var isLoggedIn: Bool {
        !token.isEmpty
    }

    private init() {
        let defaults = UserDefaults.standard
        userID = defaults.integer(forKey: "userId")
        userName = defaults.string(forKey: "userName") ?? ""
        avatar = defaults.string(forKey: "userAvatar") ?? ""
        token = API.token ?? ""
    }

    func signIn(_ result: AuthResult) {
        API.token = result.token
        token = result.token
        userID = result.userID
        userName = result.userName
        avatar = result.avatar

        let defaults = UserDefaults.standard
        defaults.set(result.userID, forKey: "userId")
        defaults.set(result.userName, forKey: "userName")
        defaults.set(result.avatar, forKey: "userAvatar")
    }

    func signOut() {
        API.token = nil
        token = ""
        userID = 0
        userName = ""
        avatar = ""

        let defaults = UserDefaults.standard
        defaults.removeObject(forKey: "userId")
        defaults.removeObject(forKey: "userName")
        defaults.removeObject(forKey: "userAvatar")
    }
}
