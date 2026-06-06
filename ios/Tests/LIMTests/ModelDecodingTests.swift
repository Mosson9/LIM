import XCTest
@testable import LIM

/// Verifies the Codable models decode the backend's snake_case JSON exactly as
/// `APIClient` does (convertFromSnakeCase + ISO8601 dates), including enums,
/// optionals, and computed properties.
final class ModelDecodingTests: XCTestCase {

    /// A decoder configured identically to APIClient's — reuses the same robust
    /// date parser so these tests exercise the real decoding path.
    private func decoder() -> JSONDecoder {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        d.dateDecodingStrategy = .custom { dec in
            let c = try dec.singleValueContainer()
            let s = try c.decode(String.self)
            if let date = LIMDate.parse(s) { return date }
            throw DecodingError.dataCorruptedError(in: c, debugDescription: "bad date: \(s)")
        }
        return d
    }

    /// Directly exercise the nanosecond-precision parsing fix.
    func testDateParsingHandlesNanoseconds() {
        XCTAssertNotNil(LIMDate.parse("2026-06-06T03:59:00.416404297Z"))
        XCTAssertNotNil(LIMDate.parse("2026-06-06T03:59:00.41Z"))
        XCTAssertNotNil(LIMDate.parse("2026-06-06T03:59:00Z"))
        XCTAssertNil(LIMDate.parse("not-a-date"))
    }

    func testDecodeUserAndEntitlement() throws {
        let json = """
        {
          "id":"U-20413","email":"lin@lim.app","name":"林一","role":"user","city":"上海",
          "created_at":"2025-04-07T03:58:44.92399277Z",
          "last_active_at":"2026-06-06T03:58:50.61Z",
          "profile":{"monthly_income":18000,"household_income":32000,"members":3,"debt":4200,"month_budget":6000},
          "plan":"year","plus_until":"2030-01-01T00:00:00Z",
          "ai_used_today":2,"app_icon":"classic","onboarded":true
        }
        """.data(using: .utf8)!
        let u = try decoder().decode(User.self, from: json)
        XCTAssertEqual(u.id, "U-20413")
        XCTAssertEqual(u.name, "林一")
        XCTAssertEqual(u.plan, .year)
        XCTAssertEqual(u.profile.monthBudget, 6000)
        XCTAssertEqual(u.aiUsedToday, 2)
        XCTAssertTrue(u.isPlus, "year plan with future plus_until should be Plus")
    }

    func testFreeUserIsNotPlus() throws {
        let json = """
        {"id":"U-1","email":"a@b.c","name":"A","role":"user","profile":
        {"monthly_income":0,"household_income":0,"members":1,"debt":0,"month_budget":1000},
        "plan":"free","ai_used_today":0,"app_icon":"classic","onboarded":false}
        """.data(using: .utf8)!
        let u = try decoder().decode(User.self, from: json)
        XCTAssertFalse(u.isPlus)
        XCTAssertNil(u.city)
        XCTAssertNil(u.plusUntil)
    }

    func testDecodeDecisionEnumsAndDims() throws {
        let json = """
        {
          "id":"D-88011","user_id":"U-20413","item":"索尼降噪耳机","price":2299,"cat":"digital",
          "reason":"直播打折","dims":{"need":3,"alt":1,"emo":1.5,"value":2.5,"money":2,"env":3},
          "impulse":72,"verdict":"resist","message":"…","note":"已有同类","saved":0,
          "status":"pending","created_at":"2026-06-06T03:59:00.41Z"
        }
        """.data(using: .utf8)!
        let d = try decoder().decode(Decision.self, from: json)
        XCTAssertEqual(d.verdict, .resist)
        XCTAssertEqual(d.verdict.label, "也许不必买")
        XCTAssertEqual(d.status, .pending)
        XCTAssertEqual(d.impulse, 72)
        XCTAssertEqual(d.dims.alt, 1, accuracy: 0.001)
        XCTAssertNil(d.decidedAt)
        // ordered() yields six labelled axes.
        XCTAssertEqual(d.dims.ordered().count, 6)
        XCTAssertEqual(d.dims.ordered().first?.label, "需求")
    }

    func testDecodeWishlistComputedFields() throws {
        let json = """
        [{
          "id":"w1","user_id":"U-1","decision_id":"D-1","item":"机械键盘","price":899,
          "cat":"digital","impulse":64,"added_at":"2026-06-06T00:00:00Z",
          "expires_at":"2026-06-07T00:00:00Z","progress":0.71,"expired":false,"remain_sec":25200
        }]
        """.data(using: .utf8)!
        let items = try decoder().decode([WishlistItem].self, from: json)
        let w = try XCTUnwrap(items.first)
        XCTAssertEqual(w.impulse, 64)
        XCTAssertFalse(w.expired)
        XCTAssertEqual(w.remainText, "剩 7 小时")
    }

    func testDecodeStats() throws {
        let json = """
        {
          "month_saved":2299,"month_spent":0,"total_saved":3698,"total_spent":88,
          "resist_count":2,"buy_count":1,"streak":20,"tree_stage":2,"restraint_rate":66,
          "by_month":[{"label":"5月","spent":88,"saved":1399},{"label":"6月","spent":0,"saved":2299}],
          "top_cats":[{"cat":"digital","label":"数码电子","color":"#4B4DAE","saved":2299}]
        }
        """.data(using: .utf8)!
        let s = try decoder().decode(Stats.self, from: json)
        XCTAssertEqual(s.totalSaved, 3698)
        XCTAssertEqual(s.treeStage, 2)
        XCTAssertEqual(s.byMonth.count, 2)
        XCTAssertEqual(s.topCats.first?.label, "数码电子")
    }

    func testDecodeSubscription() throws {
        let json = """
        {"plan":"month","is_plus":true,"plus_until":"2026-07-06T00:00:00Z",
         "plans":[{"id":"year","name":"年度","price":98,"per":"/年","note":"省 ¥118","best":true}],
         "perks":[{"icon":"spark","title":"无限咨询","desc":"想问就问"}]}
        """.data(using: .utf8)!
        let sub = try decoder().decode(Subscription.self, from: json)
        XCTAssertTrue(sub.isPlus)
        XCTAssertEqual(sub.plan, .month)
        XCTAssertEqual(sub.plans.first?.best, true)
        XCTAssertEqual(sub.perks.first?.title, "无限咨询")
    }
}
