package go_runewidthex

import "testing"

func TestExpandTab(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tabs", "a\tb\tc", "a   b   c"},
		{"tabs", "a\n\tb\tc", "a\n    b   c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandTab(tt.input); got != tt.want {
				t.Errorf("ExpandTab() = %v, want %v", got, tt.want)
			}
		})
	}
}
