package util

// KobeTools fork switches, in one place so each upstream sync can re-check them.
// This build comes from reviewed source; nothing talks to the network unless
// the user asks for it.
const (
	// ForkDisableUpdates: never check for, download, or apply upstream binaries
	// (upstream verifies them with MD5 from an unsigned manifest).
	ForkDisableUpdates = true
	// ForkDisableTelemetry: no usage pings, whatever the setting says.
	ForkDisableTelemetry = true
	// ForkDisableStores: no plugin/theme/AI-command store polling. Store plugins
	// are unverified third-party code; install reviewed plugins by hand.
	ForkDisableStores = true
	// ForkDisableRemoteFavicons: website icons come from the local cache only,
	// so copied links and bookmarks aren't sent to Google or fetched from sites.
	ForkDisableRemoteFavicons = true
	// ForkDisableDNSFallback: no retry through public DNS (1.1.1.1 / 8.8.8.8)
	// when the system resolver fails, so local DNS blocking is respected.
	ForkDisableDNSFallback = true
)
