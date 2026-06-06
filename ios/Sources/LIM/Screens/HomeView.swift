import SwiftUI

/// Home tab: this-month savings hero, the primary "should I buy?" CTA, a
/// cooling-off reminder, and recent decisions.
struct HomeView: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                header
                hero
                cta
                if let w = model.wishlist.first { coolingReminder(w) }
                recentDecisions
                Color.clear.frame(height: 110)   // tab-bar inset
            }
            .padding(.horizontal, 22)
        }
        .background(Theme.paper.ignoresSafeArea())
        .refreshable { await model.refreshAll() }
    }

    private var header: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("你好，\(model.user?.name ?? "朋友")")
                    .font(Theme.sans(13)).foregroundColor(Theme.ink3)
                Text("少一点，反而更好").font(Theme.sans(19, .semibold)).foregroundColor(Theme.ink)
            }
            Spacer()
            Button { model.selectTab(.me) } label: {
                Icon(name: "user", size: 20, color: Theme.ink2)
                    .frame(width: 40, height: 40)
                    .background(Theme.surface).clipShape(Circle())
                    .shadow(color: Theme.ink.opacity(0.05), radius: 6, y: 2)
            }
        }
        .padding(.top, 58).padding(.bottom, 4)
    }

    private var hero: some View {
        Card(padding: 24, background: Theme.surface) {
            ZStack(alignment: .topTrailing) {
                GrowthTree(stage: model.stats.treeStage, size: 130)
                    .opacity(0.9)
                    .frame(maxWidth: .infinity, alignment: .trailing)
                VStack(alignment: .leading, spacing: 0) {
                    Kicker(text: "本月已省下").padding(.bottom, 12)
                    HStack(alignment: .top, spacing: 2) {
                        Text("¥").font(Theme.display(32)).foregroundColor(Theme.savedDeep)
                        Text(Format.n(model.stats.monthSaved))
                            .font(Theme.display(60)).foregroundColor(Theme.savedDeep)
                    }
                    HStack(spacing: 8) {
                        Badge(text: "忍住 \(model.stats.resistCount) 次", fg: Theme.savedDeep,
                              bg: Theme.savedSoft, systemImage: "leaf.fill")
                        Badge(text: "连续 \(model.stats.streak) 天克制", fg: Theme.indigoInk,
                              bg: Theme.indigoSoft, systemImage: "bolt.fill")
                    }
                    .padding(.top, 18)
                }
            }
        }
        .padding(.top, 14)
    }

    private var cta: some View {
        VStack(spacing: 12) {
            Button { model.go(.ask) } label: {
                HStack(spacing: 8) {
                    Icon(name: "spark", size: 20, color: .white, weight: .semibold)
                    Text("我想买点东西，该不该买？")
                }
            }
            .buttonStyle(FilledButtonStyle())

            if let u = model.user {
                HStack(spacing: 6) {
                    Icon(name: "info", size: 13, color: Theme.ink3)
                    Text(u.isPlus ? "Plus 会员 · 无限次咨询"
                         : "今日还可咨询 \(max(0, 3 - u.aiUsedToday)) 次 · Plus 无限次")
                        .font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
                }
            }
        }
        .padding(.top, 16)
    }

    private func coolingReminder(_ w: WishlistItem) -> some View {
        VStack(spacing: 10) {
            SectionHeader(title: "冷静期 · 心愿单", actionTitle: "全部") { model.go(.wishlist) }
            Card {
                HStack(spacing: 14) {
                    Icon(name: "clock", size: 22, color: Color(webHex: model.category(w.cat).color))
                        .frame(width: 44, height: 44)
                        .background(Color(webHex: model.category(w.cat).soft))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 2) {
                        Text(w.item).font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink).lineLimit(1)
                        Text("\(w.remainText) · 到期再决定").font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
                    }
                    Spacer()
                    ProgressRing(progress: w.progress).frame(width: 36, height: 36)
                }
            }
            .onTapGesture { model.go(.wishlist) }
        }
        .padding(.top, 20)
    }

    private var recentDecisions: some View {
        VStack(spacing: 10) {
            SectionHeader(title: "最近的决定", actionTitle: "全部") { model.go(.history) }
            Card(padding: 4) {
                VStack(spacing: 0) {
                    let recent = Array(model.decisions.prefix(3))
                    if recent.isEmpty {
                        Text("还没有决定，点下方 ✦ 问问 LIM 吧")
                            .font(Theme.sans(13)).foregroundColor(Theme.ink3)
                            .frame(maxWidth: .infinity).padding(.vertical, 22)
                    }
                    ForEach(recent) { d in
                        DecisionRow(decision: d).onTapGesture { model.go(.result(d)) }
                        if d.id != recent.last?.id { Hairline().padding(.leading, 58) }
                    }
                }
            }
        }
        .padding(.top, 22)
    }
}

/// Compact row used in Home/History lists.
struct DecisionRow: View {
    let decision: Decision
    @EnvironmentObject var model: AppModel

    var body: some View {
        let cat = model.category(decision.cat)
        HStack(spacing: 12) {
            Icon(name: cat.icon, size: 20, color: Color(webHex: cat.color))
                .frame(width: 38, height: 38)
                .background(Color(webHex: cat.soft))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            VStack(alignment: .leading, spacing: 1) {
                Text(decision.item).font(Theme.sans(15, .medium)).foregroundColor(Theme.ink).lineLimit(1)
                Text(relativeDate(decision)).font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
            }
            Spacer()
            if decision.status == .resisted {
                Badge(text: "省 \(Format.yuan(decision.saved))", fg: Theme.savedDeep, bg: Theme.savedSoft)
            } else if decision.status == .bought {
                Badge(text: "买了", fg: Theme.ink2, bg: Theme.spentSoft)
            } else {
                Badge(text: "冷静中", fg: Theme.warn, bg: Theme.warnSoft)
            }
        }
        .padding(.horizontal, 12).padding(.vertical, 10)
        .contentShape(Rectangle())
    }

    private func relativeDate(_ d: Decision) -> String {
        guard let date = d.decidedAt ?? d.createdAt else { return "" }
        let f = RelativeDateTimeFormatter(); f.locale = Locale(identifier: "zh_CN")
        return f.localizedString(for: date, relativeTo: Date())
    }
}

/// Circular progress ring (cooling-off countdown).
struct ProgressRing: View {
    let progress: Double
    var color: Color = Theme.indigo
    var body: some View {
        ZStack {
            Circle().stroke(Theme.paper2, lineWidth: 3)
            Circle().trim(from: 0, to: progress)
                .stroke(color, style: .init(lineWidth: 3, lineCap: .round))
                .rotationEffect(.degrees(-90))
        }
    }
}
