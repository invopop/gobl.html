package internal

import "github.com/invopop/gobl/cbc"

// Warning flags information that a tax authority requires on the rendered
// document but which the source document does not provide. Rendering still
// goes ahead; the warning lets the caller decide how to proceed.
type Warning struct {
	// Code identifies the missing information, e.g. "au-customer-identity".
	Code cbc.Key `json:"code"`
	// Message explains the problem in English.
	Message string `json:"message"`
}
