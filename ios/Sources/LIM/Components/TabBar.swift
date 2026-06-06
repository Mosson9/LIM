import SwiftUI

/// The bottom tab bar with a center "ask" FAB, matching the prototype.
struct TabBar: View {
    let active: Tab
    let onSelect: (Tab) -> Void
    @EnvironmentObject var model: AppModel

    var body: some View {
        HStack(spacing: 0) {
            item(.home, "home", "首页")
            item(.growth, "sprout", "成长")
            fab
            item(.history, "clock", "历史")
            item(.me, "user", "我的")
        }
        .padding(.horizontal, 8)
        .padding(.top, 8)
        .padding(.bottom, 6)
        .background(
            Theme.surface
                .clipShape(RoundedRectangle(cornerRadius: 28, style: .continuous))
                .shadow(color: Theme.ink.opacity(0.08), radius: 18, x: 0, y: -2)
        )
        .padding(.horizontal, 14)
        .padding(.bottom, 6)
    }

    private func item(_ tab: Tab, _ icon: String, _ label: String) -> some View {
        let on = active == tab
        return Button { onSelect(tab) } label: {
            VStack(spacing: 4) {
                Icon(name: icon, size: 23, color: on ? Theme.indigo : Theme.ink3)
                Text(label).font(Theme.sans(11, on ? .semibold : .regular))
                    .foregroundColor(on ? Theme.indigo : Theme.ink3)
            }
            .frame(maxWidth: .infinity)
        }
    }

    private var fab: some View {
        Button { model.go(.ask) } label: {
            Icon(name: "spark", size: 26, color: .white, weight: .semibold)
                .frame(width: 54, height: 54)
                .background(Theme.indigo)
                .clipShape(Circle())
                .shadow(color: Theme.indigo.opacity(0.35), radius: 12, x: 0, y: 6)
        }
        .frame(maxWidth: .infinity)
        .offset(y: -10)
    }
}
