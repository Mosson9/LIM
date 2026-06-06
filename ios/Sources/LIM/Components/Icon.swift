import SwiftUI

/// Maps the prototype's line-icon names onto SF Symbols so screens can keep
/// using the original semantic names (`spark`, `sprout`, `bolt`, …).
struct Icon: View {
    let name: String
    var size: CGFloat = 24
    var color: Color = Theme.ink
    var weight: Font.Weight = .medium

    private static let map: [String: String] = [
        "home": "house", "spark": "sparkles", "leaf": "leaf.fill", "user": "person",
        "chart": "chart.bar", "clock": "clock", "list": "list.bullet", "plus": "plus",
        "close": "xmark", "back": "chevron.left", "chevR": "chevron.right", "chevD": "chevron.down",
        "check": "checkmark", "heart": "heart.fill", "camera": "camera", "settings": "gearshape",
        "crown": "crown.fill", "bolt": "bolt.fill", "bell": "bell", "grid": "square.grid.2x2",
        "wallet": "creditcard", "tag": "tag", "scale": "scalemass", "refresh": "arrow.clockwise",
        "moon": "moon", "info": "info.circle", "search": "magnifyingglass",
        "arrowDown": "arrow.down", "arrowUp": "arrow.up", "sprout": "leaf", "target": "target",
        "lock": "lock", "send": "paperplane",
    ]

    var body: some View {
        Image(systemName: Self.map[name] ?? "circle")
            .font(.system(size: size, weight: weight))
            .foregroundColor(color)
    }
}
