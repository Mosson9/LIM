import SwiftUI

/// Input collected on the Ask screen and carried into analysis.
struct AnalyzeInput: Equatable, Hashable {
    var item: String
    var price: Int
    var cat: String
    var reason: String
}

/// The in-app navigation stack — mirrors the prototype's screen stack.
/// Root tabs (home/growth/history/me) reset the stack; everything else pushes.
enum Screen: Equatable {
    case onboarding
    case profileSetup
    case home, growth, history, me     // tab roots
    case ask
    case analyzing(AnalyzeInput)
    case result(Decision)
    case record(Decision)
    case saved(Decision)
    case wishlist
    case plus
    case appIcon

    var tab: Tab? {
        switch self {
        case .home: return .home
        case .growth: return .growth
        case .history: return .history
        case .me: return .me
        default: return nil
        }
    }
}

enum Tab: String, CaseIterable { case home, growth, history, me }

/// Central app state: session, navigation, and cached catalogue/stat data.
/// Injected as an `@EnvironmentObject` and used by every screen.
@MainActor
final class AppModel: ObservableObject {
    // Session
    @Published var user: User?
    @Published var booting = true
    @Published var authError: String?

    // Navigation
    @Published var stack: [Screen] = [.home]

    // Cached data
    @Published var stats: Stats = .zero
    @Published var decisions: [Decision] = []
    @Published var wishlist: [WishlistItem] = []
    @Published var categories: [Category] = []
    @Published var skins: [Skin] = []
    @Published var dimensionMeta: [DimensionMeta] = []

    private let api = APIClient.shared
    private let tokenKey = "lim.token"

    var top: Screen { stack.last ?? .home }
    var isAuthed: Bool { user != nil }

    // MARK: Navigation

    func go(_ screen: Screen, replace: Bool = false) {
        if let _ = screen.tab {
            stack = [screen]                       // root tab → reset stack
        } else if replace {
            stack[stack.count - 1] = screen
        } else {
            stack.append(screen)
        }
    }

    func back() { if stack.count > 1 { stack.removeLast() } }

    func selectTab(_ tab: Tab) {
        switch tab {
        case .home: go(.home)
        case .growth: go(.growth)
        case .history: go(.history)
        case .me: go(.me)
        }
    }

    // MARK: Bootstrap

    func bootstrap() async {
        defer { booting = false }
        guard let token = UserDefaults.standard.string(forKey: tokenKey) else { return }
        await api.setToken(token)
        do {
            let me = try await api.me()
            user = me
            await loadCatalogue()
            await refreshAll()
            stack = [me.onboarded ? .home : .onboarding]
        } catch {
            await signOut()       // stale token
        }
    }

    // MARK: Auth

    func login(email: String, password: String) async {
        authError = nil
        do {
            let r = try await api.login(email: email, password: password)
            try await complete(auth: r)
        } catch { authError = readable(error) }
    }

    func register(email: String, password: String, name: String) async {
        authError = nil
        do {
            let r = try await api.register(email: email, password: password, name: name, profile: nil)
            try await complete(auth: r)
        } catch { authError = readable(error) }
    }

    private func complete(auth: AuthResponse) async throws {
        UserDefaults.standard.set(auth.token, forKey: tokenKey)
        await api.setToken(auth.token)
        user = auth.user
        await loadCatalogue()
        await refreshAll()
        stack = [auth.user.onboarded ? .home : .onboarding]
    }

    func signOut() async {
        UserDefaults.standard.removeObject(forKey: tokenKey)
        await api.setToken(nil)
        user = nil
        stack = [.home]
    }

    // MARK: Data

    func loadCatalogue() async {
        async let c = api.categories()
        async let s = api.skins()
        async let d = api.dimensions()
        categories = (try? await c) ?? categories
        skins = (try? await s) ?? skins
        dimensionMeta = (try? await d) ?? dimensionMeta
    }

    func refreshAll() async {
        async let st = api.stats()
        async let de = api.decisions()
        async let wl = api.wishlist()
        if let st = try? await st { stats = st }
        if let de = try? await de { decisions = de }
        if let wl = try? await wl { wishlist = wl }
    }

    func refreshUser() async { if let me = try? await api.me() { user = me } }

    // MARK: Catalogue lookups

    func category(_ id: String) -> Category {
        categories.first { $0.id == id }
            ?? Category(id: "other", label: "其他", color: "#8C8A82", soft: "#EEEDE7", icon: "grid", free: true, count: 0)
    }

    func dimensionQuestion(_ key: String) -> String {
        dimensionMeta.first { $0.key == key }?.q ?? ""
    }

    // MARK: Mutations used by the core flow

    /// Mark the analysed decision as resisted, then show the celebratory screen.
    func resist(_ decision: Decision) async {
        if let updated = try? await api.resolveDecision(id: decision.id, action: "resist") {
            await refreshAll()
            go(.saved(updated), replace: true)
        }
    }

    /// Mark the analysed decision as bought, then show the record screen.
    func buy(_ decision: Decision) async {
        if let updated = try? await api.resolveDecision(id: decision.id, action: "buy") {
            await refreshAll()
            go(.record(updated), replace: true)
        }
    }

    /// Park the decision in the 24h cooling-off wishlist.
    func park(_ decision: Decision) async {
        _ = try? await api.addWishlist(item: decision.item, price: decision.price,
                                       cat: decision.cat, impulse: decision.impulse,
                                       decisionId: decision.id)
        await refreshAll()
        go(.wishlist, replace: true)
    }

    func readable(_ error: Error) -> String {
        if let api = error as? APIError { return api.error }
        return "网络异常，请稍后再试"
    }
}
