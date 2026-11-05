package goblhtml

import (
	"context"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components/t"
	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
)

// snippetVars collects the variables snippets can use from the document,
// formatted the way the document itself shows them. It runs with the
// options and locale already in the context, and after any adjustment-mode
// inversion, so the amounts match the rendered totals.
func snippetVars(ctx context.Context, env *gobl.Envelope) *display.Vars {
	doc := internal.ExtractDocumentFrom(env)
	if doc == nil {
		return nil
	}
	v := &display.Vars{
		Code:     doc.GetCode().String(),
		Series:   doc.GetSeries().String(),
		Number:   doc.GetSeries().Join(doc.GetCode()).String(),
		Currency: doc.GetCurrency().String(),
	}
	if d := doc.GetIssueDate(); !d.IsZero() {
		v.IssueDate = t.Localize(ctx, d)
	}
	if sup := doc.GetSupplier(); sup != nil {
		v.SupplierName = sup.Name
	}
	if cus := doc.GetCustomer(); cus != nil {
		v.CustomerName = cus.Name
	}
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok {
		return v
	}
	if inv.Totals != nil {
		v.TotalPayable = t.LocalizeMoney(ctx, inv.Totals.Payable)
		due := inv.Totals.Payable
		if inv.Totals.Due != nil {
			due = *inv.Totals.Due
		}
		v.TotalDue = t.LocalizeMoney(ctx, due)
	}
	if inv.Payment == nil {
		return v
	}
	if terms := inv.Payment.Terms; terms != nil {
		for _, dd := range terms.DueDates {
			if dd != nil && dd.Date != nil {
				v.DueDate = t.Localize(ctx, *dd.Date)
				break
			}
		}
	}
	if inst := inv.Payment.Instructions; inst != nil {
		for _, ol := range inst.Online {
			if ol != nil && ol.URL != "" {
				v.PaymentLink = ol.URL
				break
			}
		}
	}
	return v
}
