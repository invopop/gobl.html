package t

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl/org"
)

var unitDefs = func() map[org.Unit]*org.DefUnit {
	m := make(map[org.Unit]*org.DefUnit, len(org.UnitDefinitions))
	for i := range org.UnitDefinitions {
		m[org.UnitDefinitions[i].Unit] = &org.UnitDefinitions[i]
	}
	return m
}()

// UnitName provides a localized display name for the given unit. Units
// with a conventional symbol (e.g. "kg", "m²") use the symbol directly,
// while the rest are translated using the "units" locale keys, falling
// back to the English name from the GOBL unit definitions. Units not
// defined by GOBL, such as raw UN/ECE codes, are provided as-is.
func UnitName(ctx context.Context, u org.Unit) string {
	def, ok := unitDefs[u]
	if !ok {
		return string(u)
	}
	if def.Symbol != "" {
		return def.Symbol
	}
	return i18n.T(ctx, "units."+string(u), i18n.Default(def.Name))
}
