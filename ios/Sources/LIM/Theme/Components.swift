import SwiftUI

// MARK: - Card

/// A rounded white surface with the prototype's soft shadow.
struct Card<Content: View>: View {
    var padding: CGFloat = 18
    var background: Color = Theme.surface
    @ViewBuilder var content: Content

    var body: some View {
        content
            .padding(padding)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(background)
            .clipShape(RoundedRectangle(cornerRadius: Theme.radius, style: .continuous))
            .shadow(color: Theme.ink.opacity(0.05), radius: 14, x: 0, y: 6)
            .shadow(color: Theme.ink.opacity(0.04), radius: 3, x: 0, y: 1)
    }
}

// MARK: - Kicker (small uppercase label)

struct Kicker: View {
    let text: String
    var color: Color = Theme.ink3
    var body: some View {
        Text(text)
            .font(Theme.sans(11, .bold))
            .tracking(1.4)
            .textCase(.uppercase)
            .foregroundColor(color)
    }
}

// MARK: - Badge / pill

struct Badge: View {
    let text: String
    var fg: Color = Theme.savedDeep
    var bg: Color = Theme.savedSoft
    var systemImage: String? = nil

    var body: some View {
        HStack(spacing: 5) {
            if let s = systemImage { Image(systemName: s).font(.system(size: 11, weight: .semibold)) }
            Text(text).font(Theme.sans(12, .semibold))
        }
        .padding(.vertical, 5)
        .padding(.horizontal, 10)
        .foregroundColor(fg)
        .background(bg)
        .clipShape(Capsule())
    }
}

// MARK: - Button styles

/// Primary filled button (indigo / sage / ink variants).
struct FilledButtonStyle: ButtonStyle {
    var bg: Color = Theme.indigo
    var fg: Color = .white
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(Theme.sans(16.5, .semibold))
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .foregroundColor(fg)
            .background(bg)
            .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
            .opacity(configuration.isPressed ? 0.85 : 1)
            .scaleEffect(configuration.isPressed ? 0.99 : 1)
    }
}

/// Quiet ghost button.
struct GhostButtonStyle: ButtonStyle {
    var fg: Color = Theme.ink2
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(Theme.sans(16, .semibold))
            .frame(maxWidth: .infinity)
            .padding(.vertical, 15)
            .foregroundColor(fg)
            .background(Theme.surface)
            .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 18).stroke(Theme.hairline, lineWidth: 1))
            .opacity(configuration.isPressed ? 0.7 : 1)
    }
}

// MARK: - Chip (selectable tag)

struct Chip: View {
    let label: String
    var icon: String? = nil
    var selected: Bool = false
    var tint: Color = Theme.indigo

    var body: some View {
        HStack(spacing: 6) {
            if let icon { Image(systemName: icon).font(.system(size: 13, weight: .medium)) }
            Text(label).font(Theme.sans(14, .medium))
        }
        .padding(.vertical, 9)
        .padding(.horizontal, 14)
        .foregroundColor(selected ? .white : Theme.ink2)
        .background(selected ? tint : Theme.surface)
        .clipShape(Capsule())
        .overlay(Capsule().stroke(selected ? .clear : Theme.hairline, lineWidth: 1))
    }
}

// MARK: - Section header row ("title  全部 ›")

struct SectionHeader: View {
    let title: String
    var actionTitle: String? = nil
    var action: (() -> Void)? = nil

    var body: some View {
        HStack {
            Text(title).font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
            Spacer()
            if let actionTitle, let action {
                Button(action: action) {
                    Text(actionTitle + " ›").font(Theme.sans(13)).foregroundColor(Theme.ink3)
                }
            }
        }
    }
}

// MARK: - Screen scaffold

/// A full-bleed paper-colored screen background.
struct ScreenBackground: View {
    var color: Color = Theme.paper
    var body: some View { color.ignoresSafeArea() }
}

// MARK: - Hairline divider

struct Hairline: View {
    var body: some View { Rectangle().fill(Theme.hairline).frame(height: 1) }
}
