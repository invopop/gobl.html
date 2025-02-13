package bill

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
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
		if row.Code != "" {
			ls.refs = true
		}
	}
	for _, row := range inv.Charges {
		if row.Code != "" {
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

// presentableIdentities will filter the list of identities so that
// the result only includes those that can be presented in the description
// area.
func presentableIdentities(idents []*org.Identity) []*org.Identity {
	nis := make([]*org.Identity, 0)
	for _, ident := range idents {
		if ident.Label != "" || ident.Type != cbc.CodeEmpty {
			nis = append(nis, ident)
		}
	}
	return nis
}
