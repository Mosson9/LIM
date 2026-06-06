import Foundation

/// Build-time configuration. Override `LIM_API_BASE_URL` in the scheme's
/// environment / Info.plist for staging or production.
enum LIMConfig {
    static var apiBaseURL: URL {
        if let s = Bundle.main.object(forInfoDictionaryKey: "LIM_API_BASE_URL") as? String,
           let u = URL(string: s) {
            return u
        }
        if let s = ProcessInfo.processInfo.environment["LIM_API_BASE_URL"], let u = URL(string: s) {
            return u
        }
        // Default: local backend. Use your Mac's LAN IP when running on-device.
        return URL(string: "http://localhost:8080")!
    }
}

/// Number formatting helpers matching the prototype (`fmt` / `yuan`).
enum Format {
    private static let grouping: NumberFormatter = {
        let f = NumberFormatter()
        f.numberStyle = .decimal
        f.groupingSeparator = ","
        f.maximumFractionDigits = 0
        return f
    }()

    /// 18000 -> "18,000"
    static func n(_ value: Int) -> String { grouping.string(from: NSNumber(value: value)) ?? "\(value)" }

    /// 18000 -> "¥18,000"
    static func yuan(_ value: Int) -> String { "¥" + n(value) }
}
