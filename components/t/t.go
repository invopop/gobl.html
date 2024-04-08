// Package t wraps around i18n and l10n configuration.
package t

import (
	"context"

	"github.com/invopop/gobl/num"
)

type formatterKey string

const (
	formatterCTXKey formatterKey = "formatter"
)

// WithFormatter prepares the context with a formatter to use
// later. The formatter should be prepared according to the currency
// defined in the document.
func WithFormatter(ctx context.Context, f num.Formatter) context.Context {
	return context.WithValue(ctx, formatterCTXKey, f)
}

// GetFormatter returns the formatter currently in the context.
func GetFormatter(ctx context.Context) num.Formatter {
	if f, ok := ctx.Value(formatterCTXKey).(num.Formatter); ok {
		return f
	}
	return num.MakeFormatter(".", ",")
}
