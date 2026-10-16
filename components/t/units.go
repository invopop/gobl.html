package t

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl/cbc"
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

// untdidUnitExtension is the key of the UN/ECE unit code extension that GOBL
// stores raw unit codes in. The key is declared locally so that this package
// does not depend on a GOBL version that exposes it from the UNTDID catalogue.
const untdidUnitExtension cbc.Key = "untdid-unit"

// ItemUnitName provides the display name for an item's unit. The GOBL unit
// is used when present, otherwise the UN/ECE code from the item's
// "untdid-unit" extension is shown as-is.
func ItemUnitName(ctx context.Context, item *org.Item) string {
	if item == nil {
		return ""
	}
	if item.Unit != "" {
		return UnitName(ctx, item.Unit)
	}
	return item.Ext.Get(untdidUnitExtension).String()
}

// ItemHasUnit reports whether an item has a GOBL unit or a UN/ECE unit code
// extension that can be displayed.
func ItemHasUnit(item *org.Item) bool {
	return item != nil && (item.Unit != "" || item.Ext.Has(untdidUnitExtension))
}
