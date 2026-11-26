package goblhtml_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	snippetsFullExample         = "es-verifactu-invoice-full.json"
	snippetsDeliveryExample     = "es-verifactu-invoice-delivery.json"
	snippetsTermsExample        = "fr-ctc-invoice-b2b.json"
	snippetsOrderExample        = "pt-at-order-full.json"
	snippetsDeliveryNoteExample = "pt-at-delivery-note-full.json"
)

func renderSnippets(t *testing.T, env *gobl.Envelope, s display.Snippets, opts ...goblhtml.Option) string {
	t.Helper()
	require.NoError(t, s.Validate())
	opts = append(opts, goblhtml.WithSnippets(s))
	out, err := goblhtml.Render(context.Background(), env, opts...)
	require.NoError(t, err)
	return string(out)
}

func loadSnippetsExample(t *testing.T, name string) *gobl.Envelope {
	t.Helper()
	env, err := loadExample(name)
	require.NoError(t, err)
	return env
}

const taxesMarker = `class="taxes"`

func wrapped(mode display.Mode, key display.Key, html string) string {
	return `<div class="snippet snippet-` + string(mode) + `" data-at="` + string(key) + `">` + html + `</div>`
}

func wrappedItem(mode display.Mode, key display.Key, html string) string {
	return `<li class="snippet snippet-` + string(mode) + `" data-at="` + string(key) + `">` + html + `</li>`
}

func wrappedInline(mode display.Mode, key display.Key, html string) string {
	return `<span class="snippet snippet-` + string(mode) + `" data-at="` + string(key) + `">` + html + `</span>`
}

// TestSnippetAnchors places a before and an after snippet at every anchor
// on a document that has native content there, and checks they land on
// either side of it.
func TestSnippetAnchors(t *testing.T) {
	tests := []struct {
		key     display.Key
		example string
		marker  string // native content between the two snippets
		opts    []goblhtml.Option
		item    bool
		inline  bool
	}{
		{key: display.KeyHeaderLogo, example: snippetsFullExample, marker: `class="alias"`},
		{key: display.KeyHeaderTitleBadges, example: snippetsFullExample, marker: "BADGE-MARKER", opts: []goblhtml.Option{goblhtml.WithLabel("BADGE-MARKER")}, inline: true},
		{key: display.KeyHeaderSummaryOrdering, example: snippetsFullExample, marker: `class="ordering-code"`, item: true},
		{key: display.KeyHeaderContactsDelivery, example: snippetsDeliveryExample, marker: `class="receiver"`},
		{key: display.KeyTaxes, example: snippetsFullExample, marker: taxesMarker},
		{key: display.KeyPayment, example: snippetsFullExample, marker: `<section class="payment">`},
		{key: display.KeyPaymentInstructions, example: snippetsFullExample, marker: `class="instructions"`},
		{key: display.KeyPaymentBankData, example: snippetsFullExample, marker: `class="bank-data"`, item: true},
		{key: display.KeyPaymentOnline, example: "es-verifactu-invoice-online-payment.json", marker: `class="online"`, item: true},
		{key: display.KeyPaymentTerms, example: snippetsTermsExample, marker: `class="terms"`},
		{key: display.KeyNotes, example: snippetsFullExample, marker: "Thank you for your business!"},
		{key: display.KeyFooterBranding, example: snippetsFullExample, marker: `class="gobl-logo"`},
		// The anchors are shared by every document kind.
		{key: display.KeyHeaderSummaryOrdering, example: snippetsOrderExample, marker: `class="contract"`, item: true},
		{key: display.KeyTaxes, example: snippetsOrderExample, marker: taxesMarker},
		{key: display.KeyPayment, example: snippetsOrderExample, marker: `<section class="payment">`},
		{key: display.KeyPaymentInstructions, example: snippetsOrderExample, marker: `class="instructions"`},
		{key: display.KeyPaymentTerms, example: snippetsOrderExample, marker: `class="terms"`},
		{key: display.KeyNotes, example: snippetsOrderExample, marker: `<section class="notes">`},
		{key: display.KeyHeaderContactsDelivery, example: snippetsDeliveryNoteExample, marker: `class="receiver"`},
		{key: display.KeyHeaderSummaryOrdering, example: snippetsDeliveryNoteExample, marker: `class="ordering-code"`, item: true},
		{key: display.KeyTaxes, example: snippetsDeliveryNoteExample, marker: taxesMarker},
		{key: display.KeyNotes, example: snippetsDeliveryNoteExample, marker: `<section class="notes">`},
	}
	for _, tt := range tests {
		t.Run(tt.example+"/"+string(tt.key), func(t *testing.T) {
			env := loadSnippetsExample(t, tt.example)
			if tt.key == display.KeyPaymentOnline {
				// Links without a QR code are listed in the instructions.
				inv := env.Extract().(*bill.Invoice)
				for _, ol := range inv.Payment.Instructions.Online {
					ol.Meta = cbc.Meta{"qr-size": "0"}
				}
			}
			out := renderSnippets(t, env, display.Snippets{
				{At: tt.key, Mode: display.ModeAfter, HTML: "<b>after-marker</b>"},
				{At: tt.key, Mode: display.ModeBefore, HTML: "<b>before-marker</b>"},
			}, tt.opts...)
			wrap := wrapped
			switch {
			case tt.item:
				wrap = wrappedItem
			case tt.inline:
				wrap = wrappedInline
			}
			before := strings.Index(out, wrap(display.ModeBefore, tt.key, "<b>before-marker</b>"))
			native := strings.Index(out, tt.marker)
			after := strings.Index(out, wrap(display.ModeAfter, tt.key, "<b>after-marker</b>"))
			require.True(t, before >= 0, "before snippet missing")
			require.True(t, native >= 0, "native content missing")
			require.True(t, after >= 0, "after snippet missing")
			assert.True(t, before < native && native < after, "order: before %d, native %d, after %d", before, native, after)
			// The title repeats in the running header of later pages and the
			// footer has a print and a screen variant, so a snippet renders
			// as often as its anchor does, and no more.
			assert.Equal(t, strings.Count(out, tt.marker), strings.Count(out, "before-marker"), "a snippet renders once per anchor")
		})
	}
}

