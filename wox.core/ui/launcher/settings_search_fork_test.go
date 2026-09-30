package launcher

import "testing"

// KobeTools fork: dropdown options are searchable and named in the result.
func TestMatchedSettingsSearchChoice(t *testing.T) {
	candidate := settingsSearchResult{title: "Primary glance", choiceLabels: []string{"Time", "Date", "Date & time", "Battery"}}
	cases := map[string]string{
		"date & time": "Date & time",
		"battery":     "Battery",
		"primary":     "", // matches the setting's own name: no option shown
		"zzz":         "",
	}
	for query, want := range cases {
		if got := matchedSettingsSearchChoice(candidate, query, false); got != want {
			t.Errorf("%q: got %q, want %q", query, got, want)
		}
	}
}
