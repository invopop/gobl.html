// Package gr provides additional templates and helper methods
// for the Greek tax regime.
package gr

import (
	"github.com/invopop/gobl.html/internal/doc"
	"github.com/invopop/gobl/l10n"
)

// Country to check for regime-specific components
var country = l10n.EL.Tax()

func isGreek(doc doc.Document) bool {
	return doc != nil && doc.GetRegime().Country == country
}
