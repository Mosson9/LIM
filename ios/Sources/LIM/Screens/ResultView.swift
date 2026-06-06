import SwiftUI

/// Step ③–④ — the verdict. Shows the impulse dial, LIM's message, the six-axis
/// radar, expandable scorecards, and the final actions. Decisions that are
/// already resolved (opened from history) hide the action bar.
struct ResultView: View {
    @EnvironmentObject var model: AppModel
    let decision: Decision

    @State private var openDim: String?
    @State private var working = false

    private var isPending: Bool { decision.status == .pending }
    private var verdictColor: Color {
        switch decision.verdict {
        case .buy: return Theme.saved
        case .pause: return Theme.warn
        case .resist: return Theme.danger
        }
    }
    private var verdictSoft: Color {
        switch decision.verdict {
        case .buy: return Theme.savedSoft
        case .pause: return Theme.warnSoft
        case .resist: return Theme.dangerSoft
        }
    }

    var body: some View {
        VStack(spacing: 0) {
            TopBar(title: "LIM 的建议", leading: isPending ? .close : .back)
            ScrollView {
                VStack(spacing: 0) {
                    HStack(spacing: 10) {
                        Text(decision.item).font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
                        Text("·").foregroundColor(Theme.ink3)
                        Text(Format.yuan(decision.price)).font(Theme.sans(15)).foregroundColor(Theme.ink2)
                    }
                    .padding(.top, 4).padding(.bottom, 6)

                    ImpulseDial(value: Double(decision.impulse), size: 210)
                    Kicker(text: "冲动指数").offset(y: -6)

                    // verdict pill
                    HStack(spacing: 8) {
                        Icon(name: decision.verdict == .buy ? "check" : decision.verdict == .pause ? "clock" : "leaf",
                             size: 19, color: verdictColor)
                        Text(decision.verdict.label).font(Theme.sans(16, .semibold)).foregroundColor(verdictColor)
                    }
                    .padding(.vertical, 12).padding(.horizontal, 20)
                    .background(verdictSoft).clipShape(Capsule())
                    .padding(.top, 18)

                    aiMessage.padding(.top, 22)
                    radarCard.padding(.top, 14)
                    scorecards.padding(.top, 14)

                    Color.clear.frame(height: isPending ? 200 : 40)
                }
                .padding(.horizontal, 22)
            }
        }
        .background(Theme.paper.ignoresSafeArea())
        .overlay(alignment: .bottom) { if isPending { actionBar } }
    }

    private var aiMessage: some View {
        Card(background: Theme.indigoTint) {
            VStack(alignment: .leading, spacing: 10) {
                HStack(spacing: 9) {
                    Icon(name: "spark", size: 15, color: .white)
                        .frame(width: 26, height: 26).background(Theme.indigo)
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                    Text("LIM").font(Theme.sans(13.5, .bold)).foregroundColor(Theme.indigoInk)
                }
                Text(decision.message).font(Theme.serif(16.5)).foregroundColor(Theme.indigoInk).lineSpacing(6)
            }
        }
    }

    private var radarCard: some View {
        Card {
            VStack(spacing: 4) {
                HStack {
                    Text("六维分析").font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
                    Spacer()
                    Text("分数越高越值得").font(Theme.sans(12)).foregroundColor(Theme.ink3)
                }
                RadarChart(data: decision.dims.ordered().map { ($0.label, $0.value) }, size: 250)
            }
        }
    }

    private var scorecards: some View {
        VStack(spacing: 10) {
            ForEach(decision.dims.ordered(), id: \.key) { dim in
                let good = dim.value >= 4, bad = dim.value <= 2
                let col = good ? Theme.saved : bad ? Theme.danger : Theme.warn
                let isOpen = openDim == dim.key
                Card(padding: 15) {
                    VStack(alignment: .leading, spacing: 0) {
                        HStack {
                            VStack(alignment: .leading, spacing: 8) {
                                HStack(spacing: 8) {
                                    Text(dim.label).font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
                                    Text(String(format: "%.1f", dim.value))
                                        .font(Theme.sans(13, .bold)).foregroundColor(col)
                                }
                                GeometryReader { geo in
                                    ZStack(alignment: .leading) {
                                        Capsule().fill(Theme.paper2).frame(height: 5)
                                        Capsule().fill(col)
                                            .frame(width: geo.size.width * dim.value / 5, height: 5)
                                    }
                                }
                                .frame(height: 5)
                            }
                            Icon(name: "chevD", size: 18, color: Theme.ink3)
                                .rotationEffect(.degrees(isOpen ? 180 : 0))
                        }
                        if isOpen {
                            Text(model.dimensionQuestion(dim.key))
                                .font(Theme.sans(13.5)).foregroundColor(Theme.ink2)
                                .lineSpacing(4).padding(.top, 12)
                        }
                    }
                }
                .onTapGesture { withAnimation { openDim = isOpen ? nil : dim.key } }
            }
        }
    }

    private var actionBar: some View {
        VStack(spacing: 10) {
            if decision.verdict == .buy {
                Button { Task { await act { await model.buy(decision) } } } label: {
                    actionLabel("wallet", "买了，记一笔")
                }.buttonStyle(FilledButtonStyle(bg: Theme.saved))
                Button { Task { await act { await model.park(decision) } } } label: { Text("先放进心愿单") }
                    .buttonStyle(GhostButtonStyle())
            } else {
                Button { Task { await act { await model.resist(decision) } } } label: {
                    actionLabel("leaf", "我忍住了，省下 \(Format.yuan(decision.price))")
                }.buttonStyle(FilledButtonStyle(bg: Theme.saved))
                Button { Task { await act { await model.park(decision) } } } label: {
                    actionLabel("clock", "给它 24 小时冷静期")
                }.buttonStyle(GhostButtonStyle())
                Button { Task { await act { await model.buy(decision) } } } label: {
                    Text("还是买了 ›").font(Theme.sans(13.5)).foregroundColor(Theme.ink3)
                }
            }
        }
        .padding(.horizontal, 22).padding(.top, 14).padding(.bottom, 30)
        .background(LinearGradient(colors: [Theme.paper.opacity(0), Theme.paper], startPoint: .top, endPoint: .center))
        .disabled(working)
    }

    private func actionLabel(_ icon: String, _ text: String) -> some View {
        HStack(spacing: 8) { Icon(name: icon, size: 18, color: .white); Text(text) }
    }

    private func act(_ work: @escaping () async -> Void) async {
        working = true; await work(); working = false
    }
}
