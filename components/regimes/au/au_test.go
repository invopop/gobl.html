package au

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	supplierABN = "51824753556"
	customerABN = "83914571673"
)

func localized(t *testing.T) context.Context {
	t.Helper()
	locs := new(i18n.Locales)
	require.NoError(t, locs.LoadWithDefault(locales.Content, "en"))
	return locs.Get("en").WithContext(context.Background())
}

// invoice builds and calculates a minimal Australian invoice. The price is
// multiplied by the quantity of 1 so that the total equals the price.
func invoice(t *testing.T, price string, combos ...*tax.Combo) *bill.Invoice {
	t.Helper()
	inv := &bill.Invoice{
		Regime:    tax.WithRegime("AU"),
		Type:      bill.InvoiceTypeStandard,
		Series:    "INV",
		Code:      "001",
		IssueDate: cal.MakeDate(2026, 10, 1),
		Currency:  currency.AUD,
		Supplier: &org.Party{
			Name:  "Southern Cross Consulting Pty Ltd",
			TaxID: &tax.Identity{Country: "AU", Code: supplierABN},
		},
		Customer: &org.Party{
			Name:  "Harbour City Retail Pty Ltd",
			TaxID: &tax.Identity{Country: "AU", Code: customerABN},
		},
		Lines: []*bill.Line{{
			Quantity: num.MakeAmount(1, 0),
			Item:     &org.Item{Name: "Consulting", Price: num.NewAmount(0, 2)},
			Taxes:    combos,
		}},
	}
	p, err := num.AmountFromString(price)
	require.NoError(t, err)
	inv.Lines[0].Item.Price = &p
	return inv
}

func calculate(t *testing.T, inv *bill.Invoice) internal.Document {
	t.Helper()
	require.NoError(t, inv.Calculate())
	if faults := rules.Validate(inv); faults != nil {
		require.NoError(t, faults)
	}
	return internal.DocumentFor(inv)
}

func standard() *tax.Combo {
	return &tax.Combo{Category: tax.CategoryGST, Key: tax.KeyStandard, Rate: tax.RateGeneral}
}

func TestTitleKey(t *testing.T) {
	t.Run("tax invoice when the supplier has an ABN and charges GST", func(t *testing.T) {
		doc := calculate(t, invoice(t, "100.00", standard()))
		assert.Equal(t, "regimes.au.title.tax-invoice", titleKey(doc))
		assert.Equal(t, "Tax Invoice", TitleLabel(localized(t), doc))
	})
	t.Run("simplified invoices keep the tax invoice heading", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.SetTags(tax.TagSimplified)
		doc := calculate(t, inv)
		assert.Equal(t, "regimes.au.title.tax-invoice", titleKey(doc))
		assert.Equal(t, "Tax Invoice", TitleLabel(localized(t), doc))
	})
	t.Run("plain invoice without GST", func(t *testing.T) {
		doc := calculate(t, invoice(t, "100.00"))
		assert.Equal(t, "regimes.au.title.invoice", titleKey(doc))
	})
	t.Run("plain invoice without a supplier ABN", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Supplier.TaxID = nil
		assert.Equal(t, "regimes.au.title.invoice", titleKey(calculate(t, inv)))
	})
	t.Run("adjustment notes", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Type = bill.InvoiceTypeCreditNote
		doc := calculate(t, inv)
		assert.Equal(t, "regimes.au.title.adjustment-note", titleKey(doc))
		assert.Equal(t, "regimes.au.adjustment.decreasing", adjustmentKey(doc))

		inv.Type = bill.InvoiceTypeDebitNote
		assert.Equal(t, "regimes.au.adjustment.increasing", adjustmentKey(calculate(t, inv)))
	})
	t.Run("recipient created documents", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.SetTags(tax.TagSelfBilled)
		assert.Equal(t, "regimes.au.title.recipient-created-tax-invoice", titleKey(calculate(t, inv)))
		inv.Type = bill.InvoiceTypeCreditNote
		assert.Equal(t, "regimes.au.title.recipient-created-adjustment-note", titleKey(calculate(t, inv)))
	})
	t.Run("proforma", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Type = bill.InvoiceTypeProforma
		assert.Equal(t, "regimes.au.title.proforma", titleKey(calculate(t, inv)))
	})
	t.Run("other regimes keep their titles", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Regime = tax.WithRegime("SG")
		inv.Supplier.TaxID = &tax.Identity{Country: "SG", Code: "199307558M"}
		inv.Customer.TaxID = nil
		assert.Empty(t, titleKey(calculate(t, inv)))
	})
}

