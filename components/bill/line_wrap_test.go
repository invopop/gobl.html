package bill

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/invopop/gobl/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitLongWords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "empty",
			in:   "",
			want: []string{""},
		},
		{
			name: "ordinary text is left whole",
			in:   "Preventive maintenance visits, remote troubleshooting and firmware upgrades.",
			want: []string{"Preventive maintenance visits, remote troubleshooting and firmware upgrades."},
		},
		{
			name: "a word at the limit is left whole",
			in:   "abc " + strings.Repeat("x", longWordRunes),
			want: []string{"abc " + strings.Repeat("x", longWordRunes)},
		},
		{
			name: "a list is split after each separator",
			in:   "Serials: SN0001,SN0002,SN0003;SN0004/SN0005|SN0006 end",
			want: []string{"Serials: SN0001,", "SN0002,", "SN0003;", "SN0004/", "SN0005|", "SN0006 end"},
		},
		{
			name: "a trailing separator is not split",
			in:   "SN00000001,SN00000002,SN00000003,",
			want: []string{"SN00000001,", "SN00000002,", "SN00000003,"},
		},
		{
			name: "a word without separators is split every longWordRunes characters",
			in:   strings.Repeat("0123456789", 7),
			want: []string{
				strings.Repeat("0123456789", 3),
				strings.Repeat("0123456789", 3),
				"0123456789",
			},
		},
		{
			name: "characters are counted as runes",
			in:   strings.Repeat("é", longWordRunes+1),
			want: []string{strings.Repeat("é", longWordRunes), "é"},
		},
		{
			name: "characters are counted from the last split",
			in:   strings.Repeat("x", 25) + "," + strings.Repeat("y", 35),
			want: []string{strings.Repeat("x", 25) + ",", strings.Repeat("y", 30), strings.Repeat("y", 5)},
		},
		{
			name: "no line starts with a separator",
			in:   strings.Repeat("x", 30) + ",yyyyy",
			want: []string{strings.Repeat("x", 30) + ",", "yyyyy"},
		},
		{
			name: "repeated separators stay together",
			in:   "SN0000000001,,SN0000000002,,SN0000000003",
			want: []string{"SN0000000001,,", "SN0000000002,,", "SN0000000003"},
		},
		{
			name: "a letter is not split from its accent",
			in:   "a" + strings.Repeat("e\u0301", 20),
			want: []string{"a" + strings.Repeat("e\u0301", 15), strings.Repeat("e\u0301", 5)},
		},
		{
			name: "an emoji sequence is not split",
			in:   strings.Repeat("x", 29) + "👩\u200d💻" + "yyy",
			want: []string{strings.Repeat("x", 29) + "👩\u200d💻", "yyy"},
		},
		{
			name: "an emoji is not split from its skin tone",
			in:   strings.Repeat("x", 29) + "\U0001F44D\U0001F3FD" + "yyyy",
			want: []string{strings.Repeat("x", 29) + "\U0001F44D\U0001F3FD", "yyyy"},
		},
		{
			name: "a flag is not split",
			in:   strings.Repeat("x", 29) + "\U0001F1EB\U0001F1F7" + "yy",
			want: []string{strings.Repeat("x", 29) + "\U0001F1EB\U0001F1F7", "yy"},
		},
		{
			name: "a run of separators is split every longWordRunes characters",
			in:   strings.Repeat("_", 40),
			want: []string{strings.Repeat("_", 30), strings.Repeat("_", 10)},
		},
		{
			name: "non-breaking spaces end a word",
			in:   "Prix\u00a0:\u00a0" + strings.Repeat("a", 31),
			want: []string{"Prix\u00a0:\u00a0" + strings.Repeat("a", 30), "a"},
		},
		{
			name: "scripts written without spaces are left to the renderer",
			in:   strings.Repeat("漢字", 20) + " " + strings.Repeat("ภาษาไทย", 6),
			want: []string{strings.Repeat("漢字", 20) + " " + strings.Repeat("ภาษาไทย", 6)},
		},
		{
			name: "only the long word is split",
			in:   "short words " + strings.Repeat("ab,", 12) + " short words",
			want: slices.Concat(
				[]string{"short words ab,"},
				slices.Repeat([]string{"ab,"}, 10),
				[]string{"ab, short words"},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLongWords(tt.in)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.in, strings.Join(got, ""))
		})
	}
}

func TestWrappable(t *testing.T) {
	render := func(s string) string {
		t.Helper()
		var b strings.Builder
		require.NoError(t, wrappable(s).Render(context.Background(), &b))
		return b.String()
	}

	t.Run("puts a wbr between the parts of a long word", func(t *testing.T) {
		assert.Equal(t, "SN00000001,<wbr>SN00000002,<wbr>SN00000003", render("SN00000001,SN00000002,SN00000003"))
	})

	t.Run("escapes the text", func(t *testing.T) {
		assert.Equal(t, "Bolts &amp; nuts &lt;M8&gt;", render("Bolts & nuts <M8>"))
	})
}

func TestWrapLongWordsHTML(t *testing.T) {
	t.Run("keeps HTML without long words byte for byte", func(t *testing.T) {
		in := "<p>Delivered on &quot;Monday&quot; &amp; installed.<br>\n<code>X-1</code></p>\n"
		assert.Equal(t, in, wrapLongWordsHTML(in))
	})

	t.Run("adds wbr to long words in the text", func(t *testing.T) {
		assert.Equal(t,
			"<p>Serials: SN00000001,<wbr>SN00000002,<wbr>SN00000003 &amp; more</p>\n",
			wrapLongWordsHTML("<p>Serials: SN00000001,SN00000002,SN00000003 &amp; more</p>\n"),
		)
	})

	t.Run("leaves attributes alone", func(t *testing.T) {
		url := "https://example.com/catalogue/items/0123456789"
		assert.Equal(t,
			`<p><a href="`+url+`">https://<wbr>example.<wbr>com/<wbr>catalogue/<wbr>items/<wbr>0123456789</a></p>`,
			wrapLongWordsHTML(`<p><a href="`+url+`">`+url+`</a></p>`),
		)
	})

	t.Run("keeps an unfinished tag at the end", func(t *testing.T) {
		assert.Equal(t, "abc<p", wrapLongWordsHTML("abc<p"))
		assert.Equal(t, "<p>text</p", wrapLongWordsHTML("<p>text</p"))
		assert.Equal(t,
			"<p>SN00000001,<wbr>SN00000002,<wbr>SN00000003</p",
			wrapLongWordsHTML("<p>SN00000001,SN00000002,SN00000003</p"),
		)
	})

	t.Run("leaves text that isn't markup alone", func(t *testing.T) {
		long := "a&amp;b,a&amp;b,a&amp;b,a&amp;b,a&amp;b,a&amp;b,a&amp;b"
		for _, in := range []string{
			"<style>.note{color:red;background:blue;border:none}</style><p>ok</p>",
			"<noscript>" + long + "</noscript><p>ok</p>",
			"<textarea>" + long + "</textarea><p>ok</p>",
			"<script/><p>" + long + "</p></script><p>ok</p>",
		} {
			assert.Equal(t, in, wrapLongWordsHTML(in))
		}
	})

	t.Run("wraps a rendered line note", func(t *testing.T) {
		n := &org.Note{Text: "Returned: SN00000001,SN00000002,SN00000003"}
		assert.Equal(t,
			"<p>Returned: SN00000001,<wbr>SN00000002,<wbr>SN00000003</p>\n",
			wrapLongWordsHTML(renderNote(n)),
		)
	})
}
