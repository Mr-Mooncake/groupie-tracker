package helpers

import "testing"

func TestFormatLocation(t *testing.T) {
	tests := []struct{ in, want string }{
		{"london-uk", "London, UK"},
		{"north_carolina-usa", "North Carolina, USA"},
		{"playa_del_carmen-mexico", "Playa Del Carmen, Mexico"},
		{"new_zealand", "New Zealand"},
		{"paris", "Paris"},
		{"  london-uk  ", "London, UK"},
		{"LONDON-UK", "London, UK"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := FormatLocation(tt.in); got != tt.want {
			t.Errorf("FormatLocation(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
