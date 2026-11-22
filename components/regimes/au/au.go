// Package au defines extra output for Australian invoices so that the
// rendered document meets the ATO's information requirements for tax
// invoices (s 29-70 GST Act, GSTR 2013/1) and adjustment notes (Adjustment
// Note Information Requirements Determination 2022). Australia mandates no
// layout, QR code or signature for PDF invoices: validity is entirely about
// what the document shows, so the work here is titles, labels and
// statements.
package au

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

var country = l10n.AU.Tax()

// invoiceFor returns the invoice being rendered when it is issued under the
// Australian regime, and nil for anything else.
func invoiceFor(doc internal.Document) *bill.Invoice {
	if doc == nil || doc.GetRegime().Country.Code().Tax() != country {
		return nil
	}
	inv, _ := doc.Extract().(*bill.Invoice)
	return inv
}

// isNote reports whether the invoice is an adjustment note.
func isNote(inv *bill.Invoice) bool {
	return inv.Type.In(bill.InvoiceTypeCreditNote, bill.InvoiceTypeDebitNote)
}

// hasABN reports whether the party carries an Australian Business Number.
func hasABN(party *org.Party) bool {
	return party != nil && party.TaxID != nil &&
		party.TaxID.Country == country && party.TaxID.Code != cbc.CodeEmpty
}

// hasGST reports whether any line, discount or charge is subject to GST,
// whatever the key, which is the proxy used for the supplier being
// GST-registered.
func hasGST(inv *bill.Invoice) bool {
	found := false
	eachGSTCombo(inv, func(*tax.Combo) { found = true })
	return found
}

// onlyStandardGST reports whether every GST combo is a taxable supply at the
// standard rate, meaning the whole price includes GST.
func onlyStandardGST(inv *bill.Invoice) bool {
	only := hasGST(inv)
	eachGSTCombo(inv, func(c *tax.Combo) {
		if c.Key != tax.KeyStandard {
			only = false
		}
	})
	return only
}

// eachGSTCombo calls fn for every GST tax combo on the invoice's lines,
// discounts and charges.
func eachGSTCombo(inv *bill.Invoice, fn func(*tax.Combo)) {
	visit := func(set tax.Set) {
		if c := set.Get(tax.CategoryGST); c != nil {
			fn(c)
		}
	}
	for _, l := range inv.Lines {
		if l != nil {
			visit(l.Taxes)
		}
	}
	for _, d := range inv.Discounts {
		if d != nil {
			visit(d.Taxes)
		}
	}
	for _, c := range inv.Charges {
		if c != nil {
			visit(c.Taxes)
		}
	}
}

// gstAmount provides the total GST on the invoice, or nil when there is none.
func gstAmount(inv *bill.Invoice) *num.Amount {
	if inv.Totals == nil || inv.Totals.Taxes == nil {
		return nil
	}
	cat := inv.Totals.Taxes.Category(tax.CategoryGST)
	if cat == nil {
		return nil
	}
	return &cat.Amount
}

// pricesIncludeGST reports whether the line prices already include GST.
func pricesIncludeGST(inv *bill.Invoice) bool {
	return inv.Tax != nil && inv.Tax.PricesInclude == tax.CategoryGST
}

// audRate provides the exchange rate from the invoice currency into AUD,
// or nil when the invoice is already in AUD or carries no such rate.
func audRate(inv *bill.Invoice) *currency.ExchangeRate {
	return currency.MatchExchangeRate(inv.ExchangeRates, inv.Currency, currency.AUD)
}

// titleKey selects the heading. The ATO requires a document to show that it
// is intended to be a tax invoice, so a plain "Invoice" is only used when
// the supplier has no ABN or nothing on the document attracts GST.
func titleKey(doc internal.Document) string {
	inv := invoiceFor(doc)
	if inv == nil {
		return ""
	}
	selfBilled := inv.HasTags(tax.TagSelfBilled)
	switch inv.Type {
	case bill.InvoiceTypeStandard:
		switch {
		case !hasABN(inv.Supplier), !hasGST(inv):
			return "regimes.au.title.invoice"
		case selfBilled:
			return "regimes.au.title.recipient-created-tax-invoice"
		default:
			return "regimes.au.title.tax-invoice"
		}
	case bill.InvoiceTypeCreditNote, bill.InvoiceTypeDebitNote:
		if selfBilled {
			return "regimes.au.title.recipient-created-adjustment-note"
		}
		return "regimes.au.title.adjustment-note"
	case bill.InvoiceTypeProforma:
		return "regimes.au.title.proforma"
	}
	return ""
}

// TitleLabel provides the translated ATO heading for Australian invoices,
// or "" when the default title applies.
func TitleLabel(ctx context.Context, doc internal.Document) string {
	if key := titleKey(doc); key != "" {
		return i18n.T(ctx, key)
	}
	return ""
}

// adjustmentKey names the effect of an adjustment note: a credit note is a
// decreasing adjustment and a debit note an increasing one.
func adjustmentKey(doc internal.Document) string {
	inv := invoiceFor(doc)
	if inv == nil {
		return ""
	}
	switch inv.Type {
	case bill.InvoiceTypeCreditNote:
		return "regimes.au.adjustment.decreasing"
	case bill.InvoiceTypeDebitNote:
		return "regimes.au.adjustment.increasing"
	}
	return ""
}

