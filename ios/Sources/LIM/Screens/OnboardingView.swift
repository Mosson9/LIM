import SwiftUI

/// 3-slide intro. Mirrors the prototype's onboarding copy and "leaf/target/
/// sprout" artwork.
struct OnboardingView: View {
    @EnvironmentObject var model: AppModel
    @State private var index = 0

    private struct Slide { let kicker, body, art: String; let title: String }
    private let slides = [
        Slide(kicker: "LESS IS MORE",
              body: "我们被太多「想要」包围。LIM 不卖东西，只在你心动的那一刻，陪你停下来想一想。",
              art: "leaf", title: "买之前，\n先问问自己。"),
        Slide(kicker: "六个角度",
              body: "从需求、替代、情感、长期价值、经济到环境——LIM 用六个角度，帮你看清这次心动的真相。",
              art: "target", title: "不评判，\n只是一起想清楚。"),
        Slide(kicker: "看得见的克制",
              body: "没买下的东西，会变成实实在在省下的钱，和一棵慢慢长大的树。少即是多。",
              art: "sprout", title: "每一次忍住，\n都在为你积攒。"),
    ]

    var body: some View {
        let s = slides[index]
        let last = index == slides.count - 1
        VStack(alignment: .leading, spacing: 0) {
            HStack {
                Kicker(text: "LIM", color: Theme.indigo)
                Spacer()
                if !last {
                    Button("跳过") { model.go(.profileSetup) }
                        .font(Theme.sans(13)).foregroundColor(Theme.ink2)
                }
            }
            .padding(.top, 14)

            Spacer()

            VStack(alignment: .leading, spacing: 0) {
                RoundedRectangle(cornerRadius: 36, style: .continuous)
                    .fill(Theme.surface)
                    .frame(width: 120, height: 120)
                    .overlay(Icon(name: s.art, size: 56, color: Theme.saved))
                    .shadow(color: Theme.ink.opacity(0.07), radius: 30, x: 0, y: 10)
                    .padding(.bottom, 38)
                Kicker(text: s.kicker)
                    .padding(.bottom, 18)
                Text(s.title)
                    .font(Theme.display(40)).foregroundColor(Theme.ink)
                    .fixedSize(horizontal: false, vertical: true)
                    .padding(.bottom, 20)
                Text(s.body)
                    .font(Theme.sans(16.5)).foregroundColor(Theme.ink2)
                    .lineSpacing(6).frame(maxWidth: 320, alignment: .leading)
            }
            .id(index)
            .transition(.opacity.combined(with: .move(edge: .trailing)))

            Spacer()

            HStack(spacing: 7) {
                ForEach(0..<slides.count, id: \.self) { i in
                    Capsule().fill(i == index ? Theme.indigo : Theme.ink4)
                        .frame(width: i == index ? 20 : 7, height: 7)
                }
            }
            .padding(.bottom, 26)

            Button(last ? "开始吧" : "下一步") {
                if last { model.go(.profileSetup) }
                else { withAnimation { index += 1 } }
            }
            .buttonStyle(FilledButtonStyle())
            .padding(.bottom, 24)
        }
        .padding(.horizontal, 24)
        .padding(.top, 50)
        .background(Theme.paper.ignoresSafeArea())
    }
}
