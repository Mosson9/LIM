import SwiftUI

// MARK: - Impulse dial (0–100 arc gauge)

/// The "冲动指数" gauge: a 260°-ish sweep that fills from calm green through
/// clay to a high-impulse red, with the score and a label in the middle.
struct ImpulseDial: View {
    let value: Double          // 0...100
    var size: CGFloat = 210

    private var color: Color {
        value < 45 ? Theme.saved : value < 62 ? Theme.warn : Theme.danger
    }
    private var tip: String {
        value < 45 ? "理性" : value < 62 ? "有点心动" : "冲动"
    }
    // Sweep from 160° to 380° (i.e. -200°..20°) clockwise.
    private let startAngle = Angle(degrees: 160)
    private let sweep = 220.0

    var body: some View {
        ZStack {
            // track
            DialArc(start: startAngle.degrees, sweep: sweep, fraction: 1)
                .stroke(Theme.paper2, style: .init(lineWidth: 10, lineCap: .round))
            // value
            DialArc(start: startAngle.degrees, sweep: sweep, fraction: value / 100)
                .stroke(color, style: .init(lineWidth: 10, lineCap: .round))
                .animation(.easeOut(duration: 0.9), value: value)
            VStack(spacing: 2) {
                Text("\(Int(value.rounded()))")
                    .font(Theme.display(size * 0.3))
                    .foregroundColor(Theme.ink)
                Text(tip)
                    .font(Theme.sans(13, .semibold))
                    .tracking(2)
                    .foregroundColor(color)
            }
        }
        .frame(width: size, height: size * 0.82)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("冲动指数 \(Int(value.rounded()))，\(tip)")
    }
}

/// An arc described by a start angle, total sweep, and a 0..1 fill fraction.
private struct DialArc: Shape {
    var start: Double
    var sweep: Double
    var fraction: Double

    func path(in rect: CGRect) -> Path {
        var p = Path()
        let r = min(rect.width, rect.height) / 2 - 14
        let c = CGPoint(x: rect.midX, y: rect.midY)
        p.addArc(center: c, radius: r,
                 startAngle: .degrees(start),
                 endAngle: .degrees(start + sweep * fraction),
                 clockwise: false)
        return p
    }
}

// MARK: - Radar (six dimensions)

struct RadarChart: View {
    let data: [(label: String, value: Double)]
    var size: CGFloat = 250
    var maxValue: Double = 5

    var body: some View {
        Canvas { ctx, sz in
            let c = CGPoint(x: sz.width / 2, y: sz.height / 2)
            let R = min(sz.width, sz.height) / 2 - 34
            let n = data.count
            func point(_ i: Int, _ r: CGFloat) -> CGPoint {
                let a = -Double.pi / 2 + Double(i) * 2 * .pi / Double(n)
                return CGPoint(x: c.x + cos(a) * r, y: c.y + sin(a) * r)
            }
            // rings
            for g in 1...5 {
                var ring = Path()
                for i in 0..<n {
                    let pt = point(i, R * CGFloat(g) / 5)
                    if i == 0 { ring.move(to: pt) } else { ring.addLine(to: pt) }
                }
                ring.closeSubpath()
                ctx.stroke(ring, with: .color(Theme.hairline), lineWidth: 1)
            }
            // spokes
            for i in 0..<n {
                var spoke = Path()
                spoke.move(to: c); spoke.addLine(to: point(i, R))
                ctx.stroke(spoke, with: .color(Theme.hairline), lineWidth: 1)
            }
            // data polygon
            var poly = Path()
            for i in 0..<n {
                let pt = point(i, R * CGFloat(data[i].value / maxValue))
                if i == 0 { poly.move(to: pt) } else { poly.addLine(to: pt) }
            }
            poly.closeSubpath()
            ctx.fill(poly, with: .color(Theme.indigo.opacity(0.14)))
            ctx.stroke(poly, with: .color(Theme.indigo), lineWidth: 2)
            for i in 0..<n {
                let pt = point(i, R * CGFloat(data[i].value / maxValue))
                ctx.fill(Path(ellipseIn: CGRect(x: pt.x - 3, y: pt.y - 3, width: 6, height: 6)),
                         with: .color(Theme.indigo))
            }
            // labels
            for i in 0..<n {
                let pt = point(i, R + 18)
                ctx.draw(Text(data[i].label).font(Theme.sans(11.5, .semibold)).foregroundColor(Theme.ink2),
                         at: pt)
            }
        }
        .frame(width: size, height: size)
        .accessibilityElement()
        .accessibilityLabel("六维分析雷达图")
    }
}

