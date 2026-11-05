package display_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/invopop/gobl.html/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	assert.True(t, display.KeyPayment.Valid())
	assert.False(t, display.Key("totals").Valid())
	assert.False(t, display.KeyLinesSublines.Valid(), "sublines repeat per line and are hide-only")
	for _, k := range display.Keys() {
		assert.True(t, k.Valid())
	}
	assert.True(t, display.ModeAfter.Valid())
	assert.False(t, display.Mode("around").Valid())
}

func TestVarNames(t *testing.T) {
	assert.Equal(t, []string{
		"Code", "Series", "Number", "IssueDate", "DueDate", "TotalPayable", "TotalDue",
		"Currency", "SupplierName", "CustomerName", "PaymentLink",
	}, display.VarNames())
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		snippets display.Snippets
		errs     []string
	}{
		{
			name: "valid set",
			snippets: display.Snippets{
				{At: display.KeyPaymentInstructions, Mode: display.ModeAfter, HTML: `<a href="{{ .PaymentLink }}">Pay {{ .TotalDue }}</a>`},
				{At: display.KeyNotes, Mode: display.ModeReplace, HTML: `{{ if .DueDate }}<p>Due {{ .DueDate }}</p>{{ else }}<p>{{ $.Code }}</p>{{ end }}`},
				{At: display.KeyNotes, Mode: display.ModeBefore, HTML: `<p>{{ printf "%s-%s" .Series .Code }}</p>`},
			},
		},
		{name: "empty", snippets: nil},
		{
			name:     "unknown key",
			snippets: display.Snippets{{At: "totals", Mode: display.ModeAfter, HTML: "x"}},
			errs:     []string{"0: at: unknown key 'totals'"},
		},
		{
			name:     "sublines have no anchor",
			snippets: display.Snippets{{At: display.KeyLinesSublines, Mode: display.ModeAfter, HTML: "x"}},
			errs:     []string{"0: at: 'lines.sublines' repeats per line"},
		},
		{
			name:     "unknown mode",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: "around", HTML: "x"}},
			errs:     []string{"0: mode: unknown mode 'around'"},
		},
		{
			name: "duplicate replace",
			snippets: display.Snippets{
				{At: display.KeyNotes, Mode: display.ModeReplace, HTML: "a"},
				{At: display.KeyNotes, Mode: display.ModeReplace, HTML: "b"},
			},
			errs: []string{"1: duplicate replace at 'notes'"},
		},
		{
			name:     "missing snippet",
			snippets: display.Snippets{nil},
			errs:     []string{"0: missing"},
		},
		{
			name:     "snippet too large",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("x", display.MaxSnippetSize+1)}},
			errs:     []string{"0: html: exceeds 32768 bytes"},
		},
		{
			name: "total too large",
			snippets: display.Snippets{
				{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("a", display.MaxSnippetSize)},
				{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("b", display.MaxSnippetSize)},
				{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("c", display.MaxSnippetSize)},
				{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("d", display.MaxSnippetSize)},
				{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "e"},
			},
			errs: []string{"total html exceeds 131072 bytes"},
		},
		{
			name:     "unknown variable",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Total }}</p>"}},
			errs:     []string{"0: html: unknown variable 'Total'"},
		},
		{
			name:     "unknown variable inside a branch",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "{{ if .PaymentLink }}{{ .Bogus }}{{ end }}"}},
			errs:     []string{"unknown variable 'Bogus'"},
		},
		{
			name:     "unknown variable through the root",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "{{ with .Code }}{{ $.Bogus }}{{ end }}"}},
			errs:     []string{"unknown variable 'Bogus'"},
		},
		{
			name:     "variables have no fields",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "{{ .Code.Foo }}"}},
			errs:     []string{"variable 'Code' has no field 'Foo'"},
		},
		{
			name:     "bad template syntax",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Code </p>"}},
			errs:     []string{"0: html:"},
		},
		{
			name:     "markup that cannot be escaped",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: `<a href="{{ .PaymentLink }}>open`}},
			errs:     []string{"0: html:"},
		},
		{
			name: "several errors are all reported",
			snippets: display.Snippets{
				{At: "nope", Mode: "around", HTML: "{{ .Nope }}"},
			},
			errs: []string{"0: at: unknown key 'nope'", "0: mode: unknown mode 'around'", "0: html: unknown variable 'Nope'"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.snippets.Validate()
			if len(tt.errs) == 0 {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, e := range tt.errs {
				assert.Contains(t, err.Error(), e)
			}
		})
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{`<p>a</p><script>alert(1)</script><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><SCRIPT type="x">
		multi
		line</Script ><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><iframe src="https://evil"></iframe><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><object data="x.swf"><param name="y"></object><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><embed src="x.svg"><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><embed src="x.svg" /><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>a</p><script>never closed<p>b</p>`, `<p>a</p>`},
		{`<p>a</p></script><p>b</p>`, `<p>a</p><p>b</p>`},
		{`<p>scripted</p><p class="embedded">ok</p>`, `<p>scripted</p><p class="embedded">ok</p>`},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.out, display.Strip(tt.in), tt.in)
	}
}

