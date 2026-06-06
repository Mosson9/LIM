import Foundation

/// Localization helper. Wrap user-facing strings in `L("key")` (or `L("key",
/// args…)` for format strings) and add the key to every `*.lproj/Localizable.strings`.
///
/// The app ships zh-Hans (development language) and en. The entry flow (tab bar,
/// onboarding, common buttons) is localized as the reference pattern; remaining
/// screens migrate the same way — replace a literal with `L("…")` and add the key.
func L(_ key: String) -> String {
    NSLocalizedString(key, comment: "")
}

func L(_ key: String, _ args: CVarArg...) -> String {
    String(format: NSLocalizedString(key, comment: ""), arguments: args)
}
