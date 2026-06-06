import XCTest
@testable import LIM

/// Pure-logic tests: number formatting, dimension ordering, and the in-app
/// navigation stack. None of these touch the network.
@MainActor
final class AppLogicTests: XCTestCase {

    func testCurrencyFormatting() {
        XCTAssertEqual(Format.n(18000), "18,000")
        XCTAssertEqual(Format.yuan(2299), "¥2,299")
        XCTAssertEqual(Format.n(0), "0")
    }

    func testDimensionsOrdering() {
        let d = Dimensions(need: 3, alt: 1, emo: 1.5, value: 2.5, money: 2, env: 3)
        let ordered = d.ordered()
        // Tuple elements need closures here — key paths don't index tuples.
        XCTAssertEqual(ordered.map { $0.label }, ["需求", "替代", "情感", "长期价值", "经济", "环境"])
        XCTAssertEqual(ordered.map { $0.key }, ["need", "alt", "emo", "value", "money", "env"])
        XCTAssertEqual(ordered[1].value, 1, accuracy: 0.001) // alt
    }

    func testVerdictLabels() {
        XCTAssertEqual(Verdict.buy.label, "可以拥有它")
        XCTAssertEqual(Verdict.pause.label, "再想想 / 冷静一下")
        XCTAssertEqual(Verdict.resist.label, "也许不必买")
    }

    func testNavigationPushAndBack() {
        let model = AppModel()
        XCTAssertEqual(model.top, .home)

        model.go(.ask)                    // non-tab → push
        XCTAssertEqual(model.stack.count, 2)
        XCTAssertEqual(model.top, .ask)

        model.go(.wishlist)               // push again
        XCTAssertEqual(model.stack.count, 3)

        model.back()                      // pop
        XCTAssertEqual(model.top, .ask)
        model.back()
        XCTAssertEqual(model.top, .home)
        model.back()                      // never empties the stack
        XCTAssertEqual(model.stack.count, 1)
    }

    func testTabSelectionResetsStack() {
        let model = AppModel()
        model.go(.ask)
        model.go(.plus)
        XCTAssertEqual(model.stack.count, 3)

        model.selectTab(.growth)          // root tab → reset
        XCTAssertEqual(model.stack, [.growth])
        XCTAssertEqual(model.top.tab, .growth)
    }

    func testReplaceTopScreen() {
        let model = AppModel()
        model.go(.ask)
        model.go(.wishlist, replace: true) // replace, not push
        XCTAssertEqual(model.stack.count, 2)
        XCTAssertEqual(model.top, .wishlist)
    }

    func testCategoryFallback() {
        let model = AppModel()
        // Unknown category falls back to the "other" placeholder.
        let c = model.category("does-not-exist")
        XCTAssertEqual(c.id, "other")
    }

    func testStoreKitPlanMapping() {
        XCTAssertEqual(StoreKitService.productID(forPlan: "year"), StoreKitService.yearlyID)
        XCTAssertEqual(StoreKitService.productID(forPlan: "month"), StoreKitService.monthlyID)
    }
}
