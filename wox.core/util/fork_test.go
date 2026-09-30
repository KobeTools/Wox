package util

import "testing"

func TestMatchesPrivateDomain(t *testing.T) {
	entries := parsePrivateDomains("lyft  # work\n\ncorp.example.com\n.Internal.IO.\n", "\n")
	cases := map[string]bool{
		"lyft.com":              true,
		"eng.lyft.net":          true,
		"lyft.atlassian.net":    true,
		"mylyfthelper.com":      false,
		"corp.example.com":      true,
		"wiki.corp.example.com": true,
		"example.com":           false,
		"notcorp.example.com":   false,
		"a.internal.io":         true,
		"github.com":            false,
		"":                      false,
	}
	for host, want := range cases {
		if got := matchesPrivateDomain(host, entries); got != want {
			t.Errorf("%q: got %v, want %v", host, got, want)
		}
	}
}