func TestRender(t *testing.T) {
	vars := &display.Vars{
		Code:         "0042",
		Series:       "SAMPLE",
		IssueDate:    "2024-01-31",
		DueDate:      "2024-02-29",
		TotalPayable: "€1,210.00",
		TotalDue:     "€847.00",
		Currency:     "EUR",
		SupplierName: `Provide <One> & "Co"`,
		CustomerName: "Sample Consumer",
		PaymentLink:  `https://pay.example.com/i/42?ref=A&B&x="y"`,
	}

	t.Run("groups by anchor and mode in order", func(t *testing.T) {
		r, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>after-1</p>"},
			{At: display.KeyNotes, Mode: display.ModeBefore, HTML: "<p>before</p>"},
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>after-2</p>"},
			{At: display.KeyTaxes, Mode: display.ModeReplace, HTML: "<p>taxes</p>"},
		}.Render(vars)
		require.NoError(t, err)
		assert.Equal(t, []string{"<p>before</p>"}, r.Before(display.KeyNotes))
		assert.Equal(t, []string{"<p>after-1</p>", "<p>after-2</p>"}, r.After(display.KeyNotes))
		rep, ok := r.Replacement(display.KeyTaxes)
		assert.True(t, ok)
		assert.Equal(t, "<p>taxes</p>", rep)
		_, ok = r.Replacement(display.KeyNotes)
		assert.False(t, ok)
		assert.True(t, r.Has(display.KeyNotes))
		assert.True(t, r.Has(display.KeyTaxes))
		assert.False(t, r.Has(display.KeyPayment))
	})

	t.Run("escapes variables for their context", func(t *testing.T) {
		r, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: `<p title="{{ .SupplierName }}">{{ .SupplierName }}</p><a href="{{ .PaymentLink }}">Pay {{ .TotalDue }}</a>`},
		}.Render(vars)
		require.NoError(t, err)
		out := r.After(display.KeyNotes)[0]
		assert.NotContains(t, out, "<One>")
		assert.Contains(t, out, `>Provide &lt;One&gt; &amp; &#34;Co&#34;</p>`)
		assert.Contains(t, out, `title="Provide &lt;One&gt; &amp; &#34;Co&#34;"`)
		assert.Contains(t, out, `href="https://pay.example.com/i/42?ref=A&amp;B&amp;x=%22y%22"`)
		assert.Contains(t, out, "Pay €847.00")
	})

	t.Run("refuses unsafe URL schemes in links", func(t *testing.T) {
		r, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: `<a href="{{ .PaymentLink }}">x</a>`},
		}.Render(&display.Vars{PaymentLink: "javascript:alert(1)"})
		require.NoError(t, err)
		assert.NotContains(t, r.After(display.KeyNotes)[0], "javascript:")
	})

	t.Run("conditionals and empty variables", func(t *testing.T) {
		r, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: `{{ if .PaymentLink }}<a href="{{ .PaymentLink }}">pay</a>{{ else }}<p>{{ .Series }}-{{ .Code }}</p>{{ end }}`},
		}.Render(&display.Vars{Code: "1", Series: "S"})
		require.NoError(t, err)
		assert.Equal(t, "<p>S-1</p>", r.After(display.KeyNotes)[0])
	})

	t.Run("strips elements before rendering", func(t *testing.T) {
		r, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: `<p>{{ .Code }}</p><script>{{ .Code }}</script><embed src="x">`},
		}.Render(vars)
		require.NoError(t, err)
		assert.Equal(t, "<p>0042</p>", r.After(display.KeyNotes)[0])
	})

	t.Run("skips unknown keys and modes", func(t *testing.T) {
		r, err := display.Snippets{
			{At: "future.key", Mode: display.ModeAfter, HTML: "<p>x</p>"},
			{At: display.KeyNotes, Mode: "around", HTML: "<p>x</p>"},
			{At: display.KeyLinesSublines, Mode: display.ModeAfter, HTML: "<p>x</p>"},
		}.Render(vars)
		require.NoError(t, err)
		assert.False(t, r.Has("future.key"))
		assert.False(t, r.Has(display.KeyNotes))
		assert.False(t, r.Has(display.KeyLinesSublines))
	})

	t.Run("fails on unknown variables", func(t *testing.T) {
		_, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Total }}</p>"},
		}.Render(vars)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "snippets: 0: html: unknown variable 'Total'")
	})

	t.Run("fails on oversized snippets", func(t *testing.T) {
		_, err := display.Snippets{
			{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("x", display.MaxSnippetSize+1)},
		}.Render(vars)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds 32768 bytes")
	})

	t.Run("empty set and nil vars", func(t *testing.T) {
		r, err := display.Snippets{}.Render(nil)
		require.NoError(t, err)
		assert.Nil(t, r)
		assert.False(t, r.Has(display.KeyNotes))
		assert.Nil(t, r.Before(display.KeyNotes))
		r, err = display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Code }}</p>"}}.Render(nil)
		require.NoError(t, err)
		assert.Equal(t, "<p></p>", r.After(display.KeyNotes)[0])
	})
}

