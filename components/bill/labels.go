package bill

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/components/regimes/au"
	"github.com/invopop/gobl.html/internal"
)

// tl translates a label, letting the regime of the document being rendered
// replace the generic key with one of its own. The name identifies the label
// by section (e.g. "totals.taxes") and key is the generic translation key
// relative to the current scope.
func tl(ctx context.Context, name, key string) string {
	doc := internal.DocumentFrom(ctx)
	if k := au.LabelKey(doc, name); k != "" {
		return i18n.T(ctx, k)
	}
	return i18n.T(ctx, key)
}
