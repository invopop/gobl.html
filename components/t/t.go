// Package t wraps around i18n and l10n configuration.
package t

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	gi18n "github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/num"
)

func numFormatter(ctx context.Context) num.Formatter {
	opts := internal.Options(ctx)
	return *opts.NumFormatter
}

func calFormatter(ctx context.Context) internal.CalFormatter {
	opts := internal.Options(ctx)
	return *opts.CalFormatter
}

// Lang provides the current locale code from the context.
func Lang(ctx context.Context) gi18n.Lang {
	l := i18n.GetLocale(ctx)
	if l != nil {
		return gi18n.Lang(l.Code().String())
	}
	return gi18n.EN // fallback to english
}
