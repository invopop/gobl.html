// Package pt provides the Portuguese regime specific output.
package pt

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/addons/pt/saft"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/regimes/pt"
	go_qr "github.com/piglig/go-qr"
)

// Parameters required by the AT for the QR code generator
const (
	// Minimum version
	atQRMinVer = 9

	// Error correction level
	atQRCorLvl = go_qr.Medium
)

// FooterNotes handles the special case when a document contains special
// notes that need to be added to the footer
func FooterNotes(env *gobl.Envelope) string {
	if appID := atAppID(env); appID != "" {
		if hash := atHash(env); hash != "" {
			return fmt.Sprintf("<b>%s</b> &middot; Processado por programa certificado %s/AT", hash, appID)
		}
	}
	return ""
}

func InvoiceTitleKey(inv *bill.Invoice) string {
	if inv == nil || inv.Tax == nil {
		return ""
	}
	typ := inv.Tax.Ext.Get(saft.ExtKeyInvoiceType)
	if typ == cbc.CodeEmpty {
		typ = inv.Tax.Ext.Get(saft.ExtKeyWorkType)
	}
	return titleKey(typ)
}

func DeliveryTitleKey(dlv *bill.Delivery) string {
	if dlv == nil || dlv.Tax == nil {
		return ""
	}
	typ := dlv.Tax.Ext.Get(saft.ExtKeyMovementType)
	return titleKey(typ)
}

func PaymentTitleKey(pmt *bill.Payment) string {
	if pmt == nil {
		return ""
	}
	typ := pmt.Ext.Get(saft.ExtKeyPaymentType)
	return titleKey(typ)
}

func Canceled(env *gobl.Envelope) bool {
	qr := env.Head.GetStamp(pt.StampProviderATQR)
	// TODO: Find a less hacky way to check if the document is canceled
	return qr != nil && strings.Contains(qr.Value, "*E:A*")
}

func titleKey(key cbc.Code) string {
	if key == cbc.CodeEmpty {
		return ""
	}
	return "regimes.pt.title." + string(key)
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
