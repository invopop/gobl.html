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
	if ext == nil {
		return ""
	}
	dt := ext[arca.ExtKeyDocType]
	if dt.IsEmpty() {
		return ""
	}

	if role == "supplier" {
		return supplierLegend(doc, cbc.Code(dt.String()))
	}
	return customerLegend(party)
}

func supplierLegend(doc internal.Document, docType cbc.Code) string {
	switch {
	case slices.Contains(arca.DocTypesA, docType),
		slices.Contains(arca.DocTypesB, docType):
		return "IVA RESPONSABLE INSCRIPTO"
	case slices.Contains(arca.DocTypesC, docType):
		if inv, ok := doc.Extract().(*bill.Invoice); ok && inv.HasTags(arca.TagMonotax) {
			return "RESPONSABLE MONOTRIBUTO"
		}
	}
	return ""
}

func customerLegend(party *org.Party) string {
	if party.Ext == nil {
		return ""
	}
	vs := party.Ext[arca.ExtKeyVATStatus]
	if vs.IsEmpty() {
		return ""
	}
	return vatStatusLegends[vs.String()]
}
