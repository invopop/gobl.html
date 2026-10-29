package pt

import (
	"context"
	"unicode"
	"unicode/utf8"

	"github.com/invopop/gobl.html/components/t"
	saft "github.com/invopop/gobl.pt.saft/addon"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/gobl/tax"
)

// AdvanceMeans names the payment means of an advance in Portuguese
// invoices, from the SAF-T payment means extension the pt-saft addon sets
// on it, so the totals can say how each advance was paid. It returns an
// empty string for advances without the extension, which keeps every other
// regime untouched.
func AdvanceMeans(ctx context.Context, adv *pay.Record) string {
	if adv == nil || !adv.Ext.Has(saft.ExtKeyPaymentMeans) {
		return ""
	}
	code := adv.Ext.Get(saft.ExtKeyPaymentMeans)
	def := tax.ExtensionForKey(saft.ExtKeyPaymentMeans)
	if def == nil {
		return code.String()
	}
	val := def.CodeDef(code)
	if val == nil {
		return code.String()
	}
	return lowerFirst(val.Name.In(t.Lang(ctx)))
}

// lowerFirst lowers the first letter so the means reads as an aside after
// the advance description, like "(cash)".
func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToLower(r)) + s[size:]
}