func TestLabels(t *testing.T) {
	ctx := localized(t)
	t.Run("GST labels only when the invoice has GST", func(t *testing.T) {
		doc := calculate(t, invoice(t, "100.00", standard()))
		assert.Equal(t, "regimes.au.totals.gst", LabelKey(doc, "totals.taxes"))
		assert.Equal(t, "regimes.au.totals.amount_payable", LabelKey(doc, "totals.payable"))
		assert.Equal(t, "regimes.au.lines.price_excl", LabelKey(doc, "lines.price"))
		assert.Empty(t, LabelKey(calculate(t, invoice(t, "100.00")), "totals.taxes"))
	})
	t.Run("inclusive prices and notes", func(t *testing.T) {
		inv := invoice(t, "110.00", standard())
		inv.Type = bill.InvoiceTypeCreditNote
		inv.Tax = &bill.Tax{PricesInclude: tax.CategoryGST}
		doc := calculate(t, inv)
		assert.Equal(t, "regimes.au.lines.price_incl", LabelKey(doc, "lines.price"))
		assert.Equal(t, "regimes.au.totals.gst_adjustment", LabelKey(doc, "totals.taxes"))
		assert.Equal(t, "regimes.au.totals.adjustment_amount", LabelKey(doc, "totals.payable"))
	})
	t.Run("key labels follow the key, not the percent", func(t *testing.T) {
		doc := calculate(t, invoice(t, "100.00", standard()))
		assert.Equal(t, "GST-free", KeyLabel(ctx, doc, tax.CategoryGST, tax.KeyZero))
		assert.Equal(t, "Input taxed", KeyLabel(ctx, doc, tax.CategoryGST, tax.KeyExempt))
		assert.Equal(t, "N/T", KeyLabel(ctx, doc, tax.CategoryGST, tax.KeyOutsideScope))
		assert.Empty(t, KeyLabel(ctx, doc, tax.CategoryGST, tax.KeyStandard))
		assert.Empty(t, KeyLabel(ctx, doc, tax.CategoryVAT, tax.KeyZero))
	})
	t.Run("legend lists the keys in use", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Lines = append(inv.Lines, &bill.Line{
			Quantity: num.MakeAmount(1, 0),
			Item:     &org.Item{Name: "Rent", Price: num.NewAmount(50000, 2)},
			Taxes:    tax.Set{{Category: tax.CategoryGST, Key: tax.KeyExempt}},
		})
		assert.Equal(t, []cbc.Key{tax.KeyExempt}, legendKeys(calculate(t, inv)))
		assert.Empty(t, legendKeys(calculate(t, invoice(t, "100.00", standard()))))
	})
}

func TestStatements(t *testing.T) {
	t.Run("total price includes GST for inclusive all-taxable invoices", func(t *testing.T) {
		inv := invoice(t, "110.00", standard())
		inv.Tax = &bill.Tax{PricesInclude: tax.CategoryGST}
		assert.Equal(t, []string{"regimes.au.statements.includes_gst"}, statements(calculate(t, inv)))
	})
	t.Run("not for exclusive prices or mixed supplies", func(t *testing.T) {
		assert.Empty(t, statements(calculate(t, invoice(t, "100.00", standard()))))

		inv := invoice(t, "110.00", standard())
		inv.Tax = &bill.Tax{PricesInclude: tax.CategoryGST}
		inv.Lines = append(inv.Lines, &bill.Line{
			Quantity: num.MakeAmount(1, 0),
			Item:     &org.Item{Name: "Food", Price: num.NewAmount(1000, 2)},
			Taxes:    tax.Set{{Category: tax.CategoryGST, Key: tax.KeyZero, Percent: num.NewPercentage(0, 2)}},
		})
		assert.Empty(t, statements(calculate(t, inv)))
	})
	t.Run("supplier liability on recipient created invoices", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.SetTags(tax.TagSelfBilled)
		assert.Equal(t, []string{"regimes.au.statements.self_billed"}, statements(calculate(t, inv)))
	})
}

