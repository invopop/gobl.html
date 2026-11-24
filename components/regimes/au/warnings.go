package au

import (
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

// Warning codes for information the ATO requires but the document lacks.
const (
	// WarningCustomerIdentity flags a document of 1,000 AUD or more whose
	// recipient has neither a name nor an ABN.
	WarningCustomerIdentity cbc.Key = "au-customer-identity"
	// WarningSelfBilledABN flags a recipient created document where either
	// party lacks an ABN.
	WarningSelfBilledABN cbc.Key = "au-self-billed-abn"
	// WarningExchangeRate flags a foreign currency document with no rate into
	// AUD, so the GST payable cannot be ascertained in AUD.
	WarningExchangeRate cbc.Key = "au-exchange-rate"
	// WarningAdjustmentReason flags an adjustment note that gives no reason.
	WarningAdjustmentReason cbc.Key = "au-adjustment-reason"
)

// identityThreshold is the GST-inclusive total from which the recipient's
// identity or ABN must appear on a tax invoice.
var identityThreshold = num.MakeAmount(100000, 2)

// Warnings reports the information the ATO requires on the rendered document
// that the Australian invoice does not provide. The document still renders;
// the warnings let the caller decide whether the output is usable.
func Warnings(doc internal.Document) []internal.Warning {
	inv := invoiceFor(doc)
	if inv == nil || inv.Totals == nil {
		return nil
	}
	if inv.Type != bill.InvoiceTypeStandard && !isNote(inv) {
		return nil
	}
	var ws []internal.Warning
	add := func(code cbc.Key, msg string) {
		ws = append(ws, internal.Warning{Code: code, Message: msg})
	}

	if inv.Currency != currency.AUD && audRate(inv) == nil {
		add(WarningExchangeRate, "no exchange rate into AUD, so the GST payable cannot be ascertained in AUD")
	}
	if needsCustomerIdentity(inv) && !hasIdentity(inv.Customer) {
		add(WarningCustomerIdentity, "total is 1,000 AUD or more but the customer has neither a name nor an ABN")
	}
	if inv.HasTags(tax.TagSelfBilled) && (!hasABN(inv.Supplier) || !hasABN(inv.Customer)) {
		add(WarningSelfBilledABN, "recipient created documents must show the ABN of both the supplier and the recipient")
	}
	if isNote(inv) && len(adjustmentReasons(inv)) == 0 {
		add(WarningAdjustmentReason, "adjustment notes must give a brief explanation of the reason for the adjustment")
	}
	return ws
}

// needsCustomerIdentity reports whether the GST-inclusive total reaches the
// threshold. Foreign currency totals are compared in AUD; without a rate the
// check is skipped, since the missing rate is reported on its own.
func needsCustomerIdentity(inv *bill.Invoice) bool {
	total := inv.Totals.TotalWithTax
	if inv.Currency != currency.AUD {
		rate := audRate(inv)
		if rate == nil {
			return false
		}
		total = rate.Convert(total)
	}
	return total.Compare(identityThreshold.MatchPrecision(total)) >= 0
}

// hasIdentity reports whether the party is identified by name or ABN.
func hasIdentity(party *org.Party) bool {
	return party != nil && (party.Name != "" || hasABN(party))
}
