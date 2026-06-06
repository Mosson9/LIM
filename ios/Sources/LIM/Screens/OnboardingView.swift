import SwiftUI

/// 3-slide intro. Mirrors the prototype's onboarding copy and "leaf/target/
/// sprout" artwork.
struct OnboardingView: View {
    @EnvironmentObject var model: AppModel
    @State private var index = 0

    private struct Slide { let kicker, body, art: String; let title: String }
    private var slides: [Slide] {
        [
            Slide(kicker: L("onboarding.s1.kicker"), body: L("onboarding.s1.body"),
                  art: "leaf", title: L("onboarding.s1.title")),
            Slide(kicker: L("onboarding.s2.kicker"), body: L("onboarding.s2.body"),
                  art: "target", title: L("onboarding.s2.title")),
            Slide(kicker: L("onboarding.s3.kicker"), body: L("onboarding.s3.body"),
                  art: "sprout", title: L("onboarding.s3.title")),
        ]
    }

    var body: some View {
        let s = slides[index]
        let last = index == slides.count - 1
        VStack(alignment: .leading, spacing: 0) {
            HStack {
                Kicker(text: "LIM", color: Theme.indigo)
                Spacer()
                if !last {
                    Button(L("common.skip")) { model.go(.profileSetup) }
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

            Button(last ? L("common.start") : L("common.next")) {
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
