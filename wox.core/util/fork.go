package util

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

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
	// ForkDisableRemoteFavicons: when true, website icons come from the local
	// cache only. Off: icons are fetched normally, except for the user's private
	// domains (see IsPrivateFaviconHost), which are never sent to Google or fetched.
	ForkDisableRemoteFavicons = false
	// ForkDisableDNSFallback: no retry through public DNS (1.1.1.1 / 8.8.8.8)
	// when the system resolver fails, so local DNS blocking is respected.
	ForkDisableDNSFallback = true
)

// PrivateDomainsFileName lives in the Wox settings folder (so settings backups
// carry it). One entry per line, "#" starts a comment:
//
//	lyft              any host with "lyft" as a label: lyft.com, x.lyft.net, lyft.atlassian.net
//	corp.example.com  that domain and its subdomains
//
// WOX_PRIVATE_DOMAINS (comma-separated) is read too.
const PrivateDomainsFileName = "private-domains.txt"

var privateDomains struct {
	sync.Mutex
	entries []string
	modTime time.Time
	loaded  bool
}

// IsPrivateFaviconHost reports whether host matches one of the user's private
// domains, whose icons must never be fetched from Google or the site itself.
func IsPrivateFaviconHost(host string) bool {
	return matchesPrivateDomain(strings.ToLower(strings.TrimSuffix(host, ".")), loadPrivateDomains())
}

func matchesPrivateDomain(host string, entries []string) bool {
	if host == "" {
		return false
	}
	labels := strings.Split(host, ".")
	for _, entry := range entries {
		if strings.Contains(entry, ".") {
			if host == entry || strings.HasSuffix(host, "."+entry) {
				return true
			}
			continue
		}
		for _, label := range labels {
			if label == entry {
				return true
			}
		}
	}
	return false
}

func loadPrivateDomains() []string {
	privateDomains.Lock()
	defer privateDomains.Unlock()

	path := filepath.Join(GetLocation().GetPluginSettingDirectory(), PrivateDomainsFileName)
	info, statErr := os.Stat(path)
	if privateDomains.loaded && (statErr != nil || info.ModTime().Equal(privateDomains.modTime)) {
		return privateDomains.entries
	}

	entries := parsePrivateDomains(os.Getenv("WOX_PRIVATE_DOMAINS"), ",")
	if statErr == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			entries = append(entries, parsePrivateDomains(string(data), "\n")...)
		}
		privateDomains.modTime = info.ModTime()
	}
	privateDomains.entries = entries
	privateDomains.loaded = true
	return entries
}

func parsePrivateDomains(text, sep string) []string {
	var entries []string
	for _, line := range strings.Split(text, sep) {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		if entry := strings.ToLower(strings.Trim(strings.TrimSpace(line), ".")); entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}
