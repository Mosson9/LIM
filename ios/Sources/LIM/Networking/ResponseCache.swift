import Foundation
import CryptoKit

/// A tiny on-disk cache of raw GET response bodies, keyed by request path.
///
/// It lets the app show last-known data when the network is unavailable: the
/// API client writes successful GET bodies here and reads them back on a
/// transport failure. Stored in the Caches directory (the OS may evict it).
actor ResponseCache {
    static let shared = ResponseCache()

    private let dir: URL

    init() {
        let base = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask)[0]
        dir = base.appendingPathComponent("lim-response-cache", isDirectory: true)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
    }

    private func file(for key: String) -> URL {
        let digest = SHA256.hash(data: Data(key.utf8))
        let name = digest.map { String(format: "%02x", $0) }.joined()
        return dir.appendingPathComponent(name)
    }

    func set(_ key: String, _ data: Data) {
        try? data.write(to: file(for: key), options: .atomic)
    }

    func get(_ key: String) -> Data? {
        try? Data(contentsOf: file(for: key))
    }

    /// Clear all cached responses (e.g. on sign-out).
    func clear() {
        try? FileManager.default.removeItem(at: dir)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
    }
}
