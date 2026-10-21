// Package taxid formats tax identities according to the way each country
// officially presents them on invoices.
package taxid

import (
	"context"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/tax"
)

// labelKeyPrefix is the absolute translation key under which the names of
// the tax identities are kept. Keys are absolute so that the labels resolve
// wherever a tax identity is rendered, regardless of the current scope.
const labelKeyPrefix = "org.party.labels."

// Formatted holds a tax identity broken down into the parts needed to
// present it.
type Formatted struct {
	// Label names the identifier as it is known locally, e.g. "USt-IdNr.".
	Label string
	// Code is the identifier laid out with the separators used locally, and
	// prefixed with the country code when Country is empty.
	Code string
	// Country is the tax country code to show alongside the code, and is
	// empty when the prefix already forms part of Code.
	Country string
}

// rule describes how a country's tax identity is officially presented.
// The name of the identifier is not held here but translated, see label.
type rule struct {
	// prefixed indicates that the country code forms part of the identifier
	// itself, and so is always written joined to the code, both at home and
	// abroad.
	prefixed bool
	// separator goes between the country code and the code when prefixed.
	separator string
	// format lays the code out with the separators used locally.
	format func(code string) string
}

// rules defines the country specific presentation of tax identities. GOBL
// never stores the country code inside `tax_id.code`, so a country whose
// official identifier includes the prefix is marked as `prefixed` here.
//
// The remaining countries write their national identifier without a prefix.
// They only gain one when it is used as an intra-community VAT
// identification number, which is handled by intraCommunity below.
var rules = map[l10n.TaxCountryCode]rule{
	l10n.AT.Tax(): {prefixed: true},
	l10n.CH.Tax(): {prefixed: true, format: formatCH},
	l10n.CO.Tax(): {format: formatCO},
	l10n.DE.Tax(): {prefixed: true},
	l10n.FR.Tax(): {prefixed: true},
	l10n.IE.Tax(): {prefixed: true},
	l10n.NL.Tax(): {prefixed: true},
	l10n.NO.Tax(): {prefixed: true, separator: " ", format: formatNO},
	l10n.SE.Tax(): {prefixed: true},
}

// Format prepares a tax identity for presentation, using the official
// representation of the country that issued it. The document being rendered
// is taken from the context to determine whether the identity is acting as
// an intra-community VAT identification number.
func Format(ctx context.Context, tID *tax.Identity) Formatted {
	if tID == nil {
		return Formatted{}
	}
	r := rules[tID.Country]

	f := Formatted{
		Label: label(ctx, tID.Country),
		Code:  tID.Code.String(),
	}
	if r.format != nil {
		f.Code = r.format(f.Code)
	}
	if r.prefixed || intraCommunity(ctx, tID) {
		f.Code = tID.Country.String() + r.separator + f.Code
	} else {
		f.Country = tID.Country.String()
	}
	return f
}

// label provides the name by which the identifier is known, e.g. "USt-IdNr."
// for Germany. Names are kept in the locales so that they can be adapted to
// the language the document is rendered in, and countries without one of
// their own fall back to the generic label.
func label(ctx context.Context, country l10n.TaxCountryCode) string {
	key := labelKeyPrefix + strings.ToLower(country.String())
	if i18n.Has(ctx, key) {
		return i18n.T(ctx, key)
	}
	return i18n.T(ctx, labelKeyPrefix+"default")
}

// intraCommunity reports whether the tax identity is being used as an
// intra-community VAT identification number, which happens when the document
// records a supply between two different EU member states. Article 226 of the
// VAT Directive requires the VAT identification numbers of both parties in
// that case, and those are always written with their country prefix.
func intraCommunity(ctx context.Context, tID *tax.Identity) bool {
	doc := internal.DocumentFrom(ctx)
	if doc == nil {
		return false
	}
	sup, cus := doc.GetSupplier(), doc.GetCustomer()
	if sup == nil || cus == nil {
		return false
	}
	supID, cusID := sup.TaxID, cus.TaxID
	if supID == nil || cusID == nil || supID.Country == cusID.Country {
		return false
	}
	date := doc.GetIssueDate()
	if date.IsZero() {
		date = cal.Today()
	}
	// The identity itself is checked as well so that a third party from
	// outside the union, such as a delivery receiver, keeps its national
	// format.
	return supID.InEU(date) && cusID.InEU(date) && tID.InEU(date)
}

// formatCO groups the Colombian NIT in threes and splits off the trailing
// "dígito de verificación", e.g. "9015852843" becomes "901.585.284-3".
func formatCO(code string) string {
	if len(code) < 2 {
		return code
	}
	digits, dv := code[:len(code)-1], code[len(code)-1:]
	nit := new(strings.Builder)
	for i, c := range digits {
		if i%3 == 0 && i != 0 {
			nit.WriteString(".")
		}
		nit.WriteRune(c)
	}
	return nit.String() + "-" + dv
}

// formatCH lays out the Swiss UID as "CHE-123.456.789". GOBL stores the
// identifier without the "CH" prefix, so the leading "E" and the dash that
// follows it belong to the code here.
func formatCH(code string) string {
	digits := strings.TrimPrefix(code, "E")
	if len(digits) != 9 {
		return code
	}
	return "E-" + digits[0:3] + "." + digits[3:6] + "." + digits[6:9]
}

// formatNO separates the "MVA" suffix that marks a Norwegian organisation
// number as registered for VAT, e.g. "NO 123456785 MVA".
func formatNO(code string) string {
	if base, ok := strings.CutSuffix(code, "MVA"); ok {
		return base + " MVA"
	}
	return code
}
