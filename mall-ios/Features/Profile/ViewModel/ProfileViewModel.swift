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
