// Package ar provides additional templates and helper methods
// for the Argentine tax regime.
package ar

import (
	"slices"

	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/addons/ar/arca"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
)

// chargeKeyVATRefund is the ar-arca-v4 addon's tourism refund charge
// (ChargeKeyVATRefund in gobl.ar.arca).
const chargeKeyVATRefund cbc.Key = "vat-refund"

// vatStatusLegends maps ar-arca-vat-status codes to their official legend text
// as defined by RG 1415.
var vatStatusLegends = map[string]string{
	"1":  "IVA RESPONSABLE INSCRIPTO",
	"4":  "IVA SUJETO EXENTO",
	"5":  "CONSUMIDOR FINAL",
	"6":  "RESPONSABLE MONOTRIBUTO",
	"7":  "SUJETO NO CATEGORIZADO",
	"8":  "PROVEEDOR DEL EXTERIOR",
	"9":  "CLIENTE DEL EXTERIOR",
	"10": "IVA LIBERADO – LEY Nº 19.640",
	"13": "MONOTRIBUTISTA SOCIAL",
	"15": "IVA NO ALCANZADO",
	"16": "MONOTRIBUTO TRABAJADOR INDEPENDIENTE PROMOVIDO",
}

func arcaVATLegend(doc internal.Document, party *org.Party, role string) string {
	if doc == nil || party == nil {
		return ""
	}
	ext := doc.GetExt()
	if ext.IsZero() {
		return ""
	}
	dt := ext.Get(arca.ExtKeyDocType)
	if dt.IsEmpty() {
		return ""
	}

	if role == "supplier" {
		return supplierLegend(doc, cbc.Code(dt.String()))
	}
	return customerLegend(party, cbc.Code(dt.String()))
}

func supplierLegend(_ internal.Document, docType cbc.Code) string {
	switch {
	case slices.Contains(arca.DocTypesA, docType),
		slices.Contains(arca.DocTypesB, docType),
		slices.Contains(arca.DocTypesT, docType):
		return "IVA RESPONSABLE INSCRIPTO"
	case slices.Contains(arca.DocTypesC, docType):
		return "RESPONSABLE MONOTRIBUTO"
	}
	return ""
}

func customerLegend(party *org.Party, docType cbc.Code) string {
	if slices.Contains(arca.DocTypesT, docType) {
		return "CLIENTE DEL EXTERIOR"
	}
	if party.Ext.IsZero() {
		return ""
	}
	vs := party.Ext.Get(arca.ExtKeyVATStatus)
	if vs.IsEmpty() {
		return ""
	}
	return vatStatusLegends[vs.String()]
}

func isVATRefund(c *bill.Charge) bool {
	return c != nil && c.Key == chargeKeyVATRefund
}

// VATRefund returns the tourism VAT refund charge (importe reintegro) of a Type T invoice, or nil.
func VATRefund(inv *bill.Invoice) *bill.Charge {
	if inv.Tax == nil || !slices.Contains(arca.DocTypesT, inv.Tax.GetExt(arca.ExtKeyDocType)) {
		return nil
	}
	if i := slices.IndexFunc(inv.Charges, isVATRefund); i >= 0 {
		return inv.Charges[i]
	}
	return nil
}

// Charges returns the charges to list with the lines; the VAT refund is shown in the totals instead.
func Charges(inv *bill.Invoice) []*bill.Charge {
	if VATRefund(inv) == nil {
		return inv.Charges
	}
	return slices.DeleteFunc(slices.Clone(inv.Charges), isVATRefund)
}

// Totals returns a copy of the totals with the VAT refund taken out of the charges and added back
// to the total and total with tax, so it is shown on its own row just before the payable.
func Totals(inv *bill.Invoice, totals *bill.Totals) *bill.Totals {
	refund := VATRefund(inv)
	if refund == nil || totals == nil {
		return totals
	}
	t := *totals
	t.Total = t.Total.Subtract(refund.Amount)
	t.TotalWithTax = t.TotalWithTax.Subtract(refund.Amount)
	if t.Charge != nil {
		if charge := t.Charge.Subtract(refund.Amount); charge.IsZero() {
			t.Charge = nil
		} else {
			t.Charge = &charge
		}
	}
	return &t
}
