import SwiftUI

/// 历史 tab — every "buy / don't buy" decision, filterable, with the saved total.
struct HistoryView: View {
    @EnvironmentObject var model: AppModel
    @State private var filter = "all"

    private var list: [Decision] {
        model.decisions.filter { d in
            switch filter {
            case "resist": return d.status == .resisted
            case "buy": return d.status == .bought
            default: return true
            }
        }
    }
    private var resisted: [Decision] { model.decisions.filter { $0.status == .resisted } }

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Color.clear.frame(width: 38)
                Spacer()
                Text("决策历史").font(Theme.sans(17, .semibold)).foregroundColor(Theme.ink)
                Spacer()
                Color.clear.frame(width: 38)
            }
            .padding(.top, 54).padding(.bottom, 8)

            ScrollView {
                VStack(spacing: 16) {
                    // summary
                    Card(padding: 16) {
                        HStack {
                            summary("\(resisted.count)", "次忍住")
                            Divider().frame(height: 30)
                            summary(Format.yuan(resisted.reduce(0) { $0 + $1.saved }), "累计省下")
                        }
                    }

                    // segmented filter
                    HStack(spacing: 0) {
                        ForEach([("all", "全部"), ("resist", "忍住了"), ("buy", "买了")], id: \.0) { key, label in
                            Button { filter = key } label: {
                                Text(label).font(Theme.sans(13.5, filter == key ? .semibold : .regular))
                                    .foregroundColor(filter == key ? Theme.ink : Theme.ink3)
                                    .frame(maxWidth: .infinity).padding(.vertical, 8)
                                    .background(filter == key ? Theme.surface : .clear)
                                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                            }
                        }
                    }
                    .padding(4).background(Theme.paper2)
                    .clipShape(RoundedRectangle(cornerRadius: 13, style: .continuous))

                    // timeline
                    ForEach(list) { d in historyCard(d) }
                    Color.clear.frame(height: 110)
                }
                .padding(.horizontal, 22)
            }
        }
        .background(Theme.paper.ignoresSafeArea())
        .refreshable { await model.refreshAll() }
    }

    private func summary(_ value: String, _ label: String) -> some View {
        VStack(spacing: 2) {
            Text(value).font(Theme.display(24)).foregroundColor(Theme.savedDeep)
            Text(label).font(Theme.sans(12)).foregroundColor(Theme.ink3)
        }
        .frame(maxWidth: .infinity)
    }

    private func historyCard(_ d: Decision) -> some View {
        let cat = model.category(d.cat)
        let resist = d.status == .resisted
        return Card(padding: 16) {
            VStack(alignment: .leading, spacing: 11) {
                HStack(spacing: 13) {
                    Icon(name: cat.icon, size: 21, color: Color(webHex: cat.color))
                        .frame(width: 44, height: 44)
                        .background(Color(webHex: cat.soft))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 4) {
                        HStack {
                            Text(d.item).font(Theme.sans(15.5, .semibold)).foregroundColor(Theme.ink).lineLimit(1)
                            Spacer()
                            if resist {
                                Badge(text: "省 \(Format.yuan(d.saved))", fg: Theme.savedDeep, bg: Theme.savedSoft)
                            } else if d.status == .bought {
                                Badge(text: "花 \(Format.yuan(d.price))", fg: Theme.ink2, bg: Theme.spentSoft)
                            } else {
                                Badge(text: "冷静中", fg: Theme.warn, bg: Theme.warnSoft)
                            }
                        }
                        HStack(spacing: 8) {
                            Text("冲动 \(d.impulse)")
                                .font(Theme.sans(12.5, .semibold))
                                .foregroundColor(d.impulse >= 62 ? Theme.danger : d.impulse >= 45 ? Theme.warn : Theme.saved)
                        }
                    }
                }
                if !d.note.isEmpty {
                    Text(d.note).font(Theme.sans(13)).foregroundColor(Theme.ink2)
                        .lineSpacing(3).padding(.leading, 57)
                }
            }
        }
        .onTapGesture { model.go(.result(d)) }
    }
}
