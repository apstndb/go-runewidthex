package runewidthex

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// Condition have flag EastAsianWidth whether the current locale is CJK or not.
type Condition struct {
	BaseCondition *runewidth.Condition
	TabWidth      int
}

func NewCondition() *Condition {
	return &Condition{
		BaseCondition: runewidth.NewCondition(),
		TabWidth:      4,
	}
}

func (c *Condition) StringWidth(s string) int {
	var width int
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		var chWidth int
		runes := g.Runes()
		for _, r := range runes {
			if r == '\t' {
				chWidth = c.TabWidth - width%c.TabWidth
				break
			}

			chWidth = c.BaseCondition.RuneWidth(r)

			if chWidth > 0 {
				break
			}
		}
		width += chWidth
	}
	return width
}

func (c *Condition) ExpandTab(s string) string {
	var sb strings.Builder
	var width int
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		var chWidth int
		runes := g.Runes()
		for i, r := range runes {
			if r == '\n' {
				width = 0
				chWidth = 0
				sb.WriteRune(r)
				break
			} else if r == '\t' {
				chWidth = c.TabWidth - width%c.TabWidth
				sb.WriteString(strings.Repeat(" ", chWidth))
				break
			}

			chWidth = c.BaseCondition.RuneWidth(r)
			sb.WriteRune(runes[i])

			if chWidth > 0 {
				// flush remaining runes in grapheme
				for i := i + 1; i < len(runes); i++ {
					sb.WriteRune(runes[i])
				}

				break
			}
		}
		width += chWidth
	}
	return sb.String()
}

// Wrap return string wrapped with w cells
func (c *Condition) Wrap(s string, w int) string {
	width := 0
	out := ""
	for _, r := range s {
		cw := c.BaseCondition.RuneWidth(r)
		if r == '\n' {
			out += string(r)
			width = 0
			continue
		}
		if r == '\t' {
			cw = c.TabWidth - width%c.TabWidth
			if width+cw > w {
				out += "\n"
				width = 0
				cw = c.TabWidth
			}

			out += strings.Repeat(" ", cw)
			width += cw
			continue
		}

		if width+cw > w {
			out += "\n"
			width = 0
			out += string(r)
			width += cw
			continue
		}

		out += string(r)
		width += cw
	}
	return out
}

func ExpandTab(s string) string {
	return NewCondition().ExpandTab(s)
}

func Wrap(s string, w int) string {
	return NewCondition().Wrap(s, w)
}