// LabelKey provides the translation key that replaces a generic label on
// Australian documents subject to GST, or "" to keep the generic one. Names
// identify the label by section, e.g. "totals.taxes".
func LabelKey(doc internal.Document, name string) string {
	inv := invoiceFor(doc)
	if inv == nil || !hasGST(inv) {
		return ""
	}
	switch name {
	case "lines.price":
		if pricesIncludeGST(inv) {
			return "regimes.au.lines.price_incl"
		}
		return "regimes.au.lines.price_excl"
	case "totals.taxes":
		if isNote(inv) {
			return "regimes.au.totals.gst_adjustment"
		}
		return "regimes.au.totals.gst"
	case "totals.total_with_tax":
		return "regimes.au.totals.total_incl_gst"
	case "totals.payable":
		if isNote(inv) {
			return "regimes.au.totals.adjustment_amount"
		}
		return "regimes.au.totals.amount_payable"
	}
	return ""
}

// taxKeyName maps the GOBL GST keys onto the ATO's names for supplies that
// attract no GST. Standard-rated supplies keep the percentage, and the
// mapping follows the key rather than the percentage so that a standard
// supply at 0% still reads as taxable.
var taxKeyName = map[cbc.Key]string{
	tax.KeyZero:         "gst_free",
	tax.KeyExempt:       "input_taxed",
	tax.KeyOutsideScope: "not_taxable",
}

// taxKeyOrder fixes the order of the legend entries.
var taxKeyOrder = []cbc.Key{tax.KeyZero, tax.KeyExempt, tax.KeyOutsideScope}

// KeyLabel provides the text shown in the GST column for a line or rate with
// the given key, or "" when the percentage should be shown as usual.
func KeyLabel(ctx context.Context, doc internal.Document, cat cbc.Code, key cbc.Key) string {
	if cat != tax.CategoryGST || invoiceFor(doc) == nil {
		return ""
	}
	name, ok := taxKeyName[key]
	if !ok {
		return ""
	}
	return i18n.T(ctx, "regimes.au.lines."+name)
}

// legendKeys lists the non-taxable keys used on the invoice so that the
// column values can be explained under the table.
func legendKeys(doc internal.Document) []cbc.Key {
	inv := invoiceFor(doc)
	if inv == nil {
		return nil
	}
	used := make(map[cbc.Key]bool)
	eachGSTCombo(inv, func(c *tax.Combo) {
		if _, ok := taxKeyName[c.Key]; ok {
			used[c.Key] = true
		}
	})
	keys := make([]cbc.Key, 0, len(used))
	for _, k := range taxKeyOrder {
		if used[k] {
			keys = append(keys, k)
		}
	}
	return keys
}

// gstInAUD converts the GST total into AUD using the invoice's exchange
// rate, since the GST payable must be ascertainable in AUD. Returns nil for
// AUD invoices or when no rate is available.
func gstInAUD(inv *bill.Invoice) *num.Amount {
	gst := gstAmount(inv)
	if gst == nil {
		return nil
	}
	rate := audRate(inv)
	if rate == nil {
		return nil
	}
	a := rate.Convert(*gst)
	return &a
}

// statements lists the ATO wording to print under the totals.
func statements(doc internal.Document) []string {
	inv := invoiceFor(doc)
	if inv == nil || inv.Totals == nil {
		return nil
	}
	var keys []string
	if inv.Type == bill.InvoiceTypeStandard && totalIncludesGST(inv) {
		keys = append(keys, "regimes.au.statements.includes_gst")
	}
	if gst := gstAmount(inv); gst != nil && !gst.IsZero() && inv.HasTags(tax.TagSelfBilled) {
		keys = append(keys, "regimes.au.statements.self_billed")
	}
	return keys
}

// totalIncludesGST reports whether the prices are GST-inclusive, every
// supply is taxable and the GST is one eleventh of the total, which is when
// the ATO's "Total price includes GST" statement applies.
func totalIncludesGST(inv *bill.Invoice) bool {
	if !pricesIncludeGST(inv) || !onlyStandardGST(inv) {
		return false
	}
	gst := gstAmount(inv)
	if gst == nil {
		return false
	}
	total := inv.Totals.TotalWithTax
	eleventh := total.Divide(num.MakeAmount(11, 0)).Rescale(total.Exp())
	return eleventh.Equals(*gst)
}

// adjustmentReasons provides the explanation of an adjustment note, taken from
// the preceding document references or, failing that, a note with the
// reason key.
func adjustmentReasons(inv *bill.Invoice) []string {
	var reasons []string
	for _, pre := range inv.Preceding {
		if pre != nil && pre.Reason != "" {
			reasons = append(reasons, pre.Reason)
		}
	}
	if len(reasons) > 0 {
		return reasons
	}
	for _, n := range inv.Notes {
		if n != nil && n.Key == org.NoteKeyReason && n.Text != "" {
			reasons = append(reasons, n.Text)
		}
	}
	return reasons
}
