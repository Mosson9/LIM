import Foundation

// MARK: - Enums

/// The AI's final recommendation.
enum Verdict: String, Codable, Hashable {
    case buy, pause, resist

    /// Pill label shown on the result screen.
    var label: String {
        switch self {
        case .buy:    return "可以拥有它"
        case .pause:  return "再想想 / 冷静一下"
        case .resist: return "也许不必买"
        }
    }
}

/// Where a decision ended up.
enum DecisionStatus: String, Codable, Hashable {
    case pending, resisted, bought, wishlist
}

/// Subscription tier.
enum Plan: String, Codable, Hashable {
    case free, month, year
}

// MARK: - Core entities (mirror backend/internal/models)

/// Six 1–5 sub-scores. Higher = the purchase is more justified on that axis.
struct Dimensions: Codable, Hashable {
    var need: Double
    var alt: Double
    var emo: Double
    var value: Double
    var money: Double
    var env: Double

    /// Ordered (key, label, value) for radar/scorecards.
    func ordered() -> [(key: String, label: String, value: Double)] {
        [("need", "需求", need), ("alt", "替代", alt), ("emo", "情感", emo),
         ("value", "长期价值", value), ("money", "经济", money), ("env", "环境", env)]
    }
}

struct Profile: Codable, Hashable {
    var monthlyIncome: Int
    var householdIncome: Int
    var members: Int
    var debt: Int
    var monthBudget: Int

    static let empty = Profile(monthlyIncome: 18000, householdIncome: 32000, members: 3, debt: 4200, monthBudget: 6000)
}

struct User: Codable, Identifiable, Hashable {
    let id: String
    var email: String
    var name: String
    var role: String
    var city: String?
    var createdAt: Date?
    var lastActiveAt: Date?
    var profile: Profile
    var plan: Plan
    var plusUntil: Date?
    var aiUsedToday: Int
    var appIcon: String
    var onboarded: Bool

    var isPlus: Bool {
        if plan == .free { return false }
        if let until = plusUntil { return until > Date() }
        return true
    }
}

struct Decision: Codable, Identifiable, Hashable {
    let id: String
    var userId: String
    var item: String
    var price: Int
    var cat: String
    var reason: String?
    var dims: Dimensions
    var impulse: Int
    var verdict: Verdict
    var message: String
    var note: String
    var saved: Int
    var status: DecisionStatus
    var createdAt: Date?
    var decidedAt: Date?
}

/// Wishlist item decorated with computed cooling-off fields from the server.
struct WishlistItem: Codable, Identifiable, Hashable {
    let id: String
    var userId: String
    var decisionId: String?
    var item: String
    var price: Int
    var cat: String
    var impulse: Int
    var addedAt: Date?
    var expiresAt: Date?
    var progress: Double
    var expired: Bool
    var remainSec: Int

    /// "剩 7 小时" / "冷静期已过"
    var remainText: String {
        if expired { return "冷静期已过" }
        let h = remainSec / 3600
        if h >= 1 { return "剩 \(h) 小时" }
        return "剩 \(max(1, remainSec / 60)) 分钟"
    }
}

// MARK: - Stats (derived server-side)

struct MonthBucket: Codable, Hashable, Identifiable {
    var label: String
    var spent: Int
    var saved: Int
    var id: String { label }
}

struct CatSaved: Codable, Hashable, Identifiable {
    var cat: String
    var label: String
    var color: String
    var saved: Int
    var id: String { cat }
}

struct Stats: Codable, Hashable {
    var monthSaved: Int
    var monthSpent: Int
    var totalSaved: Int
    var totalSpent: Int
    var resistCount: Int
    var buyCount: Int
    var streak: Int
    var treeStage: Int
    var restraintRate: Int
    var byMonth: [MonthBucket]
    var topCats: [CatSaved]

    static let zero = Stats(monthSaved: 0, monthSpent: 0, totalSaved: 0, totalSpent: 0,
                            resistCount: 0, buyCount: 0, streak: 0, treeStage: 0,
                            restraintRate: 0, byMonth: [], topCats: [])
}

// MARK: - Catalogue / meta

struct Category: Codable, Identifiable, Hashable {
    let id: String
    var label: String
    var color: String
    var soft: String
    var icon: String
    var free: Bool
    var count: Int
}

struct Skin: Codable, Identifiable, Hashable {
    let id: String
    var name: String
    var bg: String
    var dark: Bool
    var free: Bool
    var used: String?
}

struct PlanOption: Codable, Identifiable, Hashable {
    let id: String
    var name: String
    var price: Int
    var per: String
    var note: String
    var best: Bool
}

struct PlusPerk: Codable, Hashable, Identifiable {
    var icon: String
    var title: String
    var desc: String
    var id: String { title }
}

struct DimensionMeta: Codable, Hashable, Identifiable {
    var key: String
    var label: String
    var q: String
    var id: String { key }
}

// MARK: - Responses

struct AuthResponse: Codable {
    let token: String
    let user: User
}

struct Subscription: Codable {
    var plan: Plan
    var isPlus: Bool
    var plusUntil: Date?
    var plans: [PlanOption]
    var perks: [PlusPerk]
}

struct APIError: Codable, Error, LocalizedError {
    let error: String
    var errorDescription: String? { error }
}
