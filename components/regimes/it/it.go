// Package it provides additional templates and helper methods
// for the Italian tax regime.
package it

import (
	"context"
	"slices"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/addons/it/ticket"
	"github.com/invopop/gobl/bill"
)

// hasTicketAddon reports whether the document uses the Italian AdE ticket
// (smart receipt) addon. These documents are "documenti commerciali" rather
// than "fatture", and so get a different title.
func hasTicketAddon(doc internal.Document) bool {
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok {
		return false
	}
	return slices.Contains(inv.GetAddons(), ticket.V1)
}

// titleKey returns the translation key suffix for the Italian smart-receipt
// document title, or an empty string when no Italy-specific title applies.
func titleKey(doc internal.Document) string {
	if !hasTicketAddon(doc) {
		return ""
	}
	if doc.GetType() == bill.InvoiceTypeCorrective {
		// Refunds ("resi") are issued as corrective documents.
		return "documento-commerciale-reso"
	}
	return "documento-commerciale"
}

// TitleLabel returns the localized smart-receipt document title as a plain
// string, for contexts that need text rather than a component (e.g. the HTML
// <title>). Returns "" when no Italy-specific title applies.
func TitleLabel(ctx context.Context, doc internal.Document) string {
	key := titleKey(doc)
	if key == "" {
		return ""
	}
	label := i18n.T(ctx, "regimes.it.title."+key)
	if strings.HasPrefix(label, "!") {
		return ""
	}
	return label
}
