package t

import (
	"context"
	"fmt"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl/catalogues/untdid"
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

// ItemUnitName provides the display name for an item's unit. When the item
// contains a UNTDID unit extension, the code is appended to the GOBL unit
// name, or used on its own if the item does not have a GOBL unit.
func ItemUnitName(ctx context.Context, item *org.Item) string {
	if item == nil {
		return ""
	}
	name := UnitName(ctx, item.Unit)
	code := item.Ext.Get(untdid.ExtKeyUnit).String()
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
	return item != nil && (item.Unit != "" || item.Ext.Has(untdid.ExtKeyUnit))
}
