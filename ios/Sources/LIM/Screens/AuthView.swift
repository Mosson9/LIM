import SwiftUI

/// Sign-in / sign-up. The backend requires an account before any analysis, so
/// this gates the app. Demo accounts (see docs): lin@lim.app / password.
struct AuthView: View {
    @EnvironmentObject var model: AppModel
    @State private var mode: Mode = .login
    @State private var email = ""
    @State private var password = ""
    @State private var name = ""
    @State private var loading = false

    enum Mode { case login, register }

    var body: some View {
        ZStack {
            Theme.paper.ignoresSafeArea()
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    RoundedRectangle(cornerRadius: 22, style: .continuous)
                        .fill(Theme.indigo)
                        .frame(width: 64, height: 64)
                        .overlay(Text("L").font(Theme.display(34)).foregroundColor(.white))
                        .padding(.top, 60).padding(.bottom, 28)

                    Kicker(text: "LESS IS MORE", color: Theme.indigo)
                    Text(mode == .login ? "欢迎回来" : "开始更清醒的生活")
                        .font(Theme.display(34)).foregroundColor(Theme.ink)
                        .padding(.top, 8)
                    Text("买之前，先问问自己。")
                        .font(Theme.sans(15)).foregroundColor(Theme.ink2)
                        .padding(.top, 6).padding(.bottom, 30)

                    if mode == .register {
                        field("昵称", text: $name, placeholder: "怎么称呼你")
                    }
                    field("邮箱", text: $email, placeholder: "you@example.com",
                          keyboard: .emailAddress)
                    field("密码", text: $password, placeholder: "至少 6 位", secure: true)

                    if let err = model.authError {
                        Text(err).font(Theme.sans(13)).foregroundColor(Theme.danger)
                            .padding(.top, 14)
                    }

                    Button {
                        Task { await submit() }
                    } label: {
                        if loading { ProgressView().tint(.white) }
                        else { Text(mode == .login ? "登录" : "注册并开始") }
                    }
                    .buttonStyle(FilledButtonStyle())
                    .disabled(loading || email.isEmpty || password.count < 6)
                    .opacity((email.isEmpty || password.count < 6) ? 0.5 : 1)
                    .padding(.top, 26)

                    Button {
                        model.authError = nil
                        mode = mode == .login ? .register : .login
                    } label: {
                        Text(mode == .login ? "还没有账号？去注册" : "已有账号？去登录")
                            .font(Theme.sans(14)).foregroundColor(Theme.indigo)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.top, 18)
                }
                .padding(.horizontal, 24)
            }
        }
    }

    private func submit() async {
        loading = true
        defer { loading = false }
        if mode == .login {
            await model.login(email: email, password: password)
        } else {
            await model.register(email: email, password: password,
                                 name: name.isEmpty ? "LIM 用户" : name)
        }
    }

    private func field(_ label: String, text: Binding<String>, placeholder: String,
                       keyboard: UIKeyboardType = .default, secure: Bool = false) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(label).font(Theme.sans(12.5, .semibold)).foregroundColor(Theme.ink3)
            Group {
                if secure { SecureField(placeholder, text: text) }
                else { TextField(placeholder, text: text).keyboardType(keyboard) }
            }
            .textInputAutocapitalization(.never)
            .autocorrectionDisabled()
            .font(Theme.sans(16))
            .padding(14)
            .background(Theme.surface)
            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 14).stroke(Theme.hairline, lineWidth: 1))
        }
        .padding(.bottom, 16)
    }
}
