package goblhtml_test

import (
	"context"
	"strings"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	displayFullExample     = "es-verifactu-invoice-full.json"
	displayDeliveryExample = "es-verifactu-invoice-delivery.json"
	displayTermsExample    = "fr-ctc-invoice-b2b.json"

	ibanMarker         = "ES31 0240 4036 4573 9340 5564"
	instructionsMarker = `class="instructions"`
)

func renderDisplay(t *testing.T, env *gobl.Envelope, d *display.Options, opts ...goblhtml.Option) string {
	t.Helper()
	require.NoError(t, d.Validate())
	opts = append(opts, goblhtml.WithDisplay(d))
	out, err := goblhtml.Render(context.Background(), env, opts...)
	require.NoError(t, err)
	return string(out)
}

func loadDisplayExample(t *testing.T, name string) *gobl.Envelope {
	t.Helper()
	env, err := loadExample(name)
	require.NoError(t, err)
	return env
}

// withLineNotes adds a note to the first line of the invoice so that the
// line notes can be checked.
func withLineNotes(t *testing.T, env *gobl.Envelope) *gobl.Envelope {
	t.Helper()
	inv, ok := env.Extract().(*bill.Invoice)
	require.True(t, ok)
	inv.Lines[0].Notes = append(inv.Lines[0].Notes, &org.Note{Text: "Line note marker"})
	return env
}

func hide(keys ...display.Key) *display.Options {
	return &display.Options{Hide: keys}
}

func TestDisplayHide(t *testing.T) {
	tests := []struct {
		name    string
		example string
		key     display.Key
		opts    []goblhtml.Option
		markers []string
		kept    []string
	}{
		{
			name:    "payment",
			example: displayFullExample,
			key:     display.KeyPayment,
			markers: []string{`class="payment"`, ibanMarker},
			kept:    []string{`class="totals"`},
		},
		{
			name:    "payment instructions",
			example: displayFullExample,
			key:     display.KeyPaymentInstructions,
			markers: []string{instructionsMarker, ibanMarker},
		},
		{
			name:    "payment bank data",
			example: displayFullExample,
			key:     display.KeyPaymentBankData,
			markers: []string{`class="bank-data"`, ibanMarker},
			kept:    []string{instructionsMarker, `class="online"`},
		},
		{
			name:    "payment online",
			example: displayFullExample,
			key:     display.KeyPaymentOnline,
			markers: []string{`class="online"`},
			kept:    []string{`class="bank-data"`},
		},
		{
			name:    "payment terms",
			example: displayTermsExample,
			key:     display.KeyPaymentTerms,
			markers: []string{`class="terms"`},
			kept:    []string{`class="payment"`, instructionsMarker},
		},
		{
			name:    "taxes",
			example: displayFullExample,
			key:     display.KeyTaxes,
			markers: []string{`class="taxes"`},
			kept:    []string{`class="totals"`},
		},
		{
			name:    "header logo",
			example: displayFullExample,
			key:     display.KeyHeaderLogo,
			markers: []string{`class="alias"`, "invopop-console-files-dev"},
			kept:    []string{`class="hero"`},
		},
		{
			name:    "header title badges",
			example: displayFullExample,
			key:     display.KeyHeaderTitleBadges,
			opts:    []goblhtml.Option{goblhtml.WithLabel("BADGE-MARKER")},
			markers: []string{"BADGE-MARKER"},
		},
		{
			name:    "header summary ordering",
			example: displayFullExample,
			key:     display.KeyHeaderSummaryOrdering,
			markers: []string{`class="ordering-`},
			kept:    []string{`class="issue-date"`},
		},
		{
			name:    "header contacts delivery",
			example: displayDeliveryExample,
			key:     display.KeyHeaderContactsDelivery,
			markers: []string{`class="receiver"`, `class="delivery-date"`},
			kept:    []string{`class="supplier"`, `class="customer"`},
		},
		{
			name:    "lines item notes",
			example: displayFullExample,
			key:     display.KeyLinesItemNotes,
			markers: []string{`class="subline"`, "DEV1234567890", "0799439112766", "Line note marker"},
			kept:    []string{`class="lines"`},
		},
		{
			name:    "notes general",
			example: displayFullExample,
			key:     display.KeyNotesGeneral,
			markers: []string{`<section class="notes">`, "Thank you for your business!"},
		},
		{
			name:    "footer branding",
			example: displayFullExample,
			key:     display.KeyFooterBranding,
			markers: []string{`class="gobl-logo"`, "https://gobl.org"},
			kept:    []string{"<footer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := withLineNotes(t, loadDisplayExample(t, tt.example))
			before := renderDisplay(t, env, nil, tt.opts...)
			after := renderDisplay(t, env, hide(tt.key), tt.opts...)
			for _, m := range tt.markers {
				assert.Contains(t, before, m, "marker missing before hiding")
				assert.NotContains(t, after, m)
			}
			for _, m := range append(tt.kept, `class="lines"`, `class="totals"`) {
				assert.Contains(t, after, m)
			}
		})
	}
}

