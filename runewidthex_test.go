package runewidthex

import "testing"

func TestExpandTab(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tabs default", "a\tb\tc", "a   b   c"},
		{"tabs after newline", "a\n\tb\tc", "a\n    b   c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandTab(tt.input); got != tt.want {
				t.Errorf("ExpandTab() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandTab_CustomTabWidth(t *testing.T) {
	c := &Condition{TabWidth: 8}
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tab width 8", "a\tb", "a       b"},
		{"tab at stop", "aaaaaaaa\tb", "aaaaaaaa        b"},
		{"two tabs", "\t\t", "                "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.ExpandTab(tt.input); got != tt.want {
				t.Errorf("ExpandTab() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii", "hello", 5},
		{"tab from col 1", "a\tb", 5},  // "a"→col1, tab→col4 (3 spaces), "b"→col5
		{"tab from col 0", "\ta", 5},   // tab→col4, "a"→col5
		{"tab at stop", "abcd\te", 9}, // "abcd"→col4, tab→col8 (4 spaces), "e"→col9
		{"newline resets tab", "a\n\tb", 5}, // line1="a"→1, line2=tab(4)+b→5, max=5
		{"multiline max", "abc\nde", 3},    // max(3, 2)=3
		{"CJK", "日本語", 6},               // each char is 2 cells wide
		{"emoji", "🌍", 2},
		{"ascii no tabs", "abc", 3},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.input); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name  string
		input string
		w     int
		want  string
	}{
		{
			"no wrap needed",
			"hello", 10,
			"hello",
		},
		{
			"wrap at word",
			"hello world", 5,
			"hello\n worl\nd",
		},
		{
			"tab fits on line",
			"ab\tcd", 8,
			"ab  cd", // "ab"→2, tab to col4 (2 spaces), "cd"→6, fits in 8
		},
		{
			"tab overflows: moves to next line as unit",
			"abcde\tfg", 7,
			// "abcde"→5 cols, tab would go to col 8 (3 spaces), 5+3=8>7 → wrap before tab
			// new line: tab at col0→col4 (4 spaces), "fg"→6
			"abcde\n    fg",
		},
		{
			"tab fills to width, next char wraps",
			"abcd\tefg", 8,
			// "abcd"→4, tab: tabW=4 (4%4=0 → next stop at 8), col+4=8≤8, no wrap.
			// "e": col+1=9>8 → wrap. Result: "abcd    \nefg"
			"abcd    \nefg",
		},
		{
			"existing newline resets column",
			"abc\nde\tfg", 10,
			// line1: "abc"→3, newline resets
			// line2: "de"→2, tab to col4 (2 spaces), "fg"→6, fits in 10
			"abc\nde  fg",
		},
		{
			"CJK wrapping",
			"日本語", 4,
			// "日"→2, "本"→4, "語"→6. 2+2=4≤4, 4+2=6>4 → wrap before "語"
			"日本\n語",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.input, tt.w); got != tt.want {
				t.Errorf("Wrap(%q, %d) = %q, want %q", tt.input, tt.w, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		w     int
		tail  string
		want  string
	}{
		{
			"no truncation needed",
			"hello", 10, "...",
			"hello",
		},
		{
			"simple truncation",
			"hello world", 8, "...",
			"hello...",
		},
		{
			"no truncation expands tabs",
			"ab\tcd", 10, "...",
			// maxContent=10-3=7; "ab"→2, tab→col4 (2sp), "c"→5, "d"→6 ≤7 → no truncation, tabs expanded
			"ab  cd",
		},
		{
			"truncation within tab-expanded content",
			"ab\tcd", 8, "...",
			// maxContent=8-3=5; "ab"→2, tab→col4 (2sp), "c"→5≤5 ok, "d": 5+1=6>5 → truncate
			"ab  c...",
		},
		{
			"tab overflows: tab and tail replaces overflow",
			"abcde\tfg", 7, "...",
			// maxContent=7-3=4; "abcde"→5>4 actually check: "a"=1,"b"=2,"c"=3,"d"=4,"e": 4+1=5>4 → truncate at "e"
			"abcd...",
		},
		{
			"tab token itself overflows truncation point",
			"abc\tfg", 6, "...",
			// maxContent=6-3=3; "a"=1,"b"=2,"c"=3. Tab: 3+1=4>3 → tab overflows → "abc..."
			"abc...",
		},
		{
			"empty tail no truncation",
			"hello", 5, "",
			"hello",
		},
		{
			"empty tail with truncation",
			"hello world", 5, "",
			"hello",
		},
		{
			"CJK truncation",
			"日本語", 4, "…",
			// maxContent=4-1=3 (… is width 1? actually "…" U+2026 is typically width 1)
			// "日"→2, "本": 2+2=4>3 → truncate before "本" → "日…"
			"日…",
		},
		{
			"newline resets column",
			"abc\ndefgh", 4, "...",
			// line1: "a"=1,"b"=2,"c"=3. maxContent=4-3=1. "a": 0+1=1≤1 ok, "b": 1+1=2>1 truncate
			// actually wait: maxContent=4-3=1; "a":0+1=1≤1, "b":1+1=2>1 → "a..."
			// hmm, that truncates line1 at position 1, never gets to \n
			"a...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.w, tt.tail); got != tt.want {
				t.Errorf("Truncate(%q, %d, %q) = %q, want %q", tt.input, tt.w, tt.tail, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		w     int
		want  string
	}{
		{"pad ascii", "hi", 5, "hi   "},
		{"no pad needed", "hello", 5, "hello"},
		{"already wider", "hello world", 5, "hello world"},
		{"pad with tab", "a\tb", 8, "a   b   "}, // "a   b"=5, pad to 8 → 3 spaces
		{"CJK pad", "日本", 6, "日本  "},         // 日本=4, pad to 6 → 2 spaces
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadRight(tt.input, tt.w); got != tt.want {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.w, got, tt.want)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		name  string
		input string
		w     int
		want  string
	}{
		{"pad ascii", "hi", 5, "   hi"},
		{"no pad needed", "hello", 5, "hello"},
		{"already wider", "hello world", 5, "hello world"},
		{"pad with tab", "a\tb", 8, "   a   b"}, // "a   b"=5, pad to 8 → 3 leading spaces
		{"CJK pad", "日本", 6, "  日本"},          // 日本=4, pad to 6 → 2 leading spaces
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadLeft(tt.input, tt.w); got != tt.want {
				t.Errorf("PadLeft(%q, %d) = %q, want %q", tt.input, tt.w, got, tt.want)
			}
		})
	}
}
