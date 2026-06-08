import SwiftUI
import UIKit

/// 更换 App 图标 — free skins plus Plus-locked ones. Applying a free skin calls
/// PUT /me/app-icon; locked skins route to the paywall.
struct AppIconView: View {
    @EnvironmentObject var model: AppModel
    @State private var sel = "classic"
    @State private var working = false

    private var skins: [Skin] { model.skins }
    private var selected: Skin? { skins.first { $0.id == sel } }

    var body: some View {
        VStack(spacing: 0) {
            TopBar(title: "更换 App 图标")
            ScrollView {
                VStack(spacing: 24) {
                    // preview
                    if let ic = selected {
                        VStack(spacing: 12) {
                            RoundedRectangle(cornerRadius: 24, style: .continuous)
                                .fill(Color(webHex: ic.bg)).frame(width: 96, height: 96)
                                .overlay(Text("L").font(Theme.display(44))
                                    .foregroundColor(ic.dark ? Theme.ink : .white))
                                .shadow(color: Theme.ink.opacity(0.12), radius: 18, y: 8)
                            Text(ic.name).font(Theme.sans(14, .semibold)).foregroundColor(Theme.ink)
                        }
                        .padding(.top, 8)
                    }

                    section("免费", skins: skins.filter { $0.free })
                    section("Plus 专属", skins: skins.filter { !$0.free }, locked: true)
                    Color.clear.frame(height: 90)
                }
                .padding(.horizontal, 22)
            }

            Button {
                Task { await apply() }
            } label: {
                if working { ProgressView().tint(.white) } else { Text("应用图标") }
            }
            .buttonStyle(FilledButtonStyle())
            .padding(.horizontal, 22).padding(.bottom, 24)
            .disabled(working)
        }
        .background(Theme.paper.ignoresSafeArea())
        .onAppear { sel = model.user?.appIcon ?? "classic" }
    }

    private func section(_ title: String, skins: [Skin], locked: Bool = false) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Kicker(text: title)
                if locked { Spacer(); Badge(text: "Plus", fg: Color(hex: 0x9A7B2E), bg: Color(hex: 0xF3ECD9), systemImage: "crown.fill") }
            }
            .padding(.leading, 4)
            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 16), count: 4), spacing: 16) {
                ForEach(skins) { ic in swatch(ic, locked: locked) }
            }
        }
    }

    private func swatch(_ ic: Skin, locked: Bool) -> some View {
        Button {
            if locked && !(model.user?.isPlus ?? false) { model.go(.plus) }
            else { sel = ic.id }
        } label: {
            ZStack {
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .fill(Color(webHex: ic.bg)).frame(width: 60, height: 60)
                    .overlay(Text("L").font(Theme.display(28)).foregroundColor(ic.dark ? Theme.ink : .white))
                    .overlay(RoundedRectangle(cornerRadius: 16)
                        .stroke(sel == ic.id ? Theme.indigo : .clear, lineWidth: 3))
                    .opacity(locked ? 0.6 : 1)
                if locked { Icon(name: "lock", size: 18, color: .white) }
            }
        }
        .buttonStyle(.plain)
    }

    /// Maps a skin id to its bundled alternate-icon asset name (classic = the
    /// primary icon, i.e. nil). The sets are in Assets.xcassets (AppIcon-*).
    private func alternateName(for id: String) -> String? {
        id == "classic" ? nil : "AppIcon-\(id.prefix(1).uppercased() + id.dropFirst())"
    }

    private func apply() async {
        working = true
        defer { working = false }
        guard let ic = selected else { return }
        if !ic.free && !(model.user?.isPlus ?? false) { model.go(.plus); return }
        // Swap the actual home-screen icon (no-op on simulators that don't support it).
        if UIApplication.shared.supportsAlternateIcons {
            try? await UIApplication.shared.setAlternateIconName(alternateName(for: ic.id))
        }
        if let updated = try? await APIClient.shared.setAppIcon(ic.id) {
            model.user = updated
        }
        model.back()
    }
}
