package bill

import (
	"context"
	"slices"
	"strings"
	"testing"

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
			in:   "abc " + strings.Repeat("ab,", longWordRunes/3),
			want: []string{"abc " + strings.Repeat("ab,", longWordRunes/3)},
		},
		{
			name: "the limit counts characters, not bytes",
			in:   strings.Repeat("é,", longWordRunes/2),
			want: []string{strings.Repeat("é,", longWordRunes/2)},
		},
		{
			name: "a list is split after each separator",
			in:   "Serials: SN0001,SN0002,SN0003;SN0004/SN0005|SN0006 end",
			want: []string{"Serials: SN0001,", "SN0002,", "SN0003;", "SN0004/", "SN0005|", "SN0006 end"},
		},
		{
			name: "a reference is split after each underscore",
			in:   "Notebook_Model_X5_Processor_Core_i5_16GB",
			want: []string{"Notebook_", "Model_", "X5_", "Processor_", "Core_", "i5_", "16GB"},
		},
		{
			name: "a trailing separator is not split",
			in:   "SN00000001,SN00000002,SN00000003,",
			want: []string{"SN00000001,", "SN00000002,", "SN00000003,"},
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
			name: "a long word without separators is left whole",
			in:   strings.Repeat("0123456789", 7),
			want: []string{strings.Repeat("0123456789", 7)},
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
