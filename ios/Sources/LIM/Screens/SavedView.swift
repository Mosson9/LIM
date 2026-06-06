import SwiftUI

/// "你忍住了 ✦" — the celebratory screen after resisting. The tree grows and the
/// saved amount is added to the running total (the original "Less" logic).
struct SavedView: View {
    @EnvironmentObject var model: AppModel
    let decision: Decision
    @State private var grown = false

    var body: some View {
        VStack(spacing: 0) {
            Spacer()
            GrowthTree(stage: grown ? min(6, model.stats.treeStage + 1) : model.stats.treeStage, size: 196)
                .scaleEffect(grown ? 1 : 0.7)
                .animation(.spring(response: 0.9, dampingFraction: 0.7), value: grown)

            Kicker(text: "你忍住了 ✦", color: Theme.savedDeep).padding(.top, 6).padding(.bottom, 14)

            VStack(spacing: 4) {
                Text("为自己省下").font(Theme.display(30)).foregroundColor(Theme.ink)
                Text(Format.yuan(decision.price)).font(Theme.display(48)).foregroundColor(Theme.savedDeep)
            }
            Text("没买下「\(decision.item)」，你的小树又长了一点。少一件不需要的东西，多一份从容。")
                .font(Theme.sans(15)).foregroundColor(Theme.ink2)
                .multilineTextAlignment(.center).lineSpacing(5)
                .frame(maxWidth: 290).padding(.top, 18)

            Card {
                HStack {
                    stat("累计省下", Format.yuan(model.stats.totalSaved))
                    Divider().frame(height: 36)
                    stat("连续克制", "\(model.stats.streak) 天")
                }
            }
            .padding(.top, 26)

            Spacer()
            VStack(spacing: 10) {
                Button { model.go(.growth) } label: {
                    HStack(spacing: 8) { Icon(name: "sprout", size: 18, color: .white); Text("看看我的成长") }
                }.buttonStyle(FilledButtonStyle(bg: Theme.saved))
                Button("回到首页") { model.go(.home) }.buttonStyle(GhostButtonStyle())
            }
            .padding(.bottom, 30)
        }
        .padding(.horizontal, 24)
        .background(
            LinearGradient(colors: [Theme.savedSoft, Theme.paper], startPoint: .top, endPoint: .center)
                .ignoresSafeArea())
        .onAppear { grown = true }
    }

    private func stat(_ label: String, _ value: String) -> some View {
        VStack(spacing: 3) {
            Text(label).font(Theme.sans(12)).foregroundColor(Theme.ink3)
            Text(value).font(Theme.display(24)).foregroundColor(Theme.savedDeep)
        }
        .frame(maxWidth: .infinity)
    }
}
