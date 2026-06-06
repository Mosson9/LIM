import SwiftUI

/// A calm, centered empty state (icon + title + optional subtitle).
struct EmptyStateView: View {
    var icon: String = "leaf"
    var title: String
    var subtitle: String? = nil

    var body: some View {
        VStack(spacing: 6) {
            Icon(name: icon, size: 40, color: Theme.ink4)
                .padding(.bottom, 8)
            Text(title).font(Theme.sans(15)).foregroundColor(Theme.ink3)
            if let subtitle {
                Text(subtitle).font(Theme.sans(13)).foregroundColor(Theme.ink4)
                    .multilineTextAlignment(.center)
            }
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 56)
        .accessibilityElement(children: .combine)
    }
}

/// A dismissible top banner used to surface transient network errors.
struct Banner: View {
    let text: String
    var onDismiss: () -> Void

    var body: some View {
        HStack(spacing: 10) {
            Icon(name: "info", size: 16, color: .white)
            Text(text).font(Theme.sans(13, .medium)).foregroundColor(.white)
            Spacer(minLength: 0)
            Button(action: onDismiss) { Icon(name: "close", size: 15, color: .white.opacity(0.8)) }
                .accessibilityLabel("关闭提示")
        }
        .padding(.horizontal, 16).padding(.vertical, 12)
        .background(Theme.ink)
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        .shadow(color: Theme.ink.opacity(0.2), radius: 12, y: 4)
        .padding(.horizontal, 16)
        .transition(.move(edge: .top).combined(with: .opacity))
    }
}
