package invoice

import (
	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components/regimes/es"
)

// HasHeaderQR returns a boolean indicating whether the envelope has some QR codes to be displayed in the header or not.
func HasHeaderQR(env *gobl.Envelope) bool {
	hasVerifactu, _, _ := es.HasVerifactuQR(env)
	hasTicketBAI, _, _ := es.HasTicketBAIQR(env)
	return hasVerifactu || hasTicketBAI
}
