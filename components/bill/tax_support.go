package bill

import "github.com/invopop/gobl/tax"

type taxSupport struct {
	exemptions bool
}

func prepareTaxSupport(taxes *tax.Total) *taxSupport {
	ts := new(taxSupport)
	for _, cat := range taxes.Categories {
		for _, rate := range cat.Rates {
			if taxExemptionCode(rate.Percent, rate.Ext) != "" {
				ts.exemptions = true
			}
		}
	}
	return ts
}
