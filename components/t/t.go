// Package t wraps around i18n and l10n configuration.
package t

import (
	"context"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	gi18n "github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/num"
)

type formatterKey string

const (
	numFormatterCTXKey formatterKey = "num"
	calFormatterCTXKey formatterKey = "cal"
)

// CalFormatter defines a simple date and datetime formatter
type CalFormatter struct {
	Date     string // Golang date format for dates, e.g. `02/01/2006`
	Location *time.Location
	DateTime string // Date-time format
}

// CalFormatterISO is the default formatter for dates and times based on
// the recommended ISO 8601 formatting.
var CalFormatterISO = CalFormatter{
	Date:     "2006-01-02",
	DateTime: "2006-01-02 15:04",
	Location: time.UTC,
}

// WithNumFormatter prepares the context with a formatter to use
// later. The formatter should be prepared according to the currency
// defined in the document.
func WithNumFormatter(ctx context.Context, f num.Formatter) context.Context {
	return context.WithValue(ctx, numFormatterCTXKey, f)
}

// WithCalFormatter prepares a simple date and datetime formatter for
// use with the localization functions.
func WithCalFormatter(ctx context.Context, f CalFormatter) context.Context {
	return context.WithValue(ctx, calFormatterCTXKey, f)
}

// numFormatter returns the formatter currently in the context.
func numFormatter(ctx context.Context) num.Formatter {
	if f, ok := ctx.Value(numFormatterCTXKey).(num.Formatter); ok {
		return f
	}
	return num.MakeFormatter(".", ",")
}

func calFormatter(ctx context.Context) CalFormatter {
	if f, ok := ctx.Value(calFormatterCTXKey).(CalFormatter); ok {
		return f
	}
	return CalFormatterISO
}

// Lang provides the current locale code from the context.
func Lang(ctx context.Context) gi18n.Lang {
	l := i18n.GetLocale(ctx)
	if l != nil {
		return gi18n.Lang(l.Code().String())
	}
	return gi18n.EN // fallback to english
}
