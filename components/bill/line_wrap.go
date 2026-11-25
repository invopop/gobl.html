package bill

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/a-h/templ"
)

// longWordRunes is the number of characters above which a word, meaning a run
// of text without spaces, gets extra wrap points in a line's description cell.
// Ordinary words are shorter, so they still wrap only at spaces.
const longWordRunes = 30

// longWordBreakAfter lists the characters after which a long word may wrap,
// so that a list such as "SN001,SN002,SN003" wraps between its items.
const longWordBreakAfter = `,;:/\|-_.+&=?`

var longWordRegexp = regexp.MustCompile(fmt.Sprintf(`[^\s\p{Z}]{%d,}`, longWordRunes+1))

// wrappable prints s with a <wbr> where splitLongWords splits it, so a long
// word, such as a list of serial numbers, wraps inside the description cell.
// It is plain Go because a templ loop would print a space before each part.
func wrappable(s string) templ.Component {
	return templ.Raw(joinWithWbr(splitLongWords(s)))
}

// joinWithWbr escapes each part for HTML and puts a <wbr> between them.
func joinWithWbr(parts []string) string {
	for i, p := range parts {
		parts[i] = templ.EscapeString(p)
	}
	return strings.Join(parts, "<wbr>")
}

// splitLongWords splits s at the points where a long word may wrap, so the
// caller can print a <wbr> between the parts. Joined together, the parts
// give back s. A long word is split right after each run of
// longWordBreakAfter characters, never before one, so no line starts with a
// separator, and never after the separators that start the word, so none
// ends with one on its own. A long word without any separator is left whole.
//
// The lines table sizes its columns from their content, so a long word with
// no wrap points sets the width of the description column and pushes the
// other columns off the page. The CSS wrapping properties don't fix this in
// PrinceXML, the PDF renderer: it supports neither `overflow-wrap: anywhere`
// nor `word-break: break-word`, `overflow-wrap: break-word` doesn't narrow a
// table column, and `word-break: break-all` would also split ordinary words.
func splitLongWords(s string) []string {
	var parts []string
	from := 0
	for _, loc := range longWordRegexp.FindAllStringIndex(s, -1) {
		// The separators are ASCII, so a byte never matches inside a
		// multi-byte character. Separators that start the word, such as the
		// "+" of a phone number, stay with what follows them.
		started := false
		for i := loc[0]; i < loc[1]-1; i++ {
			if !isBreakAfter(s[i]) {
				started = true
			} else if started && !isBreakAfter(s[i+1]) {
				parts = append(parts, s[from:i+1])
				from = i + 1
			}
		}
	}
	return append(parts, s[from:])
}

func isBreakAfter(b byte) bool {
	return strings.IndexByte(longWordBreakAfter, b) >= 0
}
