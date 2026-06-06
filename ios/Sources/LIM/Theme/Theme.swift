import SwiftUI

/// LIM design system — a faithful port of the web prototype's CSS `:root` tokens.
///
/// 极简禅意 · 暖纸白 + 克制黑白灰 + 靛蓝强调 (calm, paper-white, restrained
/// neutrals with an indigo accent). Colors are defined as hex so the app matches
/// the prototype exactly across light backgrounds.
enum Theme {
    // MARK: Paper & surface
    static let paper      = Color(hex: 0xF4F3EE)
    static let paper2     = Color(hex: 0xEFEEE7)
    static let surface    = Color(hex: 0xFFFFFF)
    static let surface2   = Color(hex: 0xFBFAF6)

    // MARK: Ink (text)
    static let ink        = Color(hex: 0x1B1B19)
    static let ink2       = Color(hex: 0x6A6963)
    static let ink3       = Color(hex: 0x9C9A92)
    static let ink4       = Color(hex: 0xC2C0B6)
    static let hairline   = Color(hex: 0xE7E5DD)

    // MARK: Brand — indigo (AI / impulse / brand)
    static let indigo     = Color(hex: 0x34357C)
    static let indigoBright = Color(hex: 0x4B4DAE)
    static let indigoSoft = Color(hex: 0xECECF6)
    static let indigoTint = Color(hex: 0xF3F3FA)
    static let indigoInk  = Color(hex: 0x2A2B5E)

    // MARK: Semantic
    static let saved      = Color(hex: 0x5E7E63)   // savings / growth — sage
    static let savedDeep  = Color(hex: 0x46613F)
    static let savedSoft  = Color(hex: 0xE9EEE7)
    static let spent      = Color(hex: 0x8C8A82)   // spending — neutral grey
    static let spentSoft  = Color(hex: 0xEEEDE7)
    static let warn       = Color(hex: 0xC0824F)   // high impulse — warm clay
    static let warnSoft   = Color(hex: 0xF4E9DD)
    static let danger     = Color(hex: 0xB5524E)
    static let dangerSoft = Color(hex: 0xF6E5E4)
    static let gold       = Color(hex: 0xE9C36B)   // Plus accent

    // MARK: Radii
    static let radius:  CGFloat = 22
    static let radiusL: CGFloat = 30
    static let radiusS: CGFloat = 14

    // MARK: Fonts
    // The prototype pairs Noto Sans SC (body) with a Newsreader / Noto Serif SC
    // serif for "display" numerals and pull-quotes. We fall back to the system
    // serif when those families aren't bundled.
    static func display(_ size: CGFloat) -> Font { .system(size: size, weight: .regular, design: .serif) }
    static func serif(_ size: CGFloat)   -> Font { .system(size: size, weight: .regular, design: .serif) }
    static func sans(_ size: CGFloat, _ weight: Font.Weight = .regular) -> Font {
        .system(size: size, weight: weight)
    }
}

extension Color {
    /// Build a Color from a 0xRRGGBB integer.
    init(hex: UInt32, alpha: Double = 1) {
        let r = Double((hex >> 16) & 0xFF) / 255
        let g = Double((hex >> 8) & 0xFF) / 255
        let b = Double(hex & 0xFF) / 255
        self.init(.sRGB, red: r, green: g, blue: b, opacity: alpha)
    }

    /// Build a Color from a "#RRGGBB" string (used for server-provided category
    /// and skin colors). Falls back to ink on a parse failure.
    init(webHex: String) {
        var s = webHex.trimmingCharacters(in: .whitespaces)
        if s.hasPrefix("#") { s.removeFirst() }
        if let v = UInt32(s, radix: 16), s.count == 6 {
            self.init(hex: v)
        } else {
            self = Theme.ink
        }
    }
}