func TestSnippetReplace(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPayment, Mode: display.ModeReplace, HTML: "<p>Pay via portal</p>"},
		// Snippets inside a replaced section are never rendered.
		{At: display.KeyPaymentTerms, Mode: display.ModeAfter, HTML: "<p>terms-after</p>"},
		{At: display.KeyPaymentBankData, Mode: display.ModeBefore, HTML: "<p>bank-before</p>"},
	})
	assert.NotContains(t, out, `<section class="payment">`)
	assert.NotContains(t, out, "ES31 0240 4036 4573 9340 5564")
	assert.NotContains(t, out, "terms-after")
	assert.NotContains(t, out, "bank-before")
	assert.Contains(t, out, wrapped(display.ModeReplace, display.KeyPayment, "<p>Pay via portal</p>"))
	assert.Contains(t, out, `class="totals"`)
}

func TestSnippetReplaceInstructionsCoversQRCodes(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	plain := renderSnippets(t, env, nil)
	require.Contains(t, plain, `class="online-qrs"`)

	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPaymentInstructions, Mode: display.ModeReplace, HTML: "<p>See portal</p>"},
	})
	assert.NotContains(t, out, `class="instructions"`)
	assert.NotContains(t, out, `class="online-qrs"`)
	assert.Contains(t, out, wrapped(display.ModeReplace, display.KeyPaymentInstructions, "<p>See portal</p>"))

	// Replacing the online links keeps them out of the QR column too.
	out = renderSnippets(t, env, display.Snippets{
		{At: display.KeyPaymentOnline, Mode: display.ModeReplace, HTML: "pay-elsewhere"},
	})
	assert.NotContains(t, out, `class="online-qrs"`)
	assert.NotContains(t, out, `class="online"`)
	assert.Contains(t, out, wrappedItem(display.ModeReplace, display.KeyPaymentOnline, "pay-elsewhere"))
	assert.Contains(t, out, `class="instructions"`)
}