func TestDisplayHideKeepsTaxNotes(t *testing.T) {
	env := loadDisplayExample(t, displayFullExample)
	inv := env.Extract().(*bill.Invoice)
	require.NotNil(t, inv.Tax)
	require.NotEmpty(t, inv.Notes)
	inv.Tax.Notes = append(inv.Tax.Notes, &tax.Note{Text: "Tax note marker"})
	out := renderDisplay(t, env, hide(display.KeyNotesGeneral))
	assert.Contains(t, out, `<section class="notes">`)
	assert.Contains(t, out, "Tax note marker")
	assert.NotContains(t, out, inv.Notes[0].Text)
}

func TestDisplayReplace(t *testing.T) {
	env := loadDisplayExample(t, displayFullExample)
	out := renderDisplay(t, env, &display.Options{
		HTML: []*display.HTML{
			{At: display.KeyPayment, Mode: display.ModeReplace, Content: "<p>Pay via portal</p>"},
			// Descendants of a replaced section are never rendered.
			{At: display.KeyPaymentTerms, Mode: display.ModeAfter, Content: "<p>terms-after</p>"},
		},
	})
	assert.NotContains(t, out, `class="payment"`)
	assert.NotContains(t, out, ibanMarker)
	assert.NotContains(t, out, "terms-after")
	assert.Contains(t, out, `<div class="custom-html custom-replace" data-at="payment"><p>Pay via portal</p></div>`)
}

func TestDisplayReplaceWithoutNativeContent(t *testing.T) {
	env := loadDisplayExample(t, displayDeliveryExample)
	inv := env.Extract().(*bill.Invoice)
	require.Nil(t, inv.Payment)
	out := renderDisplay(t, env, &display.Options{
		HTML: []*display.HTML{
			{At: display.KeyPayment, Mode: display.ModeReplace, Content: "<p>Pay via portal</p>"},
			{At: display.KeyNotesGeneral, Mode: display.ModeReplace, Content: "<p>Custom notes</p>"},
		},
	})
	assert.Contains(t, out, `<div class="custom-html custom-replace" data-at="payment"><p>Pay via portal</p></div>`)
	assert.Contains(t, out, `<div class="custom-html custom-replace" data-at="notes.general"><p>Custom notes</p></div>`)
}

func TestDisplaySnippetsInsideEmptyPayment(t *testing.T) {
	env := loadDisplayExample(t, displayDeliveryExample)
	out := renderDisplay(t, env, &display.Options{
		HTML: []*display.HTML{
			{At: display.KeyPaymentTerms, Mode: display.ModeBefore, Content: "<p>Terms intro</p>"},
		},
	})
	assert.Contains(t, out, `<section class="payment">`)
	assert.Contains(t, out, `<div class="custom-html custom-before" data-at="payment.terms"><p>Terms intro</p></div>`)
}

