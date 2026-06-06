import SwiftUI

/// "已记一笔" — a calm confirmation that a *needed* purchase was recorded.
struct RecordView: View {
    @EnvironmentObject var model: AppModel
    let decision: Decision

    var body: some View {
        let cat = model.category(decision.cat)
        VStack(spacing: 0) {
            Spacer()
            VStack(spacing: 0) {
                Icon(name: cat.icon, size: 36, color: Color(webHex: cat.color))
                    .frame(width: 78, height: 78)
                    .background(Color(webHex: cat.soft))
                    .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
                    .padding(.bottom, 24)
                Kicker(text: "已记一笔").padding(.bottom, 12)
                HStack(alignment: .top, spacing: 2) {
                    Text("¥").font(Theme.display(28)).foregroundColor(Theme.ink)
                    Text(Format.n(decision.price)).font(Theme.display(52)).foregroundColor(Theme.ink)
                }
                Text(decision.item).font(Theme.sans(16)).foregroundColor(Theme.ink2).padding(.top, 8)

                Card {
                    VStack(spacing: 0) {
                        infoRow("分类", value: cat.label)
                        Hairline()
                        infoRow("记入", value: "本月消费")
                        Hairline()
                        infoRow("本月已花", value: Format.yuan(model.stats.monthSpent), color: Theme.spent)
                    }
                }
                .padding(.top, 30)

                Text("买，也可以是清醒的选择。需要的东西，值得好好拥有。")
                    .font(Theme.sans(13)).foregroundColor(Theme.ink3)
                    .multilineTextAlignment(.center).lineSpacing(4)
                    .frame(maxWidth: 280).padding(.top, 18)
            }
            Spacer()
            Button("完成") { model.go(.home) }
                .buttonStyle(FilledButtonStyle(bg: Theme.ink))
                .padding(.bottom, 30)
        }
        .padding(.horizontal, 24)
        .background(Theme.paper.ignoresSafeArea())
    }

    private func infoRow(_ label: String, value: String, color: Color = Theme.ink) -> some View {
        HStack {
            Text(label).font(Theme.sans(14.5)).foregroundColor(Theme.ink2)
            Spacer()
            Text(value).font(Theme.sans(14.5, .semibold)).foregroundColor(color)
        }
        .padding(.vertical, 12)
    }
}
