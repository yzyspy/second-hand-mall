import SwiftUI

struct ProfileView: View {
    @State private var session = UserSession.shared

    var body: some View {
        NavigationStack {
            Group {
                if session.isLoggedIn {
                    loggedInContent
                } else {
                    AuthFormView()
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .background(Color(.systemGray6))
            .navigationTitle("我的")
            .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var loggedInContent: some View {
        VStack(spacing: 12) {
            HStack(spacing: 16) {
                AsyncImage(url: URL(string: session.avatar)) { phase in
                    if case .success(let image) = phase {
                        image.resizable().scaledToFill()
                    } else {
                        Image(systemName: "person.crop.circle.fill")
                            .resizable()
                            .foregroundStyle(Color(.systemGray3))
                    }
                }
                .frame(width: 64, height: 64)
                .clipShape(Circle())

                VStack(alignment: .leading, spacing: 4) {
                    Text(session.userName)
                        .font(.title3.bold())
                    Text("ID: \(session.userID)")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                Spacer()
            }
            .padding(16)
            .background(Color(.systemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            Button(role: .destructive) {
                session.signOut()
            } label: {
                Text("退出登录")
                    .font(.headline)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 12)
                    .background(Color(.systemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
        .padding(12)
    }
}

/// 登录 / 注册表单
struct AuthFormView: View {
    private enum Mode: String, CaseIterable {
        case login = "登录"
        case register = "注册"
    }

    @State private var mode: Mode = .login
    @State private var username = ""
    @State private var password = ""
    @State private var submitting = false
    @State private var errorMessage: String?

    private var canSubmit: Bool {
        username.trimmingCharacters(in: .whitespaces).count >= 2
            && password.count >= 6
            && !submitting
    }

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "leaf.circle.fill")
                .font(.system(size: 64))
                .foregroundStyle(Color.mallGreen)
                .padding(.top, 40)

            Text("旧物新语")
                .font(.title2.bold())

            Picker("模式", selection: $mode) {
                ForEach(Mode.allCases, id: \.self) { mode in
                    Text(mode.rawValue).tag(mode)
                }
            }
            .pickerStyle(.segmented)
            .padding(.horizontal, 40)

            VStack(spacing: 12) {
                TextField("用户名", text: $username)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .padding(14)
                    .background(Color(.systemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))

                SecureField("密码（至少6位）", text: $password)
                    .padding(14)
                    .background(Color(.systemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
            .padding(.horizontal, 24)

            if let errorMessage {
                Text(errorMessage)
                    .font(.footnote)
                    .foregroundStyle(.red)
                    .padding(.horizontal, 24)
            }

            Button {
                Task { await submit() }
            } label: {
                Group {
                    if submitting {
                        ProgressView().tint(.white)
                    } else {
                        Text(mode.rawValue)
                    }
                }
                .font(.headline)
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(canSubmit ? Color.mallGreen : Color(.systemGray3))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .disabled(!canSubmit)
            .padding(.horizontal, 24)

            Spacer()
        }
        #if DEBUG
        // 供 UI 调试/截图自动化自动提交表单：
        // SIMCTL_CHILD_DEBUG_AUTO_AUTH="<用户名>:<密码>:<login|register>"
        .onAppear {
            guard let spec = ProcessInfo.processInfo.environment["DEBUG_AUTO_AUTH"] else { return }
            let parts = spec.split(separator: ":")
            guard parts.count == 3 else { return }
            username = String(parts[0])
            password = String(parts[1])
            mode = parts[2] == "register" ? .register : .login
            Task { await submit() }
        }
        #endif
    }

    private func submit() async {
        submitting = true
        defer { submitting = false }
        do {
            let name = username.trimmingCharacters(in: .whitespaces)
            let result = mode == .login
                ? try await AuthAPI.login(username: name, password: password)
                : try await AuthAPI.register(username: name, password: password)
            UserSession.shared.signIn(result)
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

#Preview {
    ProfileView()
}