func TestSnippetsAtHiddenNotesStayVisible(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyNotes, Mode: display.ModeBefore, HTML: "<p>notes-before</p>"},
		{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>notes-after</p>"},
	}, goblhtml.WithHide("notes"))
	assert.Contains(t, out, `section.notes:not(.fixed), section.notes .note:not(.fixed), section.notes .snippet-replace { display: none !important; }`)
	before := strings.Index(out, wrapped(display.ModeBefore, display.KeyNotes, "<p>notes-before</p>"))
	section := strings.Index(out, `<section class="notes`)
	end := strings.Index(out, `</section>`+wrapped(display.ModeAfter, display.KeyNotes, "<p>notes-after</p>"))
	require.True(t, before >= 0 && section >= 0 && end >= 0, "before %d, section %d, after %d", before, section, end)
	assert.True(t, before < section && section < end, "the snippets sit outside the hidden section")
}

func TestSnippetReplaceWithoutNativeContent(t *testing.T) {
	env := loadSnippetsExample(t, snippetsDeliveryExample)
	inv := env.Extract().(*bill.Invoice)
	require.Nil(t, inv.Payment)
	require.Empty(t, inv.Notes)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPayment, Mode: display.ModeReplace, HTML: "<p>Pay via portal</p>"},
		{At: display.KeyNotes, Mode: display.ModeReplace, HTML: "<p>Custom notes</p>"},
	})
	assert.Contains(t, out, wrapped(display.ModeReplace, display.KeyPayment, "<p>Pay via portal</p>"))
	assert.Contains(t, out, `<section class="notes">`)
	assert.Contains(t, out, wrapped(display.ModeReplace, display.KeyNotes, "<p>Custom notes</p>"))
}

func TestHiddenNotesKeepTaxNotes(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	inv := env.Extract().(*bill.Invoice)
	require.NotNil(t, inv.Tax)
	require.NotEmpty(t, inv.Notes)
	inv.Tax.Notes = append(inv.Tax.Notes, &tax.Note{Text: "Tax note marker"})
	out := renderSnippets(t, env, nil, goblhtml.WithHide("notes"))
	// The rule hides the regular notes, and the section only when it has
	// no tax notes; the tax notes are marked so the rule leaves them.
	assert.Contains(t, out, `<section class="notes fixed">`)
	assert.Contains(t, out, `<div class="note fixed"><p>Tax note marker</p>`)
	assert.Contains(t, out, `section.notes:not(.fixed), section.notes .note:not(.fixed)`)
}

func TestSnippetsInsideEmptyPaymentList(t *testing.T) {
	env := loadSnippetsExample(t, snippetsDeliveryExample)
	require.Nil(t, env.Extract().(*bill.Invoice).Payment)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPaymentBankData, Mode: display.ModeBefore, HTML: "IBAN ES00"},
		{At: display.KeyPaymentOnline, Mode: display.ModeReplace, HTML: "pay online"},
	})
	assert.Contains(t, out, `<section class="payment">`)
	assert.Contains(t, out, `<ul class="instructions">`+wrappedItem(display.ModeBefore, display.KeyPaymentBankData, "IBAN ES00"))
	assert.Contains(t, out, wrappedItem(display.ModeReplace, display.KeyPaymentOnline, "pay online"))
}

func TestSnippetsInsideEmptyPayment(t *testing.T) {
	env := loadSnippetsExample(t, snippetsDeliveryExample)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPaymentTerms, Mode: display.ModeBefore, HTML: "<p>Terms intro</p>"},
		{At: display.KeyPaymentInstructions, Mode: display.ModeAfter, HTML: "<p>Pay now</p>"},
	})
	assert.Contains(t, out, `<section class="payment">`)
	assert.Contains(t, out, wrapped(display.ModeBefore, display.KeyPaymentTerms, "<p>Terms intro</p>"))
	assert.Contains(t, out, wrapped(display.ModeAfter, display.KeyPaymentInstructions, "<p>Pay now</p>"))
	assert.NotContains(t, out, `class="instructions"`)
}

