import Foundation

/// A thin async/await wrapper over the LIM REST API.
///
/// Configure `baseURL` for your environment (see `LIMConfig`). The token is set
/// after login/register and sent as a Bearer header on every authed call.
actor APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private var token: String?
    private let session: URLSession

    init(baseURL: URL = LIMConfig.apiBaseURL) {
        self.baseURL = baseURL
        let cfg = URLSessionConfiguration.default
        cfg.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: cfg)
    }

    func setToken(_ t: String?) { token = t }
    func currentToken() -> String? { token }

    // MARK: - Encoding / decoding

    private static let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        d.dateDecodingStrategy = .custom { dec in
            let c = try dec.singleValueContainer()
            let s = try c.decode(String.self)
            if let date = ISO8601DateFormatter.lim.date(from: s) { return date }
            if let date = ISO8601DateFormatter.limPlain.date(from: s) { return date }
            throw DecodingError.dataCorruptedError(in: c, debugDescription: "bad date: \(s)")
        }
        return d
    }()

    private static let encoder: JSONEncoder = {
        let e = JSONEncoder()
        e.keyEncodingStrategy = .convertToSnakeCase
        return e
    }()

    // MARK: - Core request

    private func request<T: Decodable>(
        _ method: String,
        _ path: String,
        body: Encodable? = nil,
        authed: Bool = true
    ) async throws -> T {
        // Build the URL by string so query components (?filter=…) survive — that
        // appendingPathComponent would percent-escape.
        guard let url = URL(string: baseURL.absoluteString + "/" + path) else {
            throw APIError(error: "无效的请求地址")
        }
        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if authed, let token { req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        if let body { req.httpBody = try Self.encoder.encode(AnyEncodable(body)) }

        let (data, resp) = try await session.data(for: req)
        guard let http = resp as? HTTPURLResponse else { throw URLError(.badServerResponse) }
        guard (200..<300).contains(http.statusCode) else {
            if let apiErr = try? Self.decoder.decode(APIError.self, from: data) { throw apiErr }
            throw APIError(error: "请求失败 (\(http.statusCode))")
        }
        if T.self == EmptyResponse.self { return EmptyResponse() as! T }
        return try Self.decoder.decode(T.self, from: data)
    }

    // MARK: - Auth

    func register(email: String, password: String, name: String, profile: Profile?) async throws -> AuthResponse {
        struct Req: Encodable { let email, password, name: String; let profile: Profile? }
        let r: AuthResponse = try await request("POST", "api/v1/auth/register",
            body: Req(email: email, password: password, name: name, profile: profile), authed: false)
        token = r.token
        return r
    }

    func login(email: String, password: String) async throws -> AuthResponse {
        struct Req: Encodable { let email, password: String }
        let r: AuthResponse = try await request("POST", "api/v1/auth/login",
            body: Req(email: email, password: password), authed: false)
        token = r.token
        return r
    }

    func me() async throws -> User { try await request("GET", "api/v1/me") }

    // MARK: - Profile

    func updateProfile(_ p: Profile) async throws -> User {
        try await request("PUT", "api/v1/me/profile", body: p)
    }

    func onboard() async throws -> User { try await request("POST", "api/v1/me/onboard", body: EmptyBody()) }

    func setAppIcon(_ id: String) async throws -> User {
        struct Req: Encodable { let appIcon: String }
        return try await request("PUT", "api/v1/me/app-icon", body: Req(appIcon: id))
    }

    // MARK: - Meta / catalogue

    func categories() async throws -> [Category] { try await request("GET", "api/v1/meta/categories", authed: false) }
    func skins() async throws -> [Skin] { try await request("GET", "api/v1/meta/skins", authed: false) }
    func plans() async throws -> [PlanOption] { try await request("GET", "api/v1/meta/plans", authed: false) }
    func perks() async throws -> [PlusPerk] { try await request("GET", "api/v1/meta/perks", authed: false) }
    func dimensions() async throws -> [DimensionMeta] { try await request("GET", "api/v1/meta/dimensions", authed: false) }

    // MARK: - Analyze & decisions

    func analyze(item: String, price: Int, cat: String, reason: String) async throws -> Decision {
        struct Req: Encodable { let item: String; let price: Int; let cat, reason: String }
        return try await request("POST", "api/v1/analyze", body: Req(item: item, price: price, cat: cat, reason: reason))
    }

    func decisions(filter: String = "all") async throws -> [Decision] {
        try await request("GET", "api/v1/decisions?filter=\(filter)")
    }

    func resolveDecision(id: String, action: String) async throws -> Decision {
        struct Req: Encodable { let action: String }
        return try await request("POST", "api/v1/decisions/\(id)/resolve", body: Req(action: action))
    }

    // MARK: - Wishlist

    func wishlist() async throws -> [WishlistItem] { try await request("GET", "api/v1/wishlist") }

    func addWishlist(item: String, price: Int, cat: String, impulse: Int, decisionId: String?) async throws -> WishlistItem {
        struct Req: Encodable { let item: String; let price: Int; let cat: String; let impulse: Int; let decisionId: String? }
        return try await request("POST", "api/v1/wishlist",
            body: Req(item: item, price: price, cat: cat, impulse: impulse, decisionId: decisionId))
    }

    func resolveWishlist(id: String, action: String) async throws {
        struct Req: Encodable { let action: String }
        let _: EmptyResponse = try await request("POST", "api/v1/wishlist/\(id)/resolve", body: Req(action: action))
    }

    func deleteWishlist(id: String) async throws {
        let _: EmptyResponse = try await request("DELETE", "api/v1/wishlist/\(id)")
    }

    // MARK: - Stats & subscription

    func stats() async throws -> Stats { try await request("GET", "api/v1/stats") }
    func subscription() async throws -> Subscription { try await request("GET", "api/v1/subscription") }

    func subscribe(plan: String) async throws -> Subscription {
        struct Req: Encodable { let plan: String }
        return try await request("POST", "api/v1/subscription/subscribe", body: Req(plan: plan))
    }
}

// MARK: - Encoding helpers

/// Marker for endpoints that return no useful body.
struct EmptyResponse: Decodable {}
private struct EmptyBody: Encodable {}

/// Type-erasing wrapper so `Encodable` existentials can be encoded.
private struct AnyEncodable: Encodable {
    let value: Encodable
    init(_ v: Encodable) { value = v }
    func encode(to encoder: Encoder) throws { try value.encode(to: encoder) }
}

extension ISO8601DateFormatter {
    static let lim: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()
    static let limPlain: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()
}
