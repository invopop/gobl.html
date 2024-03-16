package invoice

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/tax"
)

type lineSupport struct {
	categories []*tax.Category
	discounts  bool
	charges    bool
}

// prepareLineSupport goes through the invoice lines and determines
// some of the columns that will need to be shown
func prepareLineSupport(inv *bill.Invoice) *lineSupport {
	ls := new(lineSupport)
	r := inv.TaxRegime()

	cats := make([]*tax.Category, 0)
	for _, l := range inv.Lines {
		if len(l.Discounts) > 0 {
			ls.discounts = true
		}
		if len(l.Charges) > 0 {
			ls.charges = true
		}
		for _, combo := range l.Taxes {
			cats = addCategory(cats, r.Category(combo.Category))
		}
	}
	ls.categories = cats

	return ls
}

// addCategory if not there already
func addCategory(cats []*tax.Category, cat *tax.Category) []*tax.Category {
	for _, c := range cats {
		if c.Code == cat.Code {
			return cats
		}
	}
	return append(cats, cat)
}
