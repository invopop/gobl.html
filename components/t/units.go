package t

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
)

// UnitName provides a localized display name for the given unit. Units
// with a conventional symbol (e.g. "kg", "m²") use the symbol directly,
// while the rest are translated using the "units" locale keys, falling
// back to the English name from the GOBL unit definitions. Units not
// defined by GOBL, such as raw UN/ECE codes, are provided as-is.
func UnitName(ctx context.Context, u cbc.Key) string {
	def := cbc.GetKeyDefinition(u, org.UnitDefinitions)
	if def == nil {
		return string(u)
	}
	if s := def.Meta[org.UnitMetaKeySymbol]; s != "" {
		return s
	}
	return i18n.T(ctx, "units."+string(u), i18n.Default(def.Name.String()))
}
