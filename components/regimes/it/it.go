// Package it provides additional output for Italian invoices.
package it

import (
	"slices"

	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/addons/it/ticket"
	"github.com/invopop/gobl/bill"
)

// ticketTitleKey returns the translation key suffix for documents issued
// under the AdE "documento commerciale" (it-ticket-v1) addon, or an empty
// string when the addon is not present. All documents with this addon are
// receipts, so they must keep the official document title; corrective ones
// get the cancellation variant.
func ticketTitleKey(doc internal.Document) string {
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok {
		return ""
	}
	if !slices.Contains(inv.GetAddons(), ticket.V1) {
		return ""
	}
	if inv.Type == bill.InvoiceTypeCorrective {
		return "ticket-corrective"
	}
	return "ticket"
}