func TestSnippetsOrderAtOneAnchor(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyPaymentBankData, Mode: display.ModeAfter, HTML: "after-2"},
		{At: display.KeyPaymentBankData, Mode: display.ModeBefore, HTML: "before-1"},
		{At: display.KeyPaymentBankData, Mode: display.ModeAfter, HTML: "after-3"},
	}, goblhtml.WithHide("payment.instructions.bank-data"))
	b1 := strings.Index(out, wrappedItem(display.ModeBefore, display.KeyPaymentBankData, "before-1"))
	a2 := strings.Index(out, wrappedItem(display.ModeAfter, display.KeyPaymentBankData, "after-2"))
	a3 := strings.Index(out, wrappedItem(display.ModeAfter, display.KeyPaymentBankData, "after-3"))
	card := strings.Index(out, `class="card"`)
	require.True(t, b1 > 0 && a2 > 0 && a3 > 0 && card > 0)
	assert.True(t, b1 < a2 && a2 < a3 && a3 < card)
	// Hiding the anchor leaves its snippets in place: the CSS rule only
	// targets the native element.
	assert.Contains(t, out, `.bank-data { display: none !important; }`)
}

func TestSnippetVariables(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	inv := env.Extract().(*bill.Invoice)
	inv.Series = "SNIP"
	inv.Payment.Terms = &pay.Terms{
		Key:      pay.TermKeyDueDate,
		DueDates: []*pay.DueDate{{Date: cal.NewDate(2022, 3, 15), Percent: nil}},
	}
	require.NoError(t, env.Calculate())

	const tpl = `<ul>` +
		`<li>code={{ .Code }}</li><li>series={{ .Series }}</li><li>number={{ .Number }}</li>` +
		`<li>issued={{ .IssueDate }}</li><li>due={{ .DueDate }}</li>` +
		`<li>payable={{ .TotalPayable }}</li><li>owed={{ .TotalDue }}</li><li>currency={{ .Currency }}</li>` +
		`<li>supplier={{ .SupplierName }}</li><li>customer={{ .CustomerName }}</li>` +
		`<li><a href="{{ .PaymentLink }}">pay</a></li>` +
		`</ul>`
	s := display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: tpl}}

	t.Run("default formats", func(t *testing.T) {
		out := renderSnippets(t, env, s)
		for _, want := range []string{
			"<li>code=FAKE20220001</li>",
			"<li>series=SNIP</li>",
			"<li>number=SNIP-FAKE20220001</li>",
			"<li>issued=2022-02-01</li>",
			"<li>due=2022-03-15</li>",
			"<li>payable=€45.110,38</li>",
			"<li>owed=€31.577,27</li>",
			"<li>currency=EUR</li>",
			"<li>supplier=Biz España S.L.</li>",
			"<li>customer=Customer Spain S.L.</li>",
			`<a href="https://www.paypal.com/pay?ref=PAY1234">pay</a>`,
		} {
			assert.Contains(t, out, want)
		}
	})

	t.Run("follow the render options", func(t *testing.T) {
		out := renderSnippets(t, env, s,
			goblhtml.WithCalFormatter("02/01/2006", "", nil),
			goblhtml.WithThousandsSeparator(","),
			goblhtml.WithDecimalMark("."),
		)
		assert.Contains(t, out, "<li>issued=01/02/2022</li>")
		assert.Contains(t, out, "<li>due=15/03/2022</li>")
		assert.Contains(t, out, "<li>payable=€45,110.38</li>")
	})

	t.Run("adjustment mode inverts the amounts", func(t *testing.T) {
		cn := loadSnippetsExample(t, "es-verifactu-credit-note.json")
		out := renderSnippets(t, cn, s, goblhtml.EnableAdjustmentMode)
		plain := renderSnippets(t, cn, s)
		assert.NotEqual(t, plain, out)
		assert.Contains(t, out, "<li>payable=€-")
	})

	t.Run("missing data gives empty variables", func(t *testing.T) {
		dlv := loadSnippetsExample(t, snippetsDeliveryExample)
		inv := dlv.Extract().(*bill.Invoice)
		require.Nil(t, inv.Payment)
		out := renderSnippets(t, dlv, display.Snippets{{
			At: display.KeyNotes, Mode: display.ModeAfter,
			HTML: `{{ if .PaymentLink }}<a href="{{ .PaymentLink }}">pay</a>{{ else }}<i>no link</i>{{ end }}<i>due={{ .DueDate }}</i>`,
		}})
		assert.Contains(t, out, "<i>no link</i><i>due=</i>")
	})

	t.Run("non-invoice documents fill the shared variables", func(t *testing.T) {
		ord := loadSnippetsExample(t, "pt-at-order-full.json")
		out := renderSnippets(t, ord, s)
		// Portuguese numbers join with a slash, as in the title.
		assert.Contains(t, out, "<li>number=NE SERIES-A/123</li>")
		assert.Contains(t, out, "<li>supplier=")
		assert.Contains(t, out, "<li>payable=</li>")
		assert.Contains(t, out, "<li>due=</li>")
	})
}

