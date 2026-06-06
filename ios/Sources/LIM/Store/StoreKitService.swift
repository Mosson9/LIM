import Foundation
import StoreKit

/// StoreKit 2 purchasing for LIM Plus.
///
/// Flow: load the auto-renewable products → `purchase()` → take the signed
/// transaction's `jwsRepresentation` → POST it to `/subscription/verify` so the
/// backend cryptographically validates it and grants the entitlement → `finish()`.
/// A background task also listens for renewals/restores and re-verifies them.
@MainActor
final class StoreKitService: ObservableObject {
    static let monthlyID = "app.lim.ios.plus.monthly"
    static let yearlyID = "app.lim.ios.plus.yearly"
    static let productIDs = [monthlyID, yearlyID]

    @Published var products: [Product] = []
    @Published var purchasing = false
    @Published var error: String?

    private var updatesTask: Task<Void, Never>?

    init() { updatesTask = listenForTransactions() }
    deinit { updatesTask?.cancel() }

    /// Maps our plan id ("month"/"year") to a StoreKit product id.
    static func productID(forPlan plan: String) -> String {
        plan == "year" ? yearlyID : monthlyID
    }

    func load() async {
        do {
            let fetched = try await Product.products(for: Self.productIDs)
            products = fetched.sorted { $0.price < $1.price }
        } catch {
            self.error = "无法加载内购商品"
        }
    }

    func product(for id: String) -> Product? { products.first { $0.id == id } }

    /// Purchase a product and verify the receipt with the backend.
    /// Returns true when the entitlement was granted.
    @discardableResult
    func purchase(_ product: Product) async -> Bool {
        purchasing = true
        defer { purchasing = false }
        do {
            let result = try await product.purchase()
            switch result {
            case .success(let verification):
                // Send Apple's signed JWS to our server for verification.
                _ = try await APIClient.shared.verifySubscription(jws: verification.jwsRepresentation)
                if case .verified(let transaction) = verification {
                    await transaction.finish()
                }
                return true
            case .userCancelled:
                return false
            case .pending:
                error = "购买待确认（家长同意 / 银行验证）"
                return false
            @unknown default:
                return false
            }
        } catch {
            self.error = readable(error)
            return false
        }
    }

    /// Restore: re-verify the current entitlements with the backend.
    @discardableResult
    func restore() async -> Bool {
        var restored = false
        for await result in Transaction.currentEntitlements {
            if (try? await APIClient.shared.verifySubscription(jws: result.jwsRepresentation)) != nil {
                restored = true
            }
        }
        if !restored { error = "没有可恢复的购买" }
        return restored
    }

    /// Watch for renewals and out-of-band transactions; re-verify each.
    private func listenForTransactions() -> Task<Void, Never> {
        Task.detached {
            for await update in Transaction.updates {
                _ = try? await APIClient.shared.verifySubscription(jws: update.jwsRepresentation)
                if case .verified(let transaction) = update {
                    await transaction.finish()
                }
            }
        }
    }

    private func readable(_ error: Error) -> String {
        if let api = error as? APIError { return api.error }
        return (error as NSError).localizedDescription
    }
}
