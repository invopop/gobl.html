// Package co provides additional output for Colombian invoices.
package co

import (
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/l10n"
)

// country is used to check for regime-specific components.
var country = l10n.CO.Tax()

// isColombian reports whether the document is issued under the Colombian regime.
func isColombian(doc internal.Document) bool {
	return doc != nil && doc.GetRegime().Country.Code().Tax() == country
}

// titleKey returns the translation key suffix for the Colombian document title,
// or an empty string when no Colombia-specific title applies.
func titleKey(doc internal.Document) string {
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok {
		return ""
	}
	if inv.Type == bill.InvoiceTypeStandard {
		return "standard"
	}
	return ""
}

// FormatTaxIDCode formats a Colombian tax ID code.
func FormatTaxIDCode(taxID string) string {
	dv := taxID[len(taxID)-1:]
	nitChars := taxID[:len(taxID)-1]
	nit := ""
	for i, c := range nitChars {
		if i%3 == 0 && i != 0 {
			nit += "."
		}
		nit += string(c)
	}
	return nit + "-" + dv
}
