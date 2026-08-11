// Package sg provides additional output for Singaporean invoices.
package sg

import (
	"strings"

	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
	sgregime "github.com/invopop/gobl/regimes/sg"
	"github.com/invopop/gobl/tax"
)

// country is used to check for regime-specific components.
var country = l10n.SG.Tax()

// isSingaporean reports whether the document is issued under the Singaporean
// regime.
func isSingaporean(doc internal.Document) bool {
	return doc != nil && doc.GetRegime().Country.Code().Tax() == country
}

// titleKey returns the translation key for the document title required by
// IRAS, or an empty string when no Singapore-specific title applies. IRAS
// requires the words "tax invoice" in a prominent place on tax invoices, and
// conversely a supplier that is not entitled to issue a tax invoice must not
// use that heading.
func titleKey(doc internal.Document) string {
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok || inv.Type != bill.InvoiceTypeStandard {
		// Credit notes and other document types keep the default titles.
		return ""
	}
	switch {
	case !isGSTRegistered(inv.Supplier), inv.HasTags(tax.TagSelfBilled):
		// A supplier that is not GST-registered must not issue a tax
		// invoice, and under self-billing it is the recipient who issues
		// the tax invoice, so the supplier's own document is a plain
		// invoice.
		return "regimes.sg.title.invoice"
	case inv.HasTags(tax.TagSimplified):
		return "regimes.sg.title.simplified-tax-invoice"
	case inv.HasTags(tax.TagReverseCharge):
		return "regimes.sg.title.customer-accounting-tax-invoice"
	default:
		return "regimes.sg.title.tax-invoice"
	}
}

// isGSTRegistered reports whether the party presents a GST registration
// number, which in GOBL is always stored in the tax identity.
func isGSTRegistered(party *org.Party) bool {
	return party != nil && party.TaxID != nil && party.TaxID.Code != cbc.CodeEmpty
}

// HideIdentity reports whether the identity duplicates the party's GST
// registration number and should not be rendered. For most Singaporean
// entities IRAS reuses the UEN as the GST registration number, in which case
// only a single "GST Reg No." line is shown.
func HideIdentity(party *org.Party, ident *org.Identity) bool {
	if !isSingaporeanParty(party) || ident.Type != sgregime.IdentityTypeUEN {
		return false
	}
	return normalizeCode(party.TaxID.Code) == normalizeCode(ident.Code)
}

// IdentityLabel returns the label to use for a Singaporean party identity, or
// an empty string when no Singapore-specific label applies. The UEN is
// required on invoices by s144(1A) of the Companies Act as the company
// registration number, and is only shown separately when it differs from the
// GST registration number.
func IdentityLabel(party *org.Party, ident *org.Identity) string {
	if !isSingaporeanParty(party) || ident.Type != sgregime.IdentityTypeUEN {
		return ""
	}
	return "Company Reg No."
}

func isSingaporeanParty(party *org.Party) bool {
	return party != nil && party.TaxID != nil && party.TaxID.Country.Code() == l10n.SG
}

// normalizeCode strips separators so that hyphenated and unhyphenated forms
// of the same identifier compare as equal.
func normalizeCode(code cbc.Code) string {
	return tax.IdentityCodeBadCharsRegexp.ReplaceAllString(strings.ToUpper(code.String()), "")
}