func TestSnippetVariablesEscaped(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	inv := env.Extract().(*bill.Invoice)
	inv.Supplier.Name = `<img src=x onerror=alert(1)> & "Co"`
	inv.Payment.Instructions.Online[0].URL = `javascript:alert(1)`
	inv.Payment.Instructions.Online = append(inv.Payment.Instructions.Online, &pay.Online{URL: "https://pay.example.com/?a=1&b=<2>"})

	out := renderSnippets(t, env, display.Snippets{{
		At: display.KeyNotes, Mode: display.ModeAfter,
		HTML: `<p data-name="{{ .SupplierName }}">{{ .SupplierName }}</p><a href="{{ .PaymentLink }}">pay</a>`,
	}})
	// Only the snippet is checked: the renderer's own output of the same
	// data is templ's concern.
	snippet := regexp.MustCompile(`<div class="snippet snippet-after" data-at="notes">.*?</div>`).FindString(out)
	require.NotEmpty(t, snippet)
	assert.NotContains(t, snippet, `<img src=x`)
	assert.Contains(t, snippet, `<p data-name="&lt;img src=x onerror=alert(1)&gt; &amp; &#34;Co&#34;">&lt;img src=x onerror=alert(1)&gt; &amp; &#34;Co&#34;</p>`)
	assert.NotContains(t, snippet, `javascript:`)
	assert.Contains(t, snippet, `href="#ZgotmplZ"`, "html/template refuses the unsafe scheme")
}

func TestSnippetsRenderErrors(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	tests := []struct {
		name     string
		snippets display.Snippets
		err      string
	}{
		{
			name:     "unknown variable",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Total }}</p>"}},
			err:      "unknown variable 'Total'",
		},
		{
			name:     "bad syntax",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Code </p>"}},
			err:      "snippets: 0: html:",
		},
		{
			name:     "snippet too large",
			snippets: display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: strings.Repeat("x", display.MaxSnippetSize+1)}},
			err:      "exceeds 32768 bytes",
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
			err: "total html exceeds 131072 bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.snippets.Validate())
			_, err := goblhtml.Render(context.Background(), env, goblhtml.WithSnippets(tt.snippets))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestSnippetsUnknownKeysIgnored(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	s := display.Snippets{
		{At: "totals", Mode: display.ModeReplace, HTML: "x"},
		{At: display.KeyNotes, Mode: "around", HTML: "x"},
		{At: display.KeyLinesSublines, Mode: display.ModeAfter, HTML: "x"},
		// Oversized or broken, an unknown anchor is still just ignored.
		{At: "future.key", Mode: display.ModeAfter, HTML: strings.Repeat("x", display.MaxSnippetSize+1)},
		{At: "future.key", Mode: display.ModeAfter, HTML: "{{ .Nope }}"},
	}
	require.Error(t, s.Validate())
	out, err := goblhtml.Render(context.Background(), env, goblhtml.WithSnippets(s))
	require.NoError(t, err)
	plain, err := goblhtml.Render(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, string(plain), string(out))
}

func TestSnippetsStripElements(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	out := renderSnippets(t, env, display.Snippets{{
		At: display.KeyNotes, Mode: display.ModeAfter,
		HTML: `<p>keep</p><script>alert(1)</script><iframe src="https://evil.example"></iframe><object data="x"></object><embed src="y"><p>also</p>`,
	}})
	assert.Contains(t, out, wrapped(display.ModeAfter, display.KeyNotes, "<p>keep</p><p>also</p>"))
	assert.NotContains(t, out, "evil.example")
	assert.NotContains(t, out, "<iframe")
	assert.NotContains(t, out, "<object")
	assert.NotContains(t, out, "<embed")
}

