package invoice

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/tax"
)

type lineSupport struct {
	categories []*tax.CategoryDef
	discounts  bool
	charges    bool
	refs       bool
	units      bool
}

// prepareLineSupport goes through the invoice lines and determines
// some of the columns that will need to be shown
func prepareLineSupport(inv *bill.Invoice) *lineSupport {
	ls := new(lineSupport)
	r := inv.RegimeDef()

	cats := make([]*tax.CategoryDef, 0)
	for _, l := range inv.Lines {
		if len(l.Discounts) > 0 {
			ls.discounts = true
		}
		if len(l.Charges) > 0 {
			ls.charges = true
		}
		if l.Item.Ref != "" {
			ls.refs = true
		}
		if l.Item.Unit != "" {
			ls.units = true
		}
		for _, combo := range l.Taxes {
			cat := r.CategoryDef(combo.Category)
			if cat != nil {
				cats = addCategory(cats, cat)
			}
		}
	}
	for _, row := range inv.Discounts {
		if row.Ref != "" {
			ls.refs = true
		}
	}
	for _, row := range inv.Charges {
		if row.Ref != "" {
			ls.refs = true
		}
	}

	ls.categories = cats

	return ls
}

// addCategory if not there already
func addCategory(cats []*tax.CategoryDef, cat *tax.CategoryDef) []*tax.CategoryDef {
	for _, c := range cats {
		if c.Code == cat.Code {
			return cats
		}
	}
	return append(cats, cat)
}
