import SwiftUI

/// Header bar for pushed screens: a leading back/close button, a centered title,
/// and an optional trailing accessory.
struct TopBar<Trailing: View>: View {
    let title: String
    let leading: Leading
    let onLeading: (() -> Void)?
    let trailing: () -> Trailing
    @EnvironmentObject var model: AppModel

    enum Leading { case back, close, none }

    init(title: String = "", leading: Leading = .back, onLeading: (() -> Void)? = nil,
         @ViewBuilder trailing: @escaping () -> Trailing) {
        self.title = title
        self.leading = leading
        self.onLeading = onLeading
        self.trailing = trailing
    }

    var body: some View {
        HStack(spacing: 12) {
            switch leading {
            case .back:
                button(icon: "back")
            case .close:
                button(icon: "close")
            case .none:
                Color.clear.frame(width: 38, height: 38)
            }
            Spacer(minLength: 0)
            if !title.isEmpty {
                Text(title).font(Theme.sans(17, .semibold)).foregroundColor(Theme.ink)
            }
            Spacer(minLength: 0)
            trailing()
        }
        .frame(minHeight: 50)
        .padding(.horizontal, 16)
        .padding(.top, 54)
    }

    private func button(icon: String) -> some View {
        Button { (onLeading ?? model.back)() } label: {
            Icon(name: icon, size: 22, color: Theme.ink)
                .frame(width: 38, height: 38)
        }
    }
}

extension TopBar where Trailing == EmptyView {
    init(title: String = "", leading: Leading = .back, onLeading: (() -> Void)? = nil) {
        self.init(title: title, leading: leading, onLeading: onLeading, trailing: { EmptyView() })
    }
}
