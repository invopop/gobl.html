package t

import (
	"context"
	"fmt"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
)

const untdidUnitExtension cbc.Key = "untdid-unit"

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

// ItemUnitName provides the display name for an item's unit. When the item
// contains a UNTDID unit extension, the code is appended to the GOBL unit
// name, or used on its own if the item does not have a GOBL unit.
func ItemUnitName(ctx context.Context, item *org.Item) string {
	if item == nil {
		return ""
	}
	name := UnitName(ctx, item.Unit)
	code := item.Ext.Get(untdidUnitExtension).String()
	if name == "" {
		return code
	}
	if code == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, code)
}

// ItemHasUnit reports whether an item has either a GOBL unit or a UNTDID unit
// extension that can be displayed.
func ItemHasUnit(item *org.Item) bool {
	return item != nil && (item.Unit != "" || item.Ext.Has(untdidUnitExtension))
}