// MARK: - Growth tree (savings → growth stages 0…6)

struct GrowthTree: View {
    var stage: Int = 3
    var size: CGFloat = 160

    var body: some View {
        Canvas { ctx, sz in
            let s = max(0, min(6, stage))
            let w = sz.width
            let trunkH = 18 + CGFloat(s) * 9
            let baseY = sz.height - 26
            let topY = baseY - trunkH
            let canopy = s >= 2
            let cR = CGFloat(18 + s * 7)

            // ground
            let ground = Path(ellipseIn: CGRect(x: w/2 - (42 + CGFloat(s)*5), y: baseY,
                                                width: (42 + CGFloat(s)*5) * 2, height: 16))
            ctx.fill(ground, with: .color(Theme.savedSoft))

            // trunk
            var trunk = Path()
            trunk.move(to: CGPoint(x: w/2, y: baseY))
            trunk.addLine(to: CGPoint(x: w/2, y: topY + (canopy ? cR * 0.5 : 0)))
            ctx.stroke(trunk, with: .color(Color(hex: 0x8B7355)),
                       style: .init(lineWidth: 3 + CGFloat(s) * 0.5, lineCap: .round))

            // canopy
            if canopy {
                func leaf(_ x: CGFloat, _ y: CGFloat, _ r: CGFloat, _ col: UInt32) {
                    ctx.fill(Path(ellipseIn: CGRect(x: x - r, y: y - r, width: r*2, height: r*2)),
                             with: .color(Color(hex: col, alpha: 0.92)))
                }
                leaf(w/2, topY, cR, 0x5E7E63)
                leaf(w/2 - cR*0.7, topY + cR*0.35, cR*0.66, 0x6E8B6A)
                leaf(w/2 + cR*0.7, topY + cR*0.3, cR*0.62, 0x557A57)
                leaf(w/2 + cR*0.2, topY - cR*0.5, cR*0.55, 0x6E8B6A)
            } else {
                // early sprout: two small leaves
                leafShape(&ctx, at: CGPoint(x: w/2, y: baseY - trunkH), flip: false)
                leafShape(&ctx, at: CGPoint(x: w/2, y: baseY - trunkH + 6), flip: true)
            }
        }
        .frame(width: size, height: size)
        .accessibilityHidden(true) // decorative
    }

    private func leafShape(_ ctx: inout GraphicsContext, at p: CGPoint, flip: Bool) {
        let dir: CGFloat = flip ? 1 : -1
        var path = Path()
        path.move(to: p)
        path.addQuadCurve(to: CGPoint(x: p.x, y: p.y + 12),
                          control: CGPoint(x: p.x + dir * 20, y: p.y + 2))
        ctx.fill(path, with: .color(Color(hex: flip ? 0x557A57 : 0x6E8B6A)))
    }
}

// MARK: - Mini paired bars (spent vs saved per month)

struct MiniBars: View {
    let data: [MonthBucket]
    var height: CGFloat = 110

    private var maxValue: Int { max(data.map { max($0.spent, $0.saved) }.max() ?? 1, 1) }

    var body: some View {
        HStack(alignment: .bottom, spacing: 10) {
            ForEach(data) { d in
                VStack(spacing: 7) {
                    HStack(alignment: .bottom, spacing: 3) {
                        bar(d.spent, color: Theme.spent.opacity(0.5))
                        bar(d.saved, color: Theme.saved)
                    }
                    .frame(maxHeight: .infinity, alignment: .bottom)
                    Text(d.label).font(Theme.sans(10.5)).foregroundColor(Theme.ink3)
                }
                .frame(maxWidth: .infinity)
            }
        }
        .frame(height: height)
        .accessibilityHidden(true) // values are also shown as text
    }

    private func bar(_ value: Int, color: Color) -> some View {
        let frac = CGFloat(value) / CGFloat(maxValue)
        return RoundedRectangle(cornerRadius: 4)
            .fill(color)
            .frame(width: 9, height: max(3, frac * (height - 22)))
    }
}
