import SwiftUI

/// 4-step financial intake (收入 / 家庭收入 / 家庭成员 / 负债) — step ② of the
/// mind-map. Saves the profile, marks the user onboarded, and enters the app.
struct ProfileSetupView: View {
    @EnvironmentObject var model: AppModel
    @State private var step = 0
    @State private var inc = 18000
    @State private var hh = 32000
    @State private var members = 3
    @State private var debt = 4200
    @State private var saving = false

    private struct StepDef {
        let kicker, title, sub, unit: String
        let presets: [Int]
        let chips: [Int]?
        let value: Binding<Int>
    }

    private var steps: [StepDef] {
        [
            .init(kicker: "第 1 步 / 4", title: "你的月收入大约是？",
                  sub: "只用于让建议更贴合你，数据只留在你的设备里。", unit: "元/月",
                  presets: [8000, 12000, 18000, 25000, 40000], chips: nil, value: $inc),
            .init(kicker: "第 2 步 / 4", title: "家庭月收入呢？",
                  sub: "帮助 LIM 理解你的整体经济节奏。", unit: "元/月",
                  presets: [15000, 25000, 35000, 50000, 80000], chips: nil, value: $hh),
            .init(kicker: "第 3 步 / 4", title: "家里有几口人？",
                  sub: "家庭规模会影响什么才算「值得」。", unit: "人",
                  presets: [], chips: [1, 2, 3, 4, 5], value: $members),
            .init(kicker: "第 4 步 / 4", title: "每月固定要还的钱？",
                  sub: "房贷、车贷、花呗… 有就填，没有填 0。", unit: "元/月",
                  presets: [0, 2000, 4000, 8000, 15000], chips: nil, value: $debt),
        ]
    }

    var body: some View {
        let c = steps[step]
        let last = step == steps.count - 1
        VStack(alignment: .leading, spacing: 0) {
            HStack {
                Button { step > 0 ? (step -= 1) : model.go(.onboarding) } label: {
                    Icon(name: "back", size: 22, color: Theme.ink)
                }
                Spacer()
            }
            .padding(.top, 54).padding(.bottom, 10)

            VStack(alignment: .leading, spacing: 0) {
                Kicker(text: c.kicker, color: Theme.indigo).padding(.bottom, 14)
                Text(c.title).font(Theme.display(30)).foregroundColor(Theme.ink).padding(.bottom, 12)
                Text(c.sub).font(Theme.sans(14.5)).foregroundColor(Theme.ink2).lineSpacing(4).padding(.bottom, 32)

                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text(Format.n(c.value.wrappedValue)).font(Theme.display(52)).foregroundColor(Theme.ink)
                    Text(c.unit).font(Theme.sans(16)).foregroundColor(Theme.ink3)
                }
                Hairline().padding(.vertical, 20)

                if let chips = c.chips {
                    HStack(spacing: 10) {
                        ForEach(chips, id: \.self) { v in
                            Button { c.value.wrappedValue = v } label: {
                                Chip(label: "\(v)", selected: c.value.wrappedValue == v)
                                    .frame(maxWidth: .infinity)
                            }
                        }
                    }
                } else {
                    FlowChips(values: c.presets, selected: c.value.wrappedValue) { c.value.wrappedValue = $0 }
                }
            }
            .id(step)
            .transition(.opacity)

            Spacer()

            Button {
                if last { Task { await finish() } } else { withAnimation { step += 1 } }
            } label: {
                if saving { ProgressView().tint(.white) } else { Text(last ? "完成，进入 LIM" : "继续") }
            }
            .buttonStyle(FilledButtonStyle())
            .padding(.bottom, 30)
        }
        .padding(.horizontal, 22)
        .background(Theme.paper.ignoresSafeArea())
    }

    private func finish() async {
        saving = true
        defer { saving = false }
        let budget = max(1000, (inc - debt) / 3)
        let profile = Profile(monthlyIncome: inc, householdIncome: hh, members: members,
                              debt: debt, monthBudget: budget)
        _ = try? await APIClient.shared.updateProfile(profile)
        _ = try? await APIClient.shared.onboard()
        await model.refreshUser()
        await model.refreshAll()
        model.go(.home)
    }
}

/// Wrapping row of preset chips.
struct FlowChips: View {
    let values: [Int]
    let selected: Int
    let onTap: (Int) -> Void

    var body: some View {
        FlexWrap(spacing: 9, lineSpacing: 9) {
            ForEach(values, id: \.self) { v in
                Button { onTap(v) } label: {
                    Chip(label: Format.n(v), selected: selected == v)
                }
            }
        }
    }
}
