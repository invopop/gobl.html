// Package pt provides the Portuguese regime specific output.
package pt

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/internal/doc"
	"github.com/invopop/gobl/addons/pt/saft"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/l10n"
	gorg "github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/pt"
	go_qr "github.com/piglig/go-qr"
)

// Country to check for regime-specific components
var country = l10n.PT.Tax()

// Parameters required by the AT for the QR code generator
const (
	// Minimum version
	atQRMinVer = 9

	// Error correction level
	atQRCorLvl = go_qr.Medium
)

// Types that serve as invoices
var invoiceTypes = []cbc.Code{
	saft.InvoiceTypeStandard,
	saft.InvoiceTypeSimplified,
	saft.InvoiceTypeInvoiceReceipt,
	saft.InvoiceTypeCreditNote,
	saft.InvoiceTypeDebitNote,
	saft.WorkTypeConsignmentInv,
	saft.WorkTypeConsignmentCredit,
}

// AdaptCustomer adapts the customer to the simplified invoice format if needed.
func AdaptCustomer(doc doc.Document, par *gorg.Party) *gorg.Party {
	if !isPortuguese(doc) {
		// no need to adapt
		return par
	}

	// Make sure there's a customer even if none is provided
	var cus gorg.Party
	if par != nil {
		cus = *par
	}

	// If the customer has no tax ID, we need to adapt it to the simplified invoice format
	if cus.TaxID == nil || cus.TaxID.Code == "" {
		cus.Alias = "NIF: Consumidor Final" // Article 2.2.5 of Despacho No. 8632/2014
	}

	return &cus
}

// generateQR implements a custom QR code generator that complies with the AT spec.
func generateQR(qrval string) string {
	// The AT spec requires the QR code to be encoded in binary/bytes mode
	seg, err := go_qr.MakeBytes([]byte(qrval))
	if err != nil {
		return ""
	}

	segs := []*go_qr.QrSegment{seg}

	// Encode according to the AT specs
	qr, err := go_qr.EncodeSegments(segs, atQRCorLvl, atQRMinVer, go_qr.MaxVersion, -1, false)
	if err != nil {
		return ""
	}

	conf := go_qr.NewQrCodeImgConfig(10, 4)

	buf := new(bytes.Buffer)
	if err := qr.WriteAsSVG(conf, buf, "#FFFFFF", "#000000"); err != nil {
		return ""
	}

	str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/svg+xml;base64," + str
}

// qrNotes returns the notes to be added after the QR code
func qrNotes(env *gobl.Envelope) []string {
	if appID := atAppID(env); appID != "" {
		var notes []string
		dt := docType(doc.ExtractFrom(env))
		if dt != cbc.CodeEmpty && !slices.Contains(invoiceTypes, dt) {
			notes = append(notes, "Este documento não serve de fatura")
		}
		if hash := atHash(env); hash != "" {
			notes = append(notes,
				fmt.Sprintf("<b>%s</b>-Processado por programa certificado n.º %s/AT", hash, appID),
			)
		} else {
			notes = append(notes,
				fmt.Sprintf("Emitido por programa certificado n.º %s/AT", appID),
			)
		}
		return notes
	}
	return nil
}

func atAppID(env *gobl.Envelope) string {
	for _, stamp := range env.Head.Stamps {
		switch stamp.Provider {
		case pt.StampProviderATAppID:
			return stamp.Value
		}
	}
	return ""
}

func atHash(env *gobl.Envelope) string {
	for _, stamp := range env.Head.Stamps {
		switch stamp.Provider {
		case pt.StampProviderATHash:
			return stamp.Value
		}
	}
	return ""
}

func atATCUD(env *gobl.Envelope) string {
	for _, stamp := range env.Head.Stamps {
		switch stamp.Provider {
		case pt.StampProviderATATCUD:
			return stamp.Value
		}
	}
	return ""
}

func atQR(env *gobl.Envelope) string {
	for _, stamp := range env.Head.Stamps {
		switch stamp.Provider {
		case pt.StampProviderATQR:
			return stamp.Value
		}
	}
	return ""
}

func docType(doc doc.Document) cbc.Code {
	ext := doc.GetExt()
	switch {
	case ext.Has(saft.ExtKeyInvoiceType):
		return ext.Get(saft.ExtKeyInvoiceType)
	case ext.Has(saft.ExtKeyWorkType):
		return ext.Get(saft.ExtKeyWorkType)
	case ext.Has(saft.ExtKeyMovementType):
		return ext.Get(saft.ExtKeyMovementType)
	case ext.Has(saft.ExtKeyPaymentType):
		return ext.Get(saft.ExtKeyPaymentType)
	default:
		return cbc.CodeEmpty
	}
}

func isPortuguese(doc doc.Document) bool {
	return doc != nil && doc.GetRegime().Country == country
}
