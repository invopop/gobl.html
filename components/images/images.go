// Package images provides image-related utilities.
package images

import (
	"bytes"
	"encoding/base64"

	go_qr "github.com/piglig/go-qr"
)

// QRSource provides the image source of a QR code for the provided value,
// or an empty string if the QR code cannot be generated.
func QRSource(code string) string {
	return generateQR(code)
}

// generateQR generates a QR code suitable to included
// inside the src attribute of an HTML img tag.
// Will fail silently if the QR code cannot be generated.
func generateQR(qrval string) string {
	corLvl := go_qr.Medium
	qr, err := go_qr.EncodeText(qrval, corLvl)
	if err != nil {
		return ""
	}

	conf := go_qr.NewQrCodeImgConfig(10, 4)

	// buf := bytes.NewBufferString("data:image/svg+xml;utf8,")
	buf := new(bytes.Buffer)
	if err := qr.WriteAsSVG(conf, buf, "#FFFFFF", "#000000"); err != nil {
		return ""
	}

	str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/svg+xml;base64," + str
}
