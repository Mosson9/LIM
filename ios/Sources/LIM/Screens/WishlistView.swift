import SwiftUI

/// 冷静期 · 心愿单 — the 24h cooling-off list. Items show a countdown ring; once
/// expired the user re-decides (resist → savings, or buy → record).
struct WishlistView: View {
    @EnvironmentObject var model: AppModel
    @State private var working: String?

    var body: some View {
        VStack(spacing: 0) {
            TopBar(title: "冷静期 · 心愿单")
            ScrollView {
                VStack(spacing: 14) {
                    explainer
                    if model.wishlist.isEmpty { empty }
                    ForEach(model.wishlist) { item in card(item) }
                    Color.clear.frame(height: 30)
                }
                .padding(.horizontal, 22).padding(.top, 4)
            }
        }
        .background(Theme.paper.ignoresSafeArea())
        .task { await model.refreshAll() }
    }

    private var explainer: some View {
        Card(padding: 15, background: Theme.indigoTint) {
            HStack(alignment: .top, spacing: 11) {
                Icon(name: "clock", size: 20, color: Theme.indigo)
                Text("还想要的东西，先在这里放 24 小时。很多冲动，睡一觉就过去了。到期我会提醒你再决定一次。")
                    .font(Theme.sans(13.5)).foregroundColor(Theme.indigoInk).lineSpacing(4)
            }
        }
        .overlay(RoundedRectangle(cornerRadius: Theme.radius).stroke(Color(hex: 0xE4E4F2), lineWidth: 1))
    }

    private var empty: some View {
        VStack(spacing: 6) {
            Icon(name: "leaf", size: 40, color: Theme.ink4).padding(.bottom, 8)
            Text("心愿单空空的").font(Theme.sans(15)).foregroundColor(Theme.ink3)
            Text("这是一种很轻盈的状态。").font(Theme.sans(13)).foregroundColor(Theme.ink4)
        }
        .frame(maxWidth: .infinity).padding(.vertical, 60)
    }

    private func card(_ w: WishlistItem) -> some View {
        let cat = model.category(w.cat)
        return Card {
            VStack(spacing: 14) {
                HStack(spacing: 13) {
                    Icon(name: cat.icon, size: 22, color: Color(webHex: cat.color))
                        .frame(width: 46, height: 46)
                        .background(Color(webHex: cat.soft))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 3) {
                        Text(w.item).font(Theme.sans(16, .semibold)).foregroundColor(Theme.ink).lineLimit(1)
                        HStack(spacing: 8) {
                            Text(Format.yuan(w.price)).font(Theme.sans(13.5)).foregroundColor(Theme.ink2)
                            Badge(text: "冲动 \(w.impulse)", fg: Theme.warn, bg: Theme.warnSoft)
                        }
                    }
                    Spacer()
                    ZStack {
                        ProgressRing(progress: w.progress, color: w.expired ? Theme.saved : Theme.indigo)
                            .frame(width: 46, height: 46)
                        Icon(name: "clock", size: 18, color: w.expired ? Theme.saved : Theme.indigo)
                    }
                }
                HStack {
                    Text(w.expired ? "冷静期已过" : w.remainText)
                        .font(Theme.sans(12.5, .semibold))
                        .foregroundColor(w.expired ? Theme.savedDeep : Theme.indigo)
                    Spacer()
                }
                actions(w)
            }
        }
    }

    @ViewBuilder
    private func actions(_ w: WishlistItem) -> some View {
        HStack(spacing: 9) {
            Button { Task { await resolve(w, "resist") } } label: {
                Text(w.expired ? "还是算了 · 省下" : "不买了")
                    .font(Theme.sans(14, .semibold)).foregroundColor(.white)
                    .frame(maxWidth: .infinity).padding(.vertical, 11)
                    .background(Theme.saved).clipShape(RoundedRectangle(cornerRadius: 12))
            }
            Button { Task { await resolve(w, "buy") } } label: {
                Text("仍然要买")
                    .font(Theme.sans(14, .semibold)).foregroundColor(Theme.ink2)
                    .frame(maxWidth: .infinity).padding(.vertical, 11)
                    .background(Theme.surface).clipShape(RoundedRectangle(cornerRadius: 12))
                    .overlay(RoundedRectangle(cornerRadius: 12).stroke(Theme.hairline, lineWidth: 1))
            }
        }
        .disabled(working == w.id)
    }

    private func resolve(_ w: WishlistItem, _ action: String) async {
        working = w.id
        try? await APIClient.shared.resolveWishlist(id: w.id, action: action)
        await model.refreshAll()
        working = nil
    }
}
