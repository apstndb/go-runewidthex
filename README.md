# go-runewidthex

A displaywidth-based, tab-aware string formatting engine for Go.

## Overview

`go-runewidthex` builds on [`clipperhouse/displaywidth`](https://github.com/clipperhouse/displaywidth) for grapheme-cluster display-width measurement and adds tab expansion, wrapping, truncation, and padding with correct tab-stop-aware layout.

It is designed as a string formatting layer suitable for downstream use in terminal table renderers such as [`apstndb/spanner-mycli`](https://github.com/apstndb/spanner-mycli), which preprocesses strings before passing them to [`olekukonko/tablewriter`](https://github.com/olekukonko/tablewriter).

## Key design choices

- **Tabs are layout tokens.** When computing wrap or truncation boundaries, each `\t` is treated as an indivisible unit whose display width depends on the current column position and the configured tab-stop interval. A tab that would overflow a line moves as a whole to the next line rather than being split.
- **String-returning functions expand tabs.** All functions that return a `string` (`ExpandTab`, `Wrap`, `Truncate`, `PadLeft`, `PadRight`) render tab characters as sequences of spaces in their output. Width-computing functions (`StringWidth`) measure tabs using the same column-aware semantics.
- **Grapheme-cluster aware.** Display widths are measured at the grapheme-cluster level using `clipperhouse/displaywidth`, so multi-codepoint sequences (emoji, combining marks, CJK characters) are handled correctly.
- **Per-line width semantics.** Newlines reset the column counter so tab stops are computed correctly per line. `StringWidth` returns the maximum line width across all lines.

## Usage

```go
import "github.com/apstndb/go-runewidthex"

// Package-level functions use default settings (tab width 4).
width := runewidthex.StringWidth("hello\t世界")  // tab + CJK
expanded := runewidthex.ExpandTab("a\tb")        // "a   b"
wrapped := runewidthex.Wrap("some long text", 40)
truncated := runewidthex.Truncate("long text", 8, "…")
padded := runewidthex.PadRight("left", 10)
```

For custom configuration use a `Condition`:

```go
c := &runewidthex.Condition{
    TabWidth: 8,
    Options:  displaywidth.Options{EastAsianWidth: true},
}
expanded := c.ExpandTab("a\tb")
wrapped  := c.Wrap(s, 80)
```

## API

### `Condition`

```go
type Condition struct {
    Options  displaywidth.Options // EastAsianWidth, ControlSequences, …
    TabWidth int                  // tab-stop interval, default 4
}
```

### Methods

| Method | Description |
|--------|-------------|
| `StringWidth(s) int` | Display width of s (max line width for multi-line strings) |
| `ExpandTab(s) string` | Replace tabs with spaces aligned to tab stops |
| `Wrap(s, w) string` | Wrap s so no line exceeds w cells; tabs are indivisible tokens |
| `Truncate(s, w, tail) string` | Truncate s to w cells, appending tail if truncated |
| `PadRight(s, w) string` | Right-pad s with spaces to reach w cells |
| `PadLeft(s, w) string` | Left-pad s with spaces to reach w cells |

Package-level convenience functions (`StringWidth`, `ExpandTab`, `Wrap`, `Truncate`, `PadRight`, `PadLeft`) call the corresponding method on a default `Condition` (tab width 4).

## License

MIT
