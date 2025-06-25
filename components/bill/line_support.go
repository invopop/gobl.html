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
	prices     bool
	totals     bool
	exemptions map[*tax.CategoryDef]bool
}

// prepareLineSupport goes through the invoice lines and determines
// some of the columns that will need to be shown
func prepareLineSupport(reg *tax.RegimeDef, lines []*bill.Line, dss []*bill.Discount, chs []*bill.Charge) *lineSupport {
	ls := new(lineSupport)
	ls.exemptions = make(map[*tax.CategoryDef]bool)

	cats := make([]*tax.CategoryDef, 0)
	for _, l := range lines {
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
		if l.Item.Price != nil {
			ls.prices = true
		}
		if l.Total != nil {
			ls.totals = true
		}
		for _, combo := range l.Taxes {
			cat := reg.CategoryDef(combo.Category)
			if cat != nil {
				cats = addCategory(cats, cat)
				if taxExemptionCode(combo.Percent, combo.Ext) != "" {
					ls.exemptions[cat] = true
				}
			}
		}
	}
	for _, row := range dss {
		if row.Code != "" {
			ls.refs = true
		}
		if row.Percent != nil {
			ls.prices = true
		}
		ls.totals = true
	}
	for _, row := range chs {
		if row.Code != "" {
			ls.refs = true
		}
		if row.Percent != nil {
			ls.prices = true
		}
		ls.totals = true
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
