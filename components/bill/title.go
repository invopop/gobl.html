package bill

import (
	"context"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/components/regimes/it"
	"github.com/invopop/gobl.html/internal"
)

// DocumentTitle returns the localized document-type label plus series+code, e.g. "Invoice SAMPLE-001".
func DocumentTitle(ctx context.Context, doc internal.Document) string {
	// Italian smart receipts (AdE ticket addon) are "documenti commerciali"
	// rather than "fatture"; keep the <title> consistent with the rendered
	// document title resolved in titleType.
	label := it.TitleLabel(ctx, doc)
	if label == "" {
		label = untdidTitleLabel(ctx, doc)
	}
	if label == "" {
		label = invoiceTitleLabel(ctx, doc)
	}
	code := doc.GetSeries().Join(doc.GetCode()).String()
	return strings.TrimSpace(label + " " + code)
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
