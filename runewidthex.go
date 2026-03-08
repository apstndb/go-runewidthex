// Package runewidthex is a displaywidth-based, tab-aware string formatting
// engine. It uses [github.com/clipperhouse/displaywidth] for grapheme-cluster
// display-width measurement and adds tab expansion, wrapping, truncation, and
// padding with tab-stop-aware layout.
//
// All string-returning functions render tabs as spaces in their output.
// Layout decisions (wrap/truncate boundaries) are made while tabs are still
// treated as indivisible tokens whose width depends on the current column
// position.
package runewidthex

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
)

// Condition holds configuration for display-width computation, tab expansion,
// wrapping, and truncation of strings.
type Condition struct {
	// Options controls display-width behaviour (e.g. EastAsianWidth, ANSI
	// escape-sequence awareness). See [displaywidth.Options] for details.
	Options displaywidth.Options
	// TabWidth is the tab-stop interval used when measuring or expanding tabs.
	// Defaults to 4.
	TabWidth int
}

// NewCondition returns a Condition with default settings (tab width 4, no
// East Asian width, no control-sequence handling).
func NewCondition() *Condition {
	return &Condition{
		TabWidth: 4,
	}
}

// StringWidth returns the display width of s. Newlines reset the column
// counter so that tab stops are computed correctly per line; the returned
// value is the maximum line width across all lines.
func (c *Condition) StringWidth(s string) int {
	col, max := 0, 0
	g := c.Options.StringGraphemes(s)
	for g.Next() {
		v := g.Value()
		switch v {
		case "\n":
			if col > max {
				max = col
			}
			col = 0
		case "\t":
			col += c.TabWidth - col%c.TabWidth
		default:
			col += g.Width()
		}
	}
	if col > max {
		max = col
	}
	return max
}

// ExpandTab returns s with all tab characters replaced by spaces aligned to
// the nearest tab stop. Newlines reset the column counter.
func (c *Condition) ExpandTab(s string) string {
	var sb strings.Builder
	col := 0
	g := c.Options.StringGraphemes(s)
	for g.Next() {
		v := g.Value()
		switch v {
		case "\n":
			sb.WriteString("\n")
			col = 0
		case "\t":
			w := c.TabWidth - col%c.TabWidth
			sb.WriteString(strings.Repeat(" ", w))
			col += w
		default:
			sb.WriteString(v)
			col += g.Width()
		}
	}
	return sb.String()
}

// Wrap returns s wrapped so that no line exceeds w display cells. Tabs are
// treated as indivisible layout tokens: if a tab does not fit on the current
// line it moves as a unit to the next line. Tabs in the returned string are
// rendered as spaces.
func (c *Condition) Wrap(s string, w int) string {
	var sb strings.Builder
	col := 0
	g := c.Options.StringGraphemes(s)
	for g.Next() {
		v := g.Value()
		switch v {
		case "\n":
			sb.WriteString("\n")
			col = 0
		case "\t":
			tabW := c.TabWidth - col%c.TabWidth
			if col > 0 && col+tabW > w {
				sb.WriteString("\n")
				col = 0
				tabW = c.TabWidth // tab at column 0 always spans a full stop
			}
			sb.WriteString(strings.Repeat(" ", tabW))
			col += tabW
		default:
			gw := g.Width()
			if col > 0 && col+gw > w {
				sb.WriteString("\n")
				col = 0
			}
			sb.WriteString(v)
			col += gw
		}
	}
	return sb.String()
}

// Truncate returns s truncated to at most w display cells, appending tail if
// truncation occurs. Tabs are treated as indivisible layout tokens: if a tab
// would overflow the truncation boundary the tab (and everything after it) is
// replaced by tail. Tabs in the returned string are rendered as spaces.
// Newlines reset the column counter so that each line is truncated to w cells
// independently.
func (c *Condition) Truncate(s string, w int, tail string) string {
	maxContent := w - c.StringWidth(tail)
	var result strings.Builder
	col := 0
	g := c.Options.StringGraphemes(s)
	for g.Next() {
		v := g.Value()
		switch v {
		case "\n":
			result.WriteString("\n")
			col = 0
		case "\t":
			tabW := c.TabWidth - col%c.TabWidth
			if col+tabW > maxContent {
				result.WriteString(tail)
				return result.String()
			}
			result.WriteString(strings.Repeat(" ", tabW))
			col += tabW
		default:
			gw := g.Width()
			if col+gw > maxContent {
				result.WriteString(tail)
				return result.String()
			}
			result.WriteString(v)
			col += gw
		}
	}
	return result.String()
}

// PadRight returns s with tabs expanded and trailing spaces appended so that
// the total display width is exactly w cells. If s is already w or wider, s
// is returned with tabs expanded but otherwise unchanged.
func (c *Condition) PadRight(s string, w int) string {
	sw := c.StringWidth(s)
	expanded := c.ExpandTab(s)
	if sw >= w {
		return expanded
	}
	return expanded + strings.Repeat(" ", w-sw)
}

// PadLeft returns s with tabs expanded and leading spaces prepended so that
// the total display width is exactly w cells. If s is already w or wider, s
// is returned with tabs expanded but otherwise unchanged.
func (c *Condition) PadLeft(s string, w int) string {
	sw := c.StringWidth(s)
	expanded := c.ExpandTab(s)
	if sw >= w {
		return expanded
	}
	return strings.Repeat(" ", w-sw) + expanded
}

// Package-level convenience functions using a default Condition (TabWidth 4).

// StringWidth returns the display width of s using default settings.
func StringWidth(s string) int {
	return NewCondition().StringWidth(s)
}

// ExpandTab returns s with tabs expanded to spaces using default settings.
func ExpandTab(s string) string {
	return NewCondition().ExpandTab(s)
}

// Wrap returns s wrapped at w display cells using default settings.
func Wrap(s string, w int) string {
	return NewCondition().Wrap(s, w)
}

// Truncate returns s truncated to w display cells with tail appended on
// truncation, using default settings.
func Truncate(s string, w int, tail string) string {
	return NewCondition().Truncate(s, w, tail)
}

// PadRight pads s on the right to w display cells using default settings.
func PadRight(s string, w int) string {
	return NewCondition().PadRight(s, w)
}

// PadLeft pads s on the left to w display cells using default settings.
func PadLeft(s string, w int) string {
	return NewCondition().PadLeft(s, w)
}
