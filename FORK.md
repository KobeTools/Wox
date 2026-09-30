# KobeTools fork of Wox

Upstream: https://github.com/Wox-launcher/Wox (branch `master`). This fork is built
from reviewed source with no telemetry, no auto-update and no background network
calls. Re-check every item below after each upstream sync.

## Fork changes

| Change | Where |
|---|---|
| All fork switches in one file | `wox.core/util/fork.go` |
| No update checks, downloads or installs (upstream verifies binaries with MD5 from an unsigned manifest) | `updater/updater.go` (`ForkDisableUpdates` in `StartAutoUpdateChecker`, `getLatestVersion`, `downloadUpdate`, `ApplyUpdate`) |
| No telemetry ping (was on by default, to `wox-telemetry.qlf.workers.dev`) | `telemetry/sender.go` |
| No plugin / theme / AI-command store polling (was every 10 minutes). Lists are fetched when you browse them (Settings, `wpm`/`theme` queries). Installing store plugins still runs unverified third-party code, so review first | `plugin/store.go`, `ui/store.go`, `ai/store.go`, `ui/settings_theme_services.go`, `plugin/system/theme.go` |
| Website icons are fetched normally, except for the user's **private domains**, which are never sent to Google or fetched (cache only). List them per machine in `%USERPROFILE%\.wox\wox-user\settings\private-domains.txt` (one per line; `lyft` matches any host with that label, `corp.example.com` matches that domain and subdomains, `!lyft.com` is an exception for the public site and its `www.`; no file means nothing is private) or `WOX_PRIVATE_DOMAINS`. `ForkDisableRemoteFavicons` turns all remote icons off | `util/fork.go` (`IsPrivateFaviconHost`), `util/websiteicon/website_icon.go` |
| No public-DNS retry (1.1.1.1 / 8.8.8.8) that bypassed local DNS blocking | `util/http.go` |
| Currency rates fetched on first currency query, not at startup + hourly | `plugin/system/converter/converter.go` |
| Typing a URL ranks "Open in browser" first (upstream scored it 100, below bookmarks whose URLs merely contain the text) | `plugin/system/url.go` |
| Defaults: auto-update off, usage stats off, AI `bash` tool disabled | `setting/wox_setting.go` |
| Plugin hosts listen on 127.0.0.1 only and reject browser connections (upstream: all interfaces, no auth) | `wox.plugin.host.python/src/wox_plugin_host/host.py`, `wox.plugin.host.nodejs/src/index.ts` |
| Python host zipapp built from `uv.lock` with `--require-hashes` (upstream: shiv ran pip against PyPI, unpinned); `ruff format` no longer part of the build | `wox.plugin.host.python/Makefile` |
| Pinned-toolchain build script | `scripts/build-install-local.sh`, `scripts/toolchain.env` |
| macOS builds sign with the local "KobeTools Dev" identity when present (upstream's fallback pins the requirement to the bundle identifier only), no hardened runtime, no notarization | `Makefile` (`_bundle_mac_app`) |

## Known gaps (accepted, review on sync)

- Dictation ASR models and `silero_vad.onnx` download without a checksum (`util/speech/model_manager.go`). Only happens if you enable dictation; DualWhisper covers dictation instead.
- `goversioninfo` runs via `go run …@v1.5.0` (verified by sum.golang.org, not `go.sum`).
- Emoji images load from `cdn.jsdelivr.net` (twemoji) when not cached.
- MCP stdio servers start at launch even when marked Disabled (upstream bug); configure only servers you trust.
- `wox.core/resource/others/webview/WebView2Loader.dll` is prebuilt (Microsoft-signed); its hash is pinned in `scripts/toolchain.env`.
- Local data (settings incl. AI keys, clipboard, chats) is stored unencrypted in `%USERPROFILE%\.wox`.
- Builds are unsigned: Smart App Control / AV may flag the keyboard-hook DLL.

## Syncing

Sync to upstream **release tags**, not `master` (tags are unsigned and the project is
essentially single-maintainer, so review by commit SHA). Run
`scripts/audit-upstream.sh wox` from mactools, then check that every row above
still holds, especially `util/fork.go` usages and the two plugin-host bind lines.
