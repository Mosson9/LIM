import SwiftUI

/// 我的 tab — profile card, Plus banner, settings groups, and sign-out.
struct MeView: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        let u = model.user
        ScrollView {
            VStack(spacing: 14) {
                HStack { Kicker(text: "我的"); Spacer() }.padding(.top, 58)

                // profile card
                Card(padding: 20) {
                    HStack(spacing: 16) {
                        RoundedRectangle(cornerRadius: 20, style: .continuous)
                            .fill(Theme.indigo).frame(width: 60, height: 60)
                            .overlay(Text(String(u?.name.prefix(1) ?? "L"))
                                .font(Theme.display(26)).foregroundColor(.white))
                        VStack(alignment: .leading, spacing: 3) {
                            Text(u?.name ?? "LIM 用户").font(Theme.sans(19, .bold)).foregroundColor(Theme.ink)
                            Text("已省 \(Format.yuan(model.stats.totalSaved)) · \(u?.isPlus == true ? "Plus 会员" : "免费版")")
                                .font(Theme.sans(13)).foregroundColor(Theme.ink3)
                        }
                        Spacer()
                    }
                }

                // Plus banner
                Button { model.go(.plus) } label: {
                    Card(padding: 18, background: Theme.indigo) {
                        HStack(spacing: 14) {
                            Icon(name: "crown", size: 28, color: Theme.gold)
                            VStack(alignment: .leading, spacing: 2) {
                                Text("升级 LIM Plus").font(Theme.sans(16, .bold)).foregroundColor(.white)
                                Text("无限咨询 · 更快 AI · 专属皮肤")
                                    .font(Theme.sans(12.5)).foregroundColor(.white.opacity(0.7))
                            }
                            Spacer()
                            Text("查看").font(Theme.sans(12.5, .bold)).foregroundColor(Theme.ink)
                                .padding(.vertical, 8).padding(.horizontal, 14)
                                .background(Theme.gold).clipShape(Capsule())
                        }
                    }
                }
                .buttonStyle(.plain)

                // groups
                group("我的画像", rows: [
                    .init(icon: "user", color: Theme.indigo, title: "基本情况",
                          detail: "月入 \(Format.n(u?.profile.monthlyIncome ?? 0)) · \(u?.profile.members ?? 0)口人",
                          action: { model.go(.profileSetup) }),
                    .init(icon: "target", color: Theme.saved, title: "本月可支配预算",
                          detail: Format.yuan(u?.profile.monthBudget ?? 0), action: nil),
                ])
                group("个性化", rows: [
                    .init(icon: "crown", color: Theme.gold, title: "更换 App 图标",
                          detail: u?.appIcon ?? "classic", action: { model.go(.appIcon) }),
                    .init(icon: "bell", color: Theme.warn, title: "提醒与冷静期通知", detail: "已开启", action: nil),
                ])
                group("关于", rows: [
                    .init(icon: "info", color: Theme.ink2, title: "关于 LIM", detail: "Less is More", action: nil),
                    .init(icon: "refresh", color: Theme.ink2, title: "重新观看引导", detail: nil,
                          action: { model.go(.onboarding) }),
                ])

                Button { Task { await model.signOut() } } label: {
                    Text("退出登录").font(Theme.sans(15)).foregroundColor(Theme.danger)
                        .frame(maxWidth: .infinity).padding(.vertical, 14)
                }

                Text("LIM · Less is More · v1.0").font(Theme.sans(12)).foregroundColor(Theme.ink3)
                Color.clear.frame(height: 100)
            }
            .padding(.horizontal, 22)
        }
        .background(Theme.paper.ignoresSafeArea())
    }

    private struct Row: Identifiable {
        let icon: String; let color: Color; let title: String; let detail: String?
        let action: (() -> Void)?
        var id: String { title }
    }

    private func group(_ header: String, rows: [Row]) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Kicker(text: header).padding(.leading, 4)
            Card(padding: 4) {
                VStack(spacing: 0) {
                    ForEach(rows) { r in
                        Button { r.action?() } label: {
                            HStack(spacing: 12) {
                                Icon(name: r.icon, size: 19, color: r.color)
                                    .frame(width: 36, height: 36)
                                    .background(Theme.paper2)
                                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                                Text(r.title).font(Theme.sans(15.5, .medium)).foregroundColor(Theme.ink)
                                Spacer()
                                if let d = r.detail {
                                    Text(d).font(Theme.sans(13.5)).foregroundColor(Theme.ink3)
                                }
                                if r.action != nil { Icon(name: "chevR", size: 16, color: Theme.ink4) }
                            }
                            .padding(.horizontal, 12).padding(.vertical, 11)
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        .disabled(r.action == nil)
                        if r.id != rows.last?.id { Hairline().padding(.leading, 60) }
                    }
                }
            }
        }
    }
}
