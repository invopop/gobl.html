package bill

import (
	"context"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/components/regimes/pt"
	"github.com/invopop/gobl.html/internal"
)

// DocumentTitle returns the localized document-type label plus series+code, e.g. "Invoice SAMPLE-001".
func DocumentTitle(ctx context.Context, doc internal.Document) string {
	label := untdidTitleLabel(ctx, doc)
	if label == "" {
		label = invoiceTitleLabel(ctx, doc)
	}
	code := doc.GetSeries().Join(doc.GetCode()).String()
	return strings.TrimSpace(label + " " + code)
}

// DocumentNumber returns the series and code joined the way the document's
// title shows them, which is regime-specific for Portugal.
func DocumentNumber(doc internal.Document) string {
	if n := pt.Number(doc, doc.GetSeries(), doc.GetCode()); n != "" {
		return n
	}
	return doc.GetSeries().Join(doc.GetCode()).String()
}

// invoiceTitleLabel resolves the generic invoice-type label, or "" if untranslated.
func invoiceTitleLabel(ctx context.Context, doc internal.Document) string {
	var key string
	switch {
	case isSimplifiedInvoice(doc):
		key = "billing.invoice.title.standard-simplified"
	case adjustmentMode(ctx, doc):
		key = "billing.invoice.title.adjustment"
	default:
		key = "billing.invoice.title." + doc.GetType().String()
	}
	if label := i18n.T(ctx, key); !strings.HasPrefix(label, "!") {
		return label
	}
	return ""
}