func TestSnippetsKeepTaxNotes(t *testing.T) {
	env := loadSnippetsExample(t, snippetsFullExample)
	inv := env.Extract().(*bill.Invoice)
	require.NotNil(t, inv.Tax)
	require.NotEmpty(t, inv.Notes)
	inv.Tax.Notes = append(inv.Tax.Notes, &tax.Note{Text: "Tax note marker"})
	out := renderSnippets(t, env, display.Snippets{
		{At: display.KeyNotes, Mode: display.ModeReplace, HTML: "<p>replaced</p>"},
	})
	assert.Contains(t, out, `<section class="notes">`)
	assert.Contains(t, out, "Tax note marker")
	assert.Contains(t, out, wrapped(display.ModeReplace, display.KeyNotes, "<p>replaced</p>"))
	assert.NotContains(t, out, inv.Notes[0].Text)
}

// TestSnippetsNeverRepeatPerLine checks that no anchor multiplies a snippet
// by the document's lines: the count is the same on a short and on a long
// invoice, and never above the two copies the running header and the
// print and screen footers account for.
func TestSnippetsNeverRepeatPerLine(t *testing.T) {
	var all display.Snippets
	for _, k := range display.Keys() {
		all = append(all, &display.Snippet{At: k, Mode: display.ModeBefore, HTML: "<b>once-" + string(k) + "</b>"})
	}
	short := renderSnippets(t, loadSnippetsExample(t, snippetsFullExample), all)

	env := loadSnippetsExample(t, snippetsFullExample)
	inv := env.Extract().(*bill.Invoice)
	for i := 0; i < 50; i++ {
		inv.Lines = append(inv.Lines, &bill.Line{
			Quantity: inv.Lines[0].Quantity,
			Item:     &org.Item{Name: "Repeated", Price: inv.Lines[0].Item.Price},
			Taxes:    inv.Lines[0].Taxes,
		})
	}
	require.NoError(t, env.Calculate())
	long := renderSnippets(t, env, all)
	require.Greater(t, len(long), len(short))
	for _, k := range display.Keys() {
		n := strings.Count(long, "once-"+string(k)+"<")
		assert.Equal(t, strings.Count(short, "once-"+string(k)+"<"), n, k)
		assert.LessOrEqual(t, n, 2, k)
		assert.GreaterOrEqual(t, n, 1, k)
	}
}

func TestSnippetsComplianceContentUnaffected(t *testing.T) {
	env := loadSnippetsExample(t, snippetsDeliveryExample)
	var all display.Snippets
	for _, k := range display.Keys() {
		all = append(all, &display.Snippet{At: k, Mode: display.ModeReplace, HTML: ""})
	}
	out := renderSnippets(t, env, all, goblhtml.WithHide(goblhtml.HideKeys()...))
	assert.Contains(t, out, `<div class="verifactu">`)
	assert.Contains(t, out, `<section class="qr">`)
	assert.Contains(t, out, `class="title"`)
	assert.Contains(t, out, `class="supplier"`)
	assert.Contains(t, out, `class="customer"`)
	assert.Contains(t, out, `class="lines"`)
	assert.Contains(t, out, `class="totals"`)
}

// TestSnippetKeysMatchHideKeys keeps the two vocabularies aligned: every
// anchor is a hide key, and the only hide key without an anchor is the
// per-line sublines.
func TestSnippetKeysMatchHideKeys(t *testing.T) {
	hide := goblhtml.HideKeys()
	for _, k := range display.Keys() {
		assert.Contains(t, hide, string(k))
	}
	anchors := make(map[string]bool)
	for _, k := range display.Keys() {
		anchors[string(k)] = true
	}
	for _, k := range hide {
		if k == string(display.KeyLinesSublines) {
			continue
		}
		assert.True(t, anchors[k], "hide key %s has no anchor", k)
	}
}
