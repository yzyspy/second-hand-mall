import SwiftUI

struct ProfileView: View {
    @State private var viewModel = ProfileViewModel()
    @State private var showLogoutConfirm = false

    var body: some View {
        Group {
            if viewModel.isLoggedIn {
                loggedInView
            } else {
                authFormView
            }
        }
        .alert(
            "提示",
            isPresented: Binding(
                get: { viewModel.errorMessage != nil },
                set: { isPresented in
                    if !isPresented {
                        viewModel.errorMessage = nil
                    }
                }
            )
        ) {
            Button("确定", role: .cancel) {}
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
    }

    private var loggedInView: some View {
        VStack(spacing: 16) {
            Image(systemName: "person.circle.fill")
                .resizable()
                .frame(width: 72, height: 72)
                .foregroundStyle(.gray)
            Text(viewModel.currentUsername ?? "")
                .font(.title3)
            Button("退出登录", role: .destructive) {
                showLogoutConfirm = true
            }
        }
        .padding()
        .confirmationDialog(
            "确认退出登录吗？",
            isPresented: $showLogoutConfirm,
            titleVisibility: .visible
        ) {
            Button("退出登录", role: .destructive) {
                viewModel.logout()
            }
            Button("取消", role: .cancel) {}
        }
    }

    private var authFormView: some View {
        VStack(spacing: 16) {
            Picker("", selection: $viewModel.isRegisterMode) {
                Text("登录").tag(false)
                Text("注册").tag(true)
            }
            .pickerStyle(.segmented)

            TextField("用户名", text: $viewModel.username)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .textFieldStyle(.roundedBorder)

            SecureField("密码", text: $viewModel.password)
                .textFieldStyle(.roundedBorder)

            Button(viewModel.isRegisterMode ? "注册" : "登录") {
                Task { await viewModel.submit() }
            }
            .disabled(viewModel.isSubmitting)
            .buttonStyle(.borderedProminent)
        }
        .padding()
    }
}
