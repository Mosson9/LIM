# LIM · iOS app (SwiftUI)

The native client for **LIM (Less is More)** — an anti-consumerism companion.
When you're about to buy something you're not sure you need, you ask LIM; it
scores the purchase across six dimensions, shows an *impulse index*, and gently
helps you decide. Resisting becomes savings and a growing tree; pausing parks
the item in a 24-hour cooling-off list; buying records a clean, deliberate
purchase.

Built with SwiftUI, targeting **iOS 16+**. It talks to the Go backend in
[`../backend`](../backend) over the REST API documented in
[`../docs/API.md`](../docs/API.md).

## Requirements

- Xcode 15+
- iOS 16 simulator or device
- A running LIM backend (see [`../backend/README.md`](../backend/README.md))
- [XcodeGen](https://github.com/yonyz/XcodeGen) to generate the project:
  `brew install xcodegen`

## Run

```bash
# 1. Start the backend (in another terminal)
cd ../backend && make run            # serves http://localhost:8080

# 2. Generate and open the Xcode project
cd ios
xcodegen generate
open LIM.xcodeproj
```

Pick an iPhone simulator and hit **Run**. Sign in with a seeded demo account
(`lin@lim.app` / `password`) or tap **去注册** to create a new one and walk
through onboarding.

## Configuring the API base URL

The client reads `LIM_API_BASE_URL` from `Resources/Info.plist` (falls back to
`http://localhost:8080`). For a device on your LAN, set it to your Mac's IP, e.g.
`http://192.168.1.20:8080`. You can also override it via the scheme's
environment variable `LIM_API_BASE_URL`.

The Info.plist ships an App Transport Security exception for `localhost` so the
simulator can use plaintext HTTP during development. **Remove it and use HTTPS
for any real deployment.**

## Project layout

```
ios/
├── project.yml                 # XcodeGen spec → LIM.xcodeproj
├── Resources/
│   ├── Info.plist              # base URL + ATS localhost exception
│   └── Assets.xcassets         # AccentColor, AppIcon, launch color
└── Sources/LIM/
    ├── LIMApp.swift            # @main, root routing (splash / auth / shell)
    ├── Theme/                  # design tokens + reusable styled components
    ├── Models/                 # Codable structs mirroring the API
    ├── Networking/             # APIClient (async/await), config
    ├── Store/                  # AppModel — session, navigation, data cache
    ├── Components/             # Icon, charts (dial/radar/tree/bars), tab bar
    └── Screens/                # 14 screens (onboarding → result → growth …)
```

## Architecture notes

- **`AppModel`** (`@MainActor`, `ObservableObject`) is the single source of
  truth: it owns the session (JWT persisted in `UserDefaults`), the navigation
  stack, and cached catalogue/stats data. Every screen reads it via
  `@EnvironmentObject`.
- **Navigation** mirrors the web prototype's screen stack: root tabs
  (home/growth/history/me) reset the stack; the ask→analyzing→result→record/saved
  flow pushes. See `Screen` in `Store/AppModel.swift`.
- **`APIClient`** is an `actor` wrapping `URLSession`; it decodes snake_case JSON
  and attaches the bearer token automatically.
- **Charts** (impulse dial, six-axis radar, growth tree, paired bars) are drawn
  with SwiftUI `Shape`/`Canvas` — no third-party dependencies.
- **No external Swift packages** — the whole app is the standard library +
  SwiftUI.

## Design system

Ported 1:1 from the prototype's CSS tokens (`Theme/Theme.swift`): paper-white
canvas `#F4F3EE`, ink `#1B1B19`, indigo brand/AI accent `#34357C`, sage savings
`#5E7E63`, clay high-impulse `#C0824F`. Numerals and pull-quotes use a serif
(`Newsreader`/system serif); body copy uses the system sans (Noto Sans SC on
device).