func TestDisplayBeforeAfter(t *testing.T) {
	env := loadDisplayExample(t, displayFullExample)
	out := renderDisplay(t, env, &display.Options{
		Hide: []display.Key{display.KeyPaymentBankData},
		HTML: []*display.HTML{
			{At: display.KeyPaymentBankData, Mode: display.ModeAfter, Content: "<li>after-2</li>"},
			{At: display.KeyPaymentBankData, Mode: display.ModeBefore, Content: "<li>before-1</li>"},
			{At: display.KeyPaymentBankData, Mode: display.ModeAfter, Content: "<li>after-3</li>"},
			{At: display.KeyTaxes, Mode: display.ModeBefore, Content: "<p>taxes-before</p>"},
			{At: display.KeyTaxes, Mode: display.ModeAfter, Content: "<p>taxes-after</p>"},
		},
	})
	// Hidden key still renders its before/after snippets, in order.
	assert.NotContains(t, out, `class="bank-data"`)
	b1 := strings.Index(out, `<div class="custom-html custom-before" data-at="payment.instructions.bank-data"><li>before-1</li></div>`)
	a2 := strings.Index(out, `<div class="custom-html custom-after" data-at="payment.instructions.bank-data"><li>after-2</li></div>`)
	a3 := strings.Index(out, `<div class="custom-html custom-after" data-at="payment.instructions.bank-data"><li>after-3</li></div>`)
	online := strings.Index(out, `class="online"`)
	require.True(t, b1 > 0 && a2 > 0 && a3 > 0 && online > 0)
	assert.True(t, b1 < a2 && a2 < a3 && a3 < online)

	tb := strings.Index(out, "taxes-before")
	tx := strings.Index(out, `class="taxes"`)
	ta := strings.Index(out, "taxes-after")
	require.True(t, tb > 0 && tx > 0 && ta > 0)
	assert.True(t, tb < tx && tx < ta)
}

func TestDisplayAncestorHiddenDropsSnippets(t *testing.T) {
	env := loadDisplayExample(t, displayFullExample)
	out := renderDisplay(t, env, &display.Options{
		Hide: []display.Key{display.KeyPayment},
		HTML: []*display.HTML{
			{At: display.KeyPayment, Mode: display.ModeAfter, Content: "<p>payment-after</p>"},
			{At: display.KeyPaymentBankData, Mode: display.ModeReplace, Content: "<p>bank</p>"},
			{At: display.KeyPaymentTerms, Mode: display.ModeBefore, Content: "<p>terms</p>"},
		},
	})
	assert.NotContains(t, out, `class="payment"`)
	assert.Contains(t, out, `data-at="payment"><p>payment-after</p>`)
	assert.NotContains(t, out, "<p>bank</p>")
	assert.NotContains(t, out, "<p>terms</p>")
}

func TestDisplayUnknownKeysIgnored(t *testing.T) {
	env := loadDisplayExample(t, displayFullExample)
	d := &display.Options{
		// Unknown keys that are not ancestors of any known key have no effect.
		Hide: []display.Key{"totals", "pay", "payment-details"},
		HTML: []*display.HTML{{At: "title", Mode: display.ModeReplace, Content: "x"}},
	}
	require.Error(t, d.Validate())
	out, err := goblhtml.Render(context.Background(), env, goblhtml.WithDisplay(d))
	require.NoError(t, err)
	plain, err := goblhtml.Render(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, string(plain), string(out))
}

func TestDisplayRegimeQRsUnaffected(t *testing.T) {
	env := loadDisplayExample(t, displayDeliveryExample)
	all := &display.Options{Hide: display.Keys()}
	for _, k := range display.Keys() {
		all.HTML = append(all.HTML, &display.HTML{At: k, Mode: display.ModeReplace, Content: ""})
	}
	out := renderDisplay(t, env, all)
	assert.Contains(t, out, `<div class="verifactu">`)
	assert.Contains(t, out, `<section class="qr">`)
	assert.Contains(t, out, `class="title"`)
	assert.Contains(t, out, `class="supplier"`)
	assert.Contains(t, out, `class="customer"`)
	assert.Contains(t, out, `class="lines"`)
	assert.Contains(t, out, `class="totals"`)
}
