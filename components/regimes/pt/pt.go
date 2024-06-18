// Package pt provides the Portuguese regime specific output.
package pt

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/invopop/gobl"
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
