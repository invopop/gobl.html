package bill

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/components/t"
	"github.com/invopop/gobl/pay"
)

type methodSupport struct {
	dates bool
}

func prepareMethodSupport(methods []*pay.Record) *methodSupport {
	ms := new(methodSupport)
	for _, m := range methods {
		if m != nil && m.Date != nil {
			ms.dates = true
		}
	}
	return ms
}

// paymentMethodName names the method of a payment record from its key, using
// the translated labels of the current scope, then the GOBL definition of the
// key, then the key itself. Records without a key, which invoice advances
// allow, get a generic label unless they carry a description of their own.
func paymentMethodName(ctx context.Context, rec *pay.Record) string {
	if rec.Key == "" {
		if rec.Description != "" {
			return ""
		}
		return i18n.T(ctx, "billing.invoice.totals.advance")
	}
	key := ".methods." + string(rec.Key)
	if i18n.Has(ctx, key) {
		return i18n.T(ctx, key)
	}
	for _, def := range pay.MeansKeyDefinitions {
		if def.Key == rec.Key {
			return def.Name.In(t.Lang(ctx))
		}
	}
	return string(rec.Key)
}