func codes(ws []internal.Warning) []cbc.Key {
	out := make([]cbc.Key, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.Code)
	}
	return out
}

func TestWarnings(t *testing.T) {
	t.Run("nothing to report", func(t *testing.T) {
		assert.Empty(t, Warnings(calculate(t, invoice(t, "2000.00", standard()))))
	})
	t.Run("customer identity from 1,000 AUD", func(t *testing.T) {
		inv := invoice(t, "1000.00", standard())
		inv.Customer = &org.Party{Addresses: []*org.Address{{Locality: "Melbourne"}}}
		assert.Equal(t, []cbc.Key{WarningCustomerIdentity}, codes(Warnings(calculate(t, inv))))

		// 909.09 + 90.91 GST = 1,000.00
		inv = invoice(t, "909.09", standard())
		inv.Customer = nil
		assert.Equal(t, []cbc.Key{WarningCustomerIdentity}, codes(Warnings(calculate(t, inv))))

		inv = invoice(t, "909.08", standard())
		inv.Customer = nil
		assert.Empty(t, Warnings(calculate(t, inv)))
	})
	t.Run("a name alone identifies the customer", func(t *testing.T) {
		inv := invoice(t, "5000.00", standard())
		inv.Customer = &org.Party{Name: "Harbour City Retail Pty Ltd"}
		assert.Empty(t, Warnings(calculate(t, inv)))
	})
	t.Run("threshold compared in AUD", func(t *testing.T) {
		inv := invoice(t, "650.00", standard()) // 715.00 USD = 1,089.23 AUD
		inv.Currency = currency.USD
		inv.Customer = nil
		inv.ExchangeRates = []*currency.ExchangeRate{{From: currency.USD, To: currency.AUD, Amount: num.MakeAmount(15234, 4)}}
		assert.Equal(t, []cbc.Key{WarningCustomerIdentity}, codes(Warnings(calculate(t, inv))))
	})
	t.Run("foreign currency without a rate into AUD", func(t *testing.T) {
		inv := invoice(t, "650.00", standard())
		inv.Currency = currency.USD
		inv.Customer = nil
		assert.Equal(t, []cbc.Key{WarningExchangeRate}, codes(Warnings(calculate(t, inv))))
	})
	t.Run("recipient created documents need both ABNs", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.SetTags(tax.TagSelfBilled)
		inv.Customer.TaxID = nil
		assert.Equal(t, []cbc.Key{WarningSelfBilledABN}, codes(Warnings(calculate(t, inv))))
	})
	t.Run("adjustment notes need a reason", func(t *testing.T) {
		inv := invoice(t, "100.00", standard())
		inv.Type = bill.InvoiceTypeCreditNote
		inv.Preceding = []*org.DocumentRef{{Series: "INV", Code: "042"}}
		assert.Equal(t, []cbc.Key{WarningAdjustmentReason}, codes(Warnings(calculate(t, inv))))

		inv.Preceding[0].Reason = "Goods returned"
		assert.Empty(t, Warnings(calculate(t, inv)))

		inv.Preceding[0].Reason = ""
		inv.Notes = []*org.Note{{Key: org.NoteKeyReason, Text: "Goods returned"}}
		assert.Empty(t, Warnings(calculate(t, inv)))
	})
	t.Run("proformas and other regimes are ignored", func(t *testing.T) {
		inv := invoice(t, "5000.00", standard())
		inv.Type = bill.InvoiceTypeProforma
		inv.Customer = nil
		assert.Empty(t, Warnings(calculate(t, inv)))
	})
}
