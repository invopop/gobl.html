package bill

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
)

// longWordRunes is the number of characters above which a word, meaning a run
// of text without spaces, gets extra wrap points in a line's description cell.
// Ordinary words are shorter, so they still wrap only at spaces.
const longWordRunes = 30

// longWordBreakAfter lists the characters after which a long word may wrap,
// so that a list such as "SN001,SN002,SN003" wraps between its items.
const longWordBreakAfter = `,;:/\|-_.+&=?`

// zeroWidthJoiner joins emoji into one symbol, so a word is never split
// next to it.
const zeroWidthJoiner = '\u200d'

// rawTextTags hold text that the HTML tokenizer doesn't read as markup, such
// as scripts and styles, so wrapLongWordsHTML leaves it alone.
var rawTextTags = map[string]bool{
	"iframe": true, "noembed": true, "noframes": true, "noscript": true, "plaintext": true,
	"script": true, "style": true, "textarea": true, "title": true, "xmp": true,
}

var longWordRegexp = regexp.MustCompile(fmt.Sprintf(`[^\s\p{Z}]{%d,}`, longWordRunes+1))

// selfWrappingScripts are written without spaces between words. The renderer
// already finds wrap points in them, so their characters are never counted.
var selfWrappingScripts = []*unicode.RangeTable{
	unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul,
	unicode.Thai, unicode.Lao, unicode.Khmer, unicode.Myanmar,
}

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

// wrapLongWordsHTML adds the wrap points of wrappable to the text of an HTML
// fragment, such as a rendered note. Tags, and text with no long words, are
// kept byte for byte.
func wrapLongWordsHTML(src string) string {
	z := html.NewTokenizer(strings.NewReader(src))
	var b strings.Builder
	inRawText := false
	read := 0
	for {
		tt := z.Next()
		raw := string(z.Raw()) // copied first: Text() unescapes in place
		out := raw
		switch tt {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return src
			}
			// keep what the tokenizer drops at the end, such as an unfinished tag
			return b.String() + src[read:]
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := z.TagName()
			inRawText = rawTextTags[string(name)]
		case html.EndTagToken:
			inRawText = false
		case html.TextToken:
			if inRawText {
				break
			}
			if parts := splitLongWords(string(z.Text())); len(parts) > 1 {
				out = joinWithWbr(parts)
			}
		}
		read += len(raw)
		b.WriteString(out)
	}
}

// splitLongWords splits s at the points where a long word may wrap, so the
// caller can print a <wbr> between the parts. Joined together, the parts
// give back s. A long word is split after each longWordBreakAfter character
// and after every longWordRunes characters that contain none of them.
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
		run := 0
		for i := loc[0]; i < loc[1]; {
			r, size := utf8.DecodeRuneInString(s[i:])
			i += size
			if i == loc[1] {
				break
			}
			run++
			if unicode.In(r, selfWrappingScripts...) {
				run = 0
				continue
			}
			next, _ := utf8.DecodeRuneInString(s[i:])
			if canSplitBetween(r, next, run) {
				parts = append(parts, s[from:i])
				from = i
				run = 0
			}
		}
	}
	return append(parts, s[from:])
}

// canSplitBetween reports whether a long word may wrap between r and next,
// where run counts the characters since the last split. It splits after a
// separator or after longWordRunes characters, but not before a separator,
// so no line starts with one, unless it follows another separator after
// longWordRunes characters (so a line of dashes can wrap). It never splits
// inside a character drawn from several runes.
func canSplitBetween(r, next rune, run int) bool {
	if joinsCharacter(r, next) {
		return false
	}
	if isBreakAfter(next) {
		return isBreakAfter(r) && run >= longWordRunes
	}
	return isBreakAfter(r) || run >= longWordRunes
}

func isBreakAfter(r rune) bool {
	return strings.ContainsRune(longWordBreakAfter, r)
}

// joinsCharacter reports whether r and next belong to one character as the
// reader sees it: a letter and its accent or vowel sign, an emoji and its
// skin tone or variation, emoji joined by a zero-width joiner, or the two
// halves of a flag.
func joinsCharacter(r, next rune) bool {
	return unicode.Is(unicode.M, next) ||
		r == zeroWidthJoiner || next == zeroWidthJoiner ||
		(next >= 0x1f3fb && next <= 0x1f3ff) ||
		(isRegionalIndicator(r) && isRegionalIndicator(next))
}

func isRegionalIndicator(r rune) bool {
	return r >= 0x1f1e6 && r <= 0x1f1ff
}
