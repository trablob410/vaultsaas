package cmd

import "testing"

func TestParseDuration(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"2h", 120},
		{"30m", 30},
		{"1h30m", 90},
		{"8h", 480},
		{"", 60}, // default 1h
		{"0m", 60}, // zero → default
	}
	for _, tc := range cases {
		got := parseDuration(tc.input)
		if got != tc.want {
			t.Errorf("parseDuration(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
