package it

import (
	"context"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	sdi "github.com/invopop/gobl.it.sdi/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
)

// sameTypeCorrections are the SDI document types that are corrected by
// sending another document of the same type with negative amounts, instead of
// a TD04 credit note. The AdE guide gives that document the value of a "nota
// di variazione" for VAT (Guida alla compilazione delle fatture elettroniche
// e dell'esterometro v1.10, TD04 section, p. 8). TD24, TD25 and TD27 are still
// corrected with a TD04 or TD08.
var sameTypeCorrections = []cbc.Code{
	"TD16", "TD17", "TD18", "TD19",
	"TD20", "TD21", "TD22", "TD23", "TD26", "TD28",
}

// sdiTitleLabel returns the title for a document sent through SDI, made of
// the name of its document type (TipoDocumento) followed by the code, or an
// empty string when the document has no SDI document type or the code has no
// title registered.
func sdiTitleLabel(ctx context.Context, doc internal.Document) string {
	if doc == nil {
		return ""
	}
	code := doc.GetExt().Get(sdi.ExtKeyDocumentType)
	if code.IsEmpty() {
		return ""
	}
	label := i18n.T(ctx, "regimes.it.title.sdi."+code.String())
	if strings.HasPrefix(label, "!") {
		return ""
	}
	return label + " (" + code.String() + ")"
}

// isSDICorrection reports whether the invoice reduces an earlier document of
// the same SDI type, which is the case when its type is one of the
// sameTypeCorrections and its total is negative.
func isSDICorrection(doc internal.Document) bool {
	if doc == nil {
		return false
	}
	inv, ok := doc.Extract().(*bill.Invoice)
	if !ok || inv.Totals == nil {
		return false
	}
	code := doc.GetExt().Get(sdi.ExtKeyDocumentType)
	return code.In(sameTypeCorrections...) && inv.Totals.Total.IsNegative()
}
