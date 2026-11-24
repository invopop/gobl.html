package components

import (
	"github.com/invopop/gobl.html/components/regimes/au"
	"github.com/invopop/gobl.html/internal"
)

// Warnings collects the regime-specific warnings about information the
// rendered document should show but which the source document lacks.
func Warnings(doc internal.Document) []internal.Warning {
	return au.Warnings(doc)
}
