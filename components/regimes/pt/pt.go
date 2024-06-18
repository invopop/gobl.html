package pt

import (
	"bytes"
	"encoding/base64"

	go_qr "github.com/piglig/go-qr"
)

// Parameters required by the AT for the QR code generator
const (
	// Minimum version
	atQRMinVer = 9

	// Error correction level
	atQRCorLvl = go_qr.Medium
)

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
