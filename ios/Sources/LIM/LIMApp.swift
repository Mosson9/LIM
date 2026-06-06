import SwiftUI

/// LIM — Less is More. App entry point.
@main
struct LIMApp: App {
    @StateObject private var model = AppModel()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(model)
                .tint(Theme.indigo)
                .task { await model.bootstrap() }
        }
    }
}

/// Decides between the splash, the auth screen, and the routed app shell.
struct RootView: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        Group {
            if model.booting {
                SplashView()
            } else if !model.isAuthed {
                AuthView()
            } else {
                AppShell()
            }
        }
    }
}

/// Brand splash shown while the session is restored.
struct SplashView: View {
    var body: some View {
        ZStack {
            Theme.paper.ignoresSafeArea()
            VStack(spacing: 16) {
                RoundedRectangle(cornerRadius: 22, style: .continuous)
                    .fill(Theme.indigo)
                    .frame(width: 78, height: 78)
                    .overlay(Text("L").font(Theme.display(40)).foregroundColor(.white))
                Text("LESS IS MORE").font(Theme.sans(11, .bold)).tracking(2).foregroundColor(Theme.ink3)
            }
        }
    }
}

/// Routed shell: renders the top screen and overlays the tab bar on root tabs.
struct AppShell: View {
    @EnvironmentObject var model: AppModel

    var body: some View {
        ZStack(alignment: .bottom) {
            Theme.paper.ignoresSafeArea()

            screen(for: model.top)
                .id(screenID(model.top))
                .transition(.asymmetric(
                    insertion: .move(edge: .trailing).combined(with: .opacity),
                    removal: .opacity))

            if model.top.tab != nil {
                TabBar(active: model.top.tab!) { model.selectTab($0) }
            }
        }
        .animation(.easeInOut(duration: 0.28), value: screenID(model.top))
    }

    @ViewBuilder
    private func screen(for s: Screen) -> some View {
        switch s {
        case .onboarding:        OnboardingView()
        case .profileSetup:      ProfileSetupView()
        case .home:              HomeView()
        case .growth:            GrowthView()
        case .history:           HistoryView()
        case .me:                MeView()
        case .ask:               AskView()
        case .analyzing(let i):  AnalyzingView(input: i)
        case .result(let d):     ResultView(decision: d)
        case .record(let d):     RecordView(decision: d)
        case .saved(let d):      SavedView(decision: d)
        case .wishlist:          WishlistView()
        case .plus:              PlusView()
        case .appIcon:           AppIconView()
        }
    }

    /// Stable identity per screen so transitions fire on navigation.
    private func screenID(_ s: Screen) -> String {
        switch s {
        case .analyzing(let i):  return "analyzing-\(i.item)-\(i.price)"
        case .result(let d):     return "result-\(d.id)"
        case .record(let d):     return "record-\(d.id)"
        case .saved(let d):      return "saved-\(d.id)"
        default:                 return "\(s)"
        }
    }
}
