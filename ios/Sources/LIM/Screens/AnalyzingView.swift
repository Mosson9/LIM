import SwiftUI

/// The "LIM 正在思考" interstitial. Steps through the six dimensions while the
/// real analysis request runs, then replaces itself with the result.
struct AnalyzingView: View {
    @EnvironmentObject var model: AppModel
    let input: AnalyzeInput

    @State private var idx = 0
    @State private var spin = false
    @State private var result: Decision?
    @State private var failed: String?

    private let dims = ["需求", "替代", "情感", "长期价值", "经济", "环境"]

    var body: some View {
        VStack(spacing: 0) {
            Spacer()
            // spinner
            ZStack {
                Circle().stroke(Theme.indigoSoft, lineWidth: 2).frame(width: 96, height: 96)
                Circle().trim(from: 0, to: 0.25)
                    .stroke(Theme.indigo, style: .init(lineWidth: 2, lineCap: .round))
                    .frame(width: 96, height: 96)
                    .rotationEffect(.degrees(spin ? 360 : 0))
                    .animation(.linear(duration: 1).repeatForever(autoreverses: false), value: spin)
                Icon(name: "spark", size: 34, color: Theme.indigo)
            }
            .padding(.bottom, 40)

            Kicker(text: "LIM 正在思考").padding(.bottom, 10)
            Text("关于「\(input.item)」，\n我从六个角度看看…")
                .font(Theme.serif(22)).foregroundColor(Theme.ink)
                .multilineTextAlignment(.center).lineSpacing(5)
                .padding(.bottom, 38).padding(.horizontal, 30)

            VStack(spacing: 12) {
                ForEach(dims.indices, id: \.self) { i in
                    HStack(spacing: 12) {
                        ZStack {
                            Circle().fill(i < idx ? Theme.saved : Theme.paper2)
                                .frame(width: 24, height: 24)
                            if i < idx { Icon(name: "check", size: 14, color: .white, weight: .bold) }
                            else { Text("\(i+1)").font(Theme.sans(11, .bold)).foregroundColor(Theme.ink3) }
                        }
                        Text("\(dims[i])分析").font(Theme.sans(15, .medium))
                            .foregroundColor(i < idx ? Theme.ink : Theme.ink2)
                        Spacer()
                        if i == idx { Text("分析中…").font(Theme.sans(12)).foregroundColor(Theme.ink3) }
                    }
                    .opacity(i <= idx ? 1 : 0.32)
                }
            }
            .frame(maxWidth: 300)

            if let failed {
                Text(failed).font(Theme.sans(13)).foregroundColor(Theme.danger).padding(.top, 24)
                Button("返回") { model.back() }.buttonStyle(GhostButtonStyle()).padding(.horizontal, 40).padding(.top, 12)
            }
            Spacer()
        }
        .padding(.horizontal, 22)
        .background(Theme.paper.ignoresSafeArea())
        .task { await run() }
    }

    private func run() async {
        spin = true
        // Fire the real request and a paced six-step animation concurrently.
        async let analysis = APIClient.shared.analyze(
            item: input.item, price: input.price, cat: input.cat, reason: input.reason)

        for i in 1...dims.count {
            try? await Task.sleep(nanoseconds: 520_000_000)
            withAnimation { idx = i }
        }

        do {
            let decision = try await analysis
            try? await Task.sleep(nanoseconds: 300_000_000)
            await model.refreshUser()
            model.go(.result(decision), replace: true)
        } catch {
            failed = model.readable(error)
        }
    }
}
