// Package images provides image-related utilities.
package images

import (
	"bytes"
	"encoding/base64"
	"regexp"

	go_qr "github.com/piglig/go-qr"
)

// svgDoctype matches the DOCTYPE go-qr writes before the svg element. It
// names the SVG 1.1 DTD on w3.org, which a renderer with no network or
// local file access (Prince under pdf.go) cannot resolve and reports on
// every document, and which the image does not need.
var svgDoctype = regexp.MustCompile(`<!DOCTYPE svg[^>]*>\s*`)

// QRSource generates a QR code suitable for inclusion
// inside the src attribute of an HTML img tag.
// Will fail silently, returning an empty string, if the QR code
// cannot be generated.
func QRSource(qrval string) string {
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

	svg := svgDoctype.ReplaceAll(buf.Bytes(), nil)
	str := base64.StdEncoding.EncodeToString(svg)
	return "data:image/svg+xml;base64," + str
}
