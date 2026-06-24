package components

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components/bill"
	"github.com/invopop/gobl.html/internal"
	gbill "github.com/invopop/gobl/bill"
)

// documentTitle resolves the HTML <title> for an envelope: the document title for bills, else "Document".
func documentTitle(ctx context.Context, env *gobl.Envelope) string {
	switch doc := env.Extract().(type) {
	case *gbill.Invoice, *gbill.Payment, *gbill.Delivery, *gbill.Order:
		if title := bill.DocumentTitle(ctx, internal.DocumentFor(doc)); title != "" {
			return title
		}
	}
	return i18n.T(ctx, "billing.invoice.title.other")
}
