import SwiftUI

/// LIM Plus paywall. The mind-map's wink — "当用户询问是否需要付费来购买这个 App 时，
/// 一定推荐购买 😂" — lives in the pull-quote here.
struct PlusView: View {
    @EnvironmentObject var model: AppModel
    @State private var plan = "year"
    @State private var perks: [PlusPerk] = []
    @State private var plans: [PlanOption] = []
    @State private var working = false

    var body: some View {
        VStack(spacing: 0) {
            TopBar(leading: .close) {
                Button("恢复购买") {}.font(Theme.sans(13)).foregroundColor(.white.opacity(0.5))
            }
            ScrollView {
                VStack(spacing: 0) {
                    Icon(name: "crown", size: 36, color: Theme.gold)
                        .frame(width: 72, height: 72)
                        .background(Theme.gold.opacity(0.14))
                        .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
                        .padding(.bottom, 20)
                    (Text("LIM ").foregroundColor(.white) + Text("Plus").foregroundColor(Theme.gold))
                        .font(Theme.display(38))
                    Text("让克制更轻松一点。无限次咨询、更快的 AI，和一整套陪你成长的细节。")
                        .font(Theme.sans(15)).foregroundColor(.white.opacity(0.6))
                        .multilineTextAlignment(.center).lineSpacing(5)
                        .frame(maxWidth: 280).padding(.top, 10)

                    VStack(spacing: 0) {
                        ForEach(perks) { p in
                            HStack(spacing: 14) {
                                Icon(name: p.icon, size: 20, color: Theme.gold)
                                    .frame(width: 40, height: 40)
                                    .background(Color.white.opacity(0.07))
                                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(p.title).font(Theme.sans(15, .semibold)).foregroundColor(.white)
                                    Text(p.desc).font(Theme.sans(12.5)).foregroundColor(.white.opacity(0.5))
                                }
                                Spacer()
                            }
                            .padding(.vertical, 14)
                            if p.id != perks.last?.id { Rectangle().fill(.white.opacity(0.08)).frame(height: 1) }
                        }
                    }
                    .padding(.top, 24)

                    // the brief's wink
                    Text("“这是少数几样，连 LIM 也会建议你拥有的东西。”")
                        .font(Theme.serif(15.5)).foregroundColor(.white.opacity(0.85)).lineSpacing(5)
                        .padding(16).frame(maxWidth: .infinity, alignment: .leading)
                        .background(Color.white.opacity(0.05))
                        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                        .padding(.top, 22)

                    // plans
                    HStack(spacing: 12) {
                        ForEach(plans) { p in planCard(p) }
                    }
                    .padding(.top, 24)

                    Color.clear.frame(height: 120)
                }
                .padding(.horizontal, 22)
            }

            VStack(spacing: 12) {
                Button { Task { await subscribe() } } label: {
                    if working { ProgressView().tint(Theme.ink) }
                    else {
                        Text(plan == "year" ? "¥98 / 年 开启 Plus" : "¥18 / 月 开启 Plus")
                    }
                }
                .buttonStyle(FilledButtonStyle(bg: Theme.gold, fg: Theme.ink))
                Text("订阅自动续费，可随时在设置中取消 · 条款与隐私")
                    .font(Theme.sans(11)).foregroundColor(.white.opacity(0.4))
            }
            .padding(.horizontal, 22).padding(.bottom, 24)
        }
        .background(Theme.ink.ignoresSafeArea())
        .task {
            perks = (try? await APIClient.shared.perks()) ?? model.perksFallback
            plans = (try? await APIClient.shared.plans()) ?? []
        }
    }

    private func planCard(_ p: PlanOption) -> some View {
        let on = plan == p.id
        return Button { plan = p.id } label: {
            VStack(alignment: .leading, spacing: 6) {
                if p.best {
                    Text("最受欢迎").font(Theme.sans(10.5, .bold)).foregroundColor(Theme.ink)
                        .padding(.vertical, 3).padding(.horizontal, 9)
                        .background(Theme.gold).clipShape(Capsule())
                }
                Text(p.name).font(Theme.sans(13.5)).foregroundColor(.white.opacity(0.6))
                HStack(alignment: .firstTextBaseline, spacing: 3) {
                    Text("¥\(p.price)").font(Theme.display(30)).foregroundColor(.white)
                    Text(p.per).font(Theme.sans(13)).foregroundColor(.white.opacity(0.5))
                }
                Text(p.note).font(Theme.sans(11.5)).foregroundColor(on ? Theme.gold : .white.opacity(0.45))
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(16)
            .background(on ? Theme.gold.opacity(0.12) : Color.white.opacity(0.04))
            .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 18).stroke(on ? Theme.gold : .white.opacity(0.1), lineWidth: 1.5))
        }
    }

    private func subscribe() async {
        working = true
        defer { working = false }
        _ = try? await APIClient.shared.subscribe(plan: plan)
        await model.refreshUser()
        model.back()
    }
}

extension AppModel {
    /// Fallback perks if the meta endpoint is unreachable (offline preview).
    var perksFallback: [PlusPerk] {
        [.init(icon: "spark", title: "每日无限次 AI 咨询", desc: "免费版每天 3 次，Plus 想问就问"),
         .init(icon: "bolt", title: "更快的 AI 响应", desc: "优先算力，分析快人一步"),
         .init(icon: "crown", title: "更换 App 图标", desc: "9 款主题图标，点亮你的桌面")]
    }
}
