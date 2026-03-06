package browser

import (
	"testing"
)

func TestResolveChromiumWindowFlags(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		config     BrowserConfig
		wantFlags  []string
		wantOk     bool
	}{
		{
			name:       "Chrome with newWindow",
			identifier: "Google Chrome",
			config: BrowserConfig{
				Name:      "Google Chrome",
				NewWindow: boolPtr(true),
			},
			wantFlags: []string{"--new-window"},
			wantOk:    true,
		},
		{
			name:       "Chrome with incognito",
			identifier: "Google Chrome",
			config: BrowserConfig{
				Name:      "Google Chrome",
				Incognito: boolPtr(true),
			},
			wantFlags: []string{"--incognito"},
			wantOk:    true,
		},
		{
			name:       "Chrome with both newWindow and incognito",
			identifier: "Google Chrome",
			config: BrowserConfig{
				Name:      "Google Chrome",
				NewWindow: boolPtr(true),
				Incognito: boolPtr(true),
			},
			wantFlags: []string{"--new-window", "--incognito"},
			wantOk:    true,
		},
		{
			name:       "Chrome with newWindow false",
			identifier: "Google Chrome",
			config: BrowserConfig{
				Name:      "Google Chrome",
				NewWindow: boolPtr(false),
			},
			wantFlags: nil,
			wantOk:    false,
		},
		{
			name:       "Non-Chromium browser (Safari)",
			identifier: "Safari",
			config: BrowserConfig{
				Name:      "Safari",
				NewWindow: boolPtr(true),
			},
			wantFlags: nil,
			wantOk:    false,
		},
		{
			name:       "Brave Browser with newWindow",
			identifier: "Brave Browser",
			config: BrowserConfig{
				Name:      "Brave Browser",
				NewWindow: boolPtr(true),
			},
			wantFlags: []string{"--new-window"},
			wantOk:    true,
		},
		{
			name:       "Unknown browser",
			identifier: "Unknown Browser",
			config: BrowserConfig{
				Name:      "Unknown Browser",
				NewWindow: boolPtr(true),
			},
			wantFlags: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlags, gotOk := resolveChromiumWindowFlags(tt.identifier, tt.config)

			if gotOk != tt.wantOk {
				t.Errorf("resolveChromiumWindowFlags() gotOk = %v, want %v", gotOk, tt.wantOk)
			}

			if len(gotFlags) != len(tt.wantFlags) {
				t.Errorf("resolveChromiumWindowFlags() gotFlags length = %v, want %v", len(gotFlags), len(tt.wantFlags))
				return
			}

			for i, flag := range gotFlags {
				if flag != tt.wantFlags[i] {
					t.Errorf("resolveChromiumWindowFlags() gotFlags[%d] = %v, want %v", i, flag, tt.wantFlags[i])
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