func TestSnippetsJSON(t *testing.T) {
	const list = `[{"at":"payment.instructions","mode":"after","html":"<p>{{ .Code }}</p>"}]`
	want := display.Snippets{{At: display.KeyPaymentInstructions, Mode: display.ModeAfter, HTML: "<p>{{ .Code }}</p>"}}

	var s display.Snippets
	require.NoError(t, json.Unmarshal([]byte(list), &s))
	assert.Equal(t, want, s)

	var wrapped struct {
		Snippets display.Snippets `json:"snippets"`
	}
	quoted, err := json.Marshal(list)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(`{"snippets":`+string(quoted)+`}`), &wrapped), "a JSON string holding the array")
	assert.Equal(t, want, wrapped.Snippets)

	require.NoError(t, json.Unmarshal([]byte(`{"snippets":""}`), &wrapped))
	assert.Nil(t, wrapped.Snippets, "an empty string clears the list")

	assert.Error(t, json.Unmarshal([]byte(`{"snippets":"not json"}`), &wrapped))
	assert.Error(t, json.Unmarshal([]byte(`{"snippets":42}`), &wrapped))

	var p display.Snippets
	require.NoError(t, p.UnmarshalParam(list))
	assert.Equal(t, want, p)
	require.NoError(t, p.UnmarshalParam(" "))
	assert.Nil(t, p)

	b, err := json.Marshal(want)
	require.NoError(t, err)
	assert.JSONEq(t, list, string(b))
}
